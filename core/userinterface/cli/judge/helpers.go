package judge

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/project/preparebaseline"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

func resolveGitInfo(ctx context.Context, source SourceContext, commitOverride, branchOverride string) (string, string, error) {
	commitSHA := commitOverride
	if commitSHA == "" {
		sha, err := source.CommitSHA(ctx)
		if err != nil {
			return "", "", err
		}
		commitSHA = sha
	}

	branch := branchOverride
	if branch == "" {
		b, err := source.Branch(ctx)
		if err != nil {
			return "", "", err
		}
		branch = b
	}
	return commitSHA, branch, nil
}

func validateStructure(writer io.Writer, verifier StructureVerifier, workspace string) error {
	issues, err := verifier.VerifyStructure(workspace)
	if err != nil {
		return err
	}

	if len(issues) > 0 {
		for _, issue := range issues {
			if _, err := fmt.Fprintf(writer, "  %s %s\n", ui.Dim.Render("·"), issue); err != nil {
				return err
			}
		}
		return fmt.Errorf("gavel structure invalid — run `gavel init` first")
	}
	return nil
}

func resolveConfigPath(explicit, workspace string) string {
	if explicit != "" {
		return explicit
	}
	candidate := filepath.Join(workspace, ".gavel", "gavel.yaml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	candidate = filepath.Join(workspace, "gavel.yaml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(workspace, ".gavel", "gavel.yaml")
}

func prepareBaselines(ctx context.Context, deps deps, projects []loadgavelspace.ProjectView) preparebaseline.Result {
	inputs := make([]preparebaseline.ProjectInput, 0, len(projects))
	for _, p := range projects {
		inputs = append(inputs, preparebaseline.ProjectInput{Name: p.Name, DefaultBranch: p.DefaultBranch})
	}
	cmd, err := preparebaseline.NewCommand(inputs)
	if err != nil {
		return preparebaseline.Result{}
	}

	var opts []preparebaseline.HandlerOption
	opts = append(opts, preparebaseline.WithBaselineLogger(deps.log))
	if deps.serverClient != nil {
		opts = append(opts, preparebaseline.WithFetcher(&clientBaselineFetcher{client: deps.serverClient}))
	}

	handler := preparebaseline.NewHandler(deps.projectRepo, deps.fpSeeder, opts...)
	result, err := handler.Execute(ctx, cmd)
	if err != nil {
		deps.log.Warn("baseline preparation failed", "error", err)
		return preparebaseline.Result{}
	}
	return result
}

type clientBaselineFetcher struct {
	client *apiclient.Client
}

func (f *clientBaselineFetcher) FetchBaseline(ctx context.Context, projectKey, branch string) (*preparebaseline.RemoteBaseline, error) {
	baseline, err := f.client.FetchBaseline(ctx, projectKey, branch)
	if err != nil {
		return nil, err
	}
	return &preparebaseline.RemoteBaseline{
		Fingerprints:     baseline.Fingerprints,
		ArchViolationIDs: baseline.ArchViolationIDs,
		HasPrevious:      baseline.HasPrevious,
	}, nil
}

func printBaselineStatus(writer io.Writer, result preparebaseline.Result) error {
	for _, baseline := range result.Baselines {
		if baseline.HasPrevious {
			if _, err := fmt.Fprintf(writer, "  %s %s: baseline (%s): %d fingerprints, %d arch violations\n",
				ui.Dim.Render("·"), baseline.ProjectName, baseline.Source, baseline.FingerprintCount, baseline.ArchIDCount); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(writer, "  %s %s: no previous baseline — evaluating all findings\n", ui.Dim.Render("·"), baseline.ProjectName); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasFailedVerdict(results []pipeline.Result) bool {
	for _, r := range results {
		if r.Verdict == "fail" {
			return true
		}
	}
	return false
}
