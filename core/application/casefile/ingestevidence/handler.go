package ingestevidence

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/shared/event"
	casefilemodel "github.com/usegavel/gavel/core/domain/casefile/model"
	caseservice "github.com/usegavel/gavel/core/domain/casefile/service"
)

type Handler struct {
	caseFiles caseservice.CaseFileRepository
}

func NewHandler(caseFiles caseservice.CaseFileRepository) *Handler {
	if caseFiles == nil {
		panic("ingestevidence: caseFiles repository must not be nil")
	}
	return &Handler{caseFiles: caseFiles}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	id, err := casefilemodel.ParseCaseFileID(cmd.CaseFileID())
	if err != nil {
		return Result{}, fmt.Errorf("case file id: %w", err)
	}
	caseFile, err := h.caseFiles.FindByID(ctx, id)
	if err != nil {
		return Result{}, fmt.Errorf("load case file: %w", err)
	}

	now := time.Now().UTC()
	ids := make([]string, 0, len(cmd.Evidences()))
	for index, e := range cmd.Evidences() {
		evidence, err := evidencedto.EvidenceToDomain(e)
		if err != nil {
			return Result{}, fmt.Errorf("evidence[%d]: %w", index, err)
		}
		if err := caseFile.AddEvidence(evidence, now); err != nil {
			return Result{}, fmt.Errorf("evidence[%d]: %w", index, err)
		}
		ids = append(ids, evidence.ID().String())
	}

	events := caseFile.Events()
	caseFile.ClearEvents()

	if err := h.caseFiles.Save(ctx, caseFile); err != nil {
		return Result{}, fmt.Errorf("save case file: %w", err)
	}

	return Result{
		EvidenceIDs: ids,
		Events:      event.EventsFromDomain(events),
	}, nil
}
