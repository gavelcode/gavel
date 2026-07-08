package submit

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Handler struct {
	createCF *createcasefile.Handler
	ingestEv *ingestevidence.Handler
	finalize *finalize.Handler
}

func NewHandler(
	createCF *createcasefile.Handler,
	ingestEv *ingestevidence.Handler,
	finalizeH *finalize.Handler,
) *Handler {
	if createCF == nil {
		panic("submit: createCF handler must not be nil")
	}
	if ingestEv == nil {
		panic("submit: ingestEv handler must not be nil")
	}
	if finalizeH == nil {
		panic("submit: finalize handler must not be nil")
	}
	return &Handler{createCF: createCF, ingestEv: ingestEv, finalize: finalizeH}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	createCmd, err := createcasefile.NewCommand(cmd.TenantID(), cmd.ProjectID(), cmd.CommitSHA(), cmd.Branch(), cmd.StartedAt())
	if err != nil {
		return Result{}, fmt.Errorf("create case file: %w", err)
	}
	createRes, err := h.createCF.Execute(ctx, createCmd)
	if err != nil {
		return Result{}, fmt.Errorf("create case file: %w", err)
	}

	ingestCmd, err := ingestevidence.NewCommand(createRes.CaseFileID, cmd.Evidences())
	if err != nil {
		return Result{}, fmt.Errorf("ingest evidence: %w", err)
	}
	if _, err := h.ingestEv.Execute(ctx, ingestCmd); err != nil {
		return Result{}, fmt.Errorf("ingest evidence: %w", err)
	}

	finalizeCmd, err := finalize.NewCommand(createRes.CaseFileID,
		finalize.WithFingerprints(cmd.Fingerprints()),
		finalize.WithArchIDs(cmd.ArchIDs()),
		finalize.WithArchDelta(cmd.ArchDelta()),
		finalize.WithFileCoverage(toFileCoverageEntries(cmd.FileCoverage())),
		finalize.WithQuick(cmd.Quick()),
		finalize.WithAbsolute(cmd.Absolute()),
	)
	if err != nil {
		return Result{}, fmt.Errorf("finalize: %w", err)
	}
	finalizeRes, err := h.finalize.Execute(ctx, finalizeCmd)
	if err != nil {
		return Result{}, fmt.Errorf("finalize: %w", err)
	}

	return Result{
		CaseFileID: finalizeRes.CaseFileID,
		Verdict:    finalizeRes.Verdict,
		Counters:   finalizeRes.Counters,
		Delta:      finalizeRes.Delta,
		Events:     finalizeRes.Events,
	}, nil
}

func toFileCoverageEntries(dtos []evidencedto.FileCoverage) []projectmodel.FileCoverageEntry {
	if len(dtos) == 0 {
		return nil
	}
	entries := make([]projectmodel.FileCoverageEntry, 0, len(dtos))
	for _, fc := range dtos {
		entry, err := projectmodel.NewFileCoverageEntry(fc.FilePath, fc.Covered, fc.Uncovered)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}
