package casefile

import (
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func caseFileFromSummary(summary casefilelist.CaseFileSummary) gen.CaseFile {
	caseF := gen.CaseFile{
		Id:               httpx.ParseUUIDOrZero(summary.ID),
		ProjectId:        httpx.ParseUUIDOrZero(summary.ProjectID),
		CommitSha:        summary.CommitSHA,
		Branch:           summary.Branch,
		StartedAt:        summary.StartedAt,
		VerdictOutcome:   summary.VerdictOutcome,
		TotalFindings:    int32(summary.TotalFindings),
		NewFindings:      int32(summary.NewFindings),
		ExistingFindings: int32(summary.ExistingFindings),
		ResolvedFindings: int32(summary.ResolvedFindings),
		CreatedAt:        summary.CreatedAt,
	}
	if summary.CoveragePercent != nil {
		cp := *summary.CoveragePercent
		caseF.CoveragePercent = &cp
	}
	return caseF
}

func caseFileDetailFrom(src *casefileget.CaseFileDetail) gen.CaseFileDetail {
	evidences := make([]gen.Evidence, 0, len(src.Evidences))
	for _, e := range src.Evidences {
		evidences = append(evidences, gen.Evidence{
			Id:          httpx.ParseUUIDOrZero(e.ID),
			Subtype:     e.Subtype,
			Source:      e.Source,
			CollectedAt: e.CollectedAt,
		})
	}
	rulings := make([]gen.Ruling, 0, len(src.Rulings))
	for _, r := range src.Rulings {
		rulings = append(rulings, gen.Ruling{
			Subtype: r.Subtype, Passed: r.Passed, Detail: r.Detail,
		})
	}
	detail := gen.CaseFileDetail{
		Id:               httpx.ParseUUIDOrZero(src.ID),
		ProjectId:        httpx.ParseUUIDOrZero(src.ProjectID),
		CommitSha:        src.CommitSHA,
		Branch:           src.Branch,
		StartedAt:        src.StartedAt,
		VerdictOutcome:   src.VerdictOutcome,
		TotalFindings:    int32(src.TotalFindings),
		NewFindings:      int32(src.NewFindings),
		ExistingFindings: int32(src.ExistingFindings),
		ResolvedFindings: int32(src.ResolvedFindings),
		CreatedAt:        src.CreatedAt,
		Evidences:        evidences,
		Rulings:          rulings,
	}
	if src.CoveragePercent != nil {
		cp := *src.CoveragePercent
		detail.CoveragePercent = &cp
	}
	return detail
}

func findingFromView(view findinglist.FindingView) gen.Finding {
	return gen.Finding{
		Tool:        view.Tool,
		RuleId:      view.RuleID,
		Severity:    view.Severity,
		FilePath:    view.FilePath,
		Line:        int32(view.Line),
		Message:     view.Message,
		Fingerprint: view.FingerprintID,
		Status:      view.Status,
		Source:      view.Source,
		CommitSha:   view.CommitSHA,
		ProjectKey:  view.ProjectKey,
		CasefileId:  httpx.ParseUUIDOrZero(view.CaseFileID),
	}
}
