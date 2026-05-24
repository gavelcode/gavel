package pipeline

import (
	"context"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/submit"
)

func RunLocal(
	ctx context.Context,
	deps Deps,
	workspace string,
	collected collectevidence.Result,
	projectID, projectName, commitSHA, branch string,
	startedAt time.Time,
	opts Options,
) (Result, error) {
	cmd, err := submit.NewCommand(
		projectID, commitSHA, branch,
		collected.Evidences,
		collected.Fingerprints,
		collected.ArchIDs,
		finalize.ArchDeltaInput{
			NewCount:      collected.ArchDelta.NewCount,
			FixedCount:    collected.ArchDelta.FixedCount,
			ExistingCount: collected.ArchDelta.ExistingCount,
			NewIDs:        collected.ArchDelta.NewIDs,
		},
		collected.CoverageByFile,
		opts.Quick,
		opts.Absolute,
		startedAt,
	)
	if err != nil {
		return Result{}, err
	}

	res, err := deps.Submit.Execute(ctx, cmd)
	if err != nil {
		return Result{}, err
	}

	result := mapSubmitResult(res, collected, projectName, opts)
	result.CommitSHA = commitSHA
	result.Branch = branch
	result.StartedAt = startedAt
	return result, nil
}

func mapSubmitResult(res submit.Result, collected collectevidence.Result, projectName string, opts Options) Result {
	return Result{
		Name:                   projectName,
		Verdict:                res.Verdict.Outcome,
		FindingsCount:          collected.FindingsCount,
		ViolationsCount:        collected.ViolationsCount,
		CoveragePercent:        collected.CovPercent,
		CoverageSkipped:        opts.Quick,
		NewCodeCoveragePercent: collected.NCCPercent,
		CoverageByFile:         collected.CoverageByFile,
		Rulings:                res.Verdict.Rulings,
		Findings:               collected.Findings,
		Violations:             collected.Violations,
		Delta: Delta{
			NewCount:                res.Delta.NewCount,
			FixedCount:              res.Delta.FixedCount,
			ExistingCount:           res.Delta.ExistingCount,
			NewFingerprints:         res.Delta.NewFingerprints,
			HasPrevious:             res.Delta.HasPrevious,
			NewViolationsCount:      res.Delta.NewViolationsCount,
			FixedViolationsCount:    res.Delta.FixedViolationsCount,
			ExistingViolationsCount: res.Delta.ExistingViolationsCount,
			NewViolationIDs:         res.Delta.NewViolationIDs,
			HasArchPrevious:         res.Delta.HasArchPrevious,
		},
		FirstRun:     !res.Delta.HasPrevious,
		RawSARIFDocs: collected.SARIFDocs,
		BuildWarning: collected.BuildWarning,
	}
}
