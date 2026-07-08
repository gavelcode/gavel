package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

func RunServer(
	ctx context.Context,
	deps Deps,
	workspace string,
	collected collectevidence.Result,
	tenantID, projectID, projectName, commitSHA, branch string,
	startedAt time.Time,
	opts Options,
) (Result, error) {
	log := deps.Log.With("project", projectName)

	localResult, err := RunLocal(ctx, deps, workspace, collected, tenantID, projectID, projectName, commitSHA, branch, startedAt, opts)
	if err != nil {
		return Result{}, err
	}

	err = submitToServer(ctx, deps, projectName, commitSHA, branch, collected, localResult, opts)
	if err != nil {
		if opts.RequireSubmit {
			return Result{}, fmt.Errorf("submit to server: %w", err)
		}
		log.Warn("server submission failed", "error", err)
		localResult.ServerFailed = true
	}

	if opts.PRNumber > 0 {
		_, plErr := deps.ServerClient.FilePleading(ctx, projectName, opts.PRNumber, opts.PRTitle, opts.PRAuthor, opts.PRBranch, branch, commitSHA)
		if plErr != nil {
			log.Warn("failed to file pleading", "error", plErr)
		}
	}

	log.Debug("results submitted to server", "project", projectName)
	return localResult, nil
}

func submitToServer(
	ctx context.Context,
	deps Deps,
	projectName, commitSHA, branch string,
	collected collectevidence.Result,
	localResult Result,
	opts Options,
) error {
	detail, err := deps.ServerClient.FetchProject(ctx, projectName)
	if err != nil {
		detail, err = tryCreateProject(ctx, deps, projectName, opts.TargetPattern)
		if err != nil {
			return fmt.Errorf("fetch project: %w", err)
		}
	}

	caseFileID, err := deps.ServerClient.OpenCaseFile(ctx, detail.ID, commitSHA, branch, false)
	if err != nil {
		return fmt.Errorf("open case file: %w", err)
	}

	files := collectEvidenceFiles(collected, projectName)
	for _, file := range files {
		dto, perr := parseEvidence(ctx, deps, file)
		if perr != nil {
			return fmt.Errorf("parse %s evidence %q: %w", file.Format, file.Source, perr)
		}
		if dto == nil {
			continue
		}
		wire := apiclient.EvidenceToWire(*dto)
		if _, err := deps.ServerClient.IngestCaseFileEvidence(ctx, caseFileID, wire); err != nil {
			return fmt.Errorf("ingest %s evidence: %w", file.Format, err)
		}
	}

	verdict := apiclient.VerdictResult{
		Outcome:     localResult.Verdict,
		EvaluatedAt: time.Now().UTC(),
	}
	for _, r := range localResult.Rulings {
		verdict.Rulings = append(verdict.Rulings, apiclient.RulingResult{
			Subtype: r.Subtype,
			Passed:  r.Passed,
			Detail:  r.Detail,
		})
	}

	counters := apiclient.CountersResult{
		FindingsCount:   localResult.FindingsCount,
		CoveragePercent: localResult.CoveragePercent,
		HasTracking:     localResult.Delta.HasPrevious,
		NewCount:        localResult.Delta.NewCount,
		ExistingCount:   localResult.Delta.ExistingCount,
		ResolvedCount:   localResult.Delta.FixedCount,
	}

	_, err = deps.ServerClient.FinalizeCaseFileWithVerdict(ctx, caseFileID, verdict, counters)
	if err != nil {
		return err
	}

	uploadSourceFiles(ctx, deps, projectName, commitSHA, localResult.Findings, opts.Workspace)
	return nil
}

func uploadSourceFiles(ctx context.Context, deps Deps, projectName, commitSHA string, findings []evidencedto.Finding, workspace string) {
	if workspace == "" || len(findings) == 0 {
		return
	}
	seen := make(map[string]struct{})
	var files []apiclient.SourceFile
	for _, finding := range findings {
		if finding.FilePath == "" {
			continue
		}
		if _, ok := seen[finding.FilePath]; ok {
			continue
		}
		seen[finding.FilePath] = struct{}{}
		absPath := filepath.Join(workspace, finding.FilePath)
		content, err := os.ReadFile(absPath)
		if err != nil {
			deps.Log.Debug("skip source upload", "file", finding.FilePath, "error", err)
			continue
		}
		files = append(files, apiclient.SourceFile{
			Path:    finding.FilePath,
			Content: string(content),
		})
	}
	if len(files) == 0 {
		return
	}
	if err := deps.ServerClient.UploadSource(ctx, projectName, commitSHA, files); err != nil {
		deps.Log.Warn("source upload failed", "error", err)
	}
}

func tryCreateProject(ctx context.Context, deps Deps, projectName, targetPattern string) (*apiclient.ProjectDetail, error) {
	if targetPattern == "" {
		targetPattern = "//..."
	}
	deps.Log.Info("project not found on server, creating", "project", projectName)
	return deps.ServerClient.CreateProject(ctx, projectName, projectName, targetPattern)
}

type evidenceFile struct {
	Format string
	Source string
	Data   []byte
}

func collectEvidenceFiles(collected collectevidence.Result, projectName string) []evidenceFile {
	var files []evidenceFile
	for _, rf := range collected.RawSARIF {
		files = append(files, evidenceFile{Format: rf.Format, Source: rf.Source, Data: rf.Data})
	}
	if collected.RawLCOV != nil {
		files = append(files, evidenceFile{Format: "lcov", Source: "coverage.lcov", Data: collected.RawLCOV})
	}
	if len(files) == 0 {
		files = append(files, emptyFindingsEvidence(projectName))
	}
	return files
}

func emptyFindingsEvidence(projectName string) evidenceFile {
	const doc = `{"$schema":"https://json.schemastore.org/sarif-2.1.0.json","version":"2.1.0","runs":[{"tool":{"driver":{"name":"gavel","informationUri":"https://gavel.dev","rules":[]}},"results":[]}]}`
	return evidenceFile{Format: "sarif", Source: projectName + ".empty.sarif", Data: []byte(doc)}
}

func parseEvidence(ctx context.Context, deps Deps, file evidenceFile) (*evidencedto.Evidence, error) {
	switch file.Format {
	case "sarif":
		cmd, err := ingestfind.NewCommand(file.Data, "sarif", file.Source, "code_quality")
		if err != nil {
			return nil, err
		}
		res, err := deps.Findings.Execute(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return &res.Evidence, nil
	case "lcov":
		cmd, err := ingestcov.NewCommand(file.Data, "lcov", file.Source)
		if err != nil {
			return nil, err
		}
		res, err := deps.Coverage.Execute(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return &res.Evidence, nil
	case "source":
		return nil, nil
	}
	return nil, fmt.Errorf("unknown evidence format %q", file.Format)
}
