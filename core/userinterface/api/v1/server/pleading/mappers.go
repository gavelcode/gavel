package pleading

import (
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func pleadingFromSummary(summary pleadinglist.PleadingSummary) gen.Pleading {
	return gen.Pleading{
		Id:           httpx.ParseUUIDOrZero(summary.ID),
		ProjectId:    httpx.ParseUUIDOrZero(summary.ProjectID),
		Number:       int32(summary.Number),
		Title:        summary.Title,
		Petitioner:   summary.Petitioner,
		SourceBranch: summary.SourceBranch,
		TargetBranch: summary.TargetBranch,
		CommitSha:    summary.CommitSHA,
		Status:       gen.PleadingStatus(summary.Status),
		GateResult:   gateResultFromList(summary.GateResult),
		CreatedAt:    summary.CreatedAt,
		UpdatedAt:    summary.UpdatedAt,
	}
}

func pleadingFromDetail(detail *pleadingget.PleadingDetail) gen.Pleading {
	return gen.Pleading{
		Id:           httpx.ParseUUIDOrZero(detail.ID),
		ProjectId:    httpx.ParseUUIDOrZero(detail.ProjectID),
		Number:       int32(detail.Number),
		Title:        detail.Title,
		Petitioner:   detail.Petitioner,
		SourceBranch: detail.SourceBranch,
		TargetBranch: detail.TargetBranch,
		CommitSha:    detail.CommitSHA,
		Status:       gen.PleadingStatus(detail.Status),
		GateResult:   gateResultFromGet(detail.GateResult),
		CreatedAt:    detail.CreatedAt,
		UpdatedAt:    detail.UpdatedAt,
	}
}

func gateResultFromList(gateResult *pleadinglist.GateResult) *gen.GateResult {
	if gateResult == nil {
		return nil
	}
	out := gen.GateResult{Passed: gateResult.Passed}
	conds := make([]gen.GateCondition, 0, len(gateResult.Conditions))
	for _, cond := range gateResult.Conditions {
		conds = append(conds, gen.GateCondition{
			Label: cond.Label, Operator: cond.Operator, Value: cond.Value, Threshold: cond.Threshold, Passed: cond.Passed,
		})
	}
	if len(conds) > 0 {
		out.Conditions = &conds
	}
	return &out
}

func gateResultFromGet(gateResult *pleadingget.GateResult) *gen.GateResult {
	if gateResult == nil {
		return nil
	}
	out := gen.GateResult{Passed: gateResult.Passed}
	conds := make([]gen.GateCondition, 0, len(gateResult.Conditions))
	for _, cond := range gateResult.Conditions {
		conds = append(conds, gen.GateCondition{
			Label: cond.Label, Operator: cond.Operator, Value: cond.Value, Threshold: cond.Threshold, Passed: cond.Passed,
		})
	}
	if len(conds) > 0 {
		out.Conditions = &conds
	}
	return &out
}
