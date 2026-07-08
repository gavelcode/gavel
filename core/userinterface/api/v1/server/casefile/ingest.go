package casefile

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func (h *Handler) CreateCaseFile(ctx context.Context, req gen.CreateCaseFileRequestObject) (gen.CreateCaseFileResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.CreateCaseFile401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.CreateCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	started := h.deps.Now()
	if req.Body.StartedAt != nil {
		started = *req.Body.StartedAt
	}
	var opts []createcasefile.Option
	if req.Body.FreshEvaluation != nil && *req.Body.FreshEvaluation {
		opts = append(opts, createcasefile.WithFreshEvaluation())
	}

	cmd, err := createcasefile.NewCommand(
		principal.TenantID,
		req.Body.ProjectId.String(),
		req.Body.CommitSha,
		req.Body.Branch,
		started,
		opts...,
	)
	if err != nil {
		return gen.CreateCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}

	res, err := h.deps.CreateCaseFile.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.CreateCaseFile404JSONResponse(httpx.NewProblem(http.StatusNotFound, "project not found")), nil
		case apperr.Validation:
			return gen.CreateCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.CreateCaseFile201JSONResponse{CaseFileId: httpx.ParseUUIDOrZero(res.CaseFileID)}, nil
}

func (h *Handler) IngestCaseFileEvidence(ctx context.Context, req gen.IngestCaseFileEvidenceRequestObject) (gen.IngestCaseFileEvidenceResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.IngestCaseFileEvidence401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.IngestCaseFileEvidence400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	dto, err := wireEvidenceToDTO(*req.Body)
	if err != nil {
		return gen.IngestCaseFileEvidence400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}

	cmd, err := ingestevidence.NewCommand(principal.TenantID, req.Id.String(), []evidencedto.Evidence{dto})
	if err != nil {
		return gen.IngestCaseFileEvidence400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.IngestEvidence.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.IngestCaseFileEvidence404JSONResponse{NotFoundJSONResponse: httpx.NotFound("case file not found")}, nil
		case apperr.Validation:
			return gen.IngestCaseFileEvidence400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	if len(res.EvidenceIDs) == 0 {
		return nil, errors.New("ingestevidence returned no ids")
	}

	if h.deps.FileCoverage != nil && dto.Coverage != nil && len(dto.Coverage.ByFile) > 0 {
		if err := h.deps.FileCoverage.Save(ctx, req.Id.String(), dto.Coverage.ByFile); err != nil {
			return nil, fmt.Errorf("save file coverage: %w", err)
		}
	}

	return gen.IngestCaseFileEvidence201JSONResponse{EvidenceId: httpx.ParseUUIDOrZero(res.EvidenceIDs[0])}, nil
}

func (h *Handler) FinalizeCaseFile(ctx context.Context, req gen.FinalizeCaseFileRequestObject) (gen.FinalizeCaseFileResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.FinalizeCaseFile401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil || req.Body.Verdict.Outcome == "" {
		return gen.FinalizeCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("verdict is required")}, nil
	}

	var rulingInputs []finalize.RulingInput
	if req.Body.Verdict.Rulings != nil {
		for _, r := range *req.Body.Verdict.Rulings {
			rulingInputs = append(rulingInputs, finalize.RulingInput{
				Subtype: r.Subtype, Passed: r.Passed, Detail: r.Detail,
			})
		}
	}

	cmd, err := finalize.NewCommand(principal.TenantID, req.Id.String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     string(req.Body.Verdict.Outcome),
			Rulings:     rulingInputs,
			EvaluatedAt: req.Body.Verdict.EvaluatedAt,
		}),
	)
	if err != nil {
		return gen.FinalizeCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}

	res, err := h.deps.FinalizeCaseFile.Execute(ctx, cmd)
	if err != nil {
		return mapFinalizeError(err)
	}

	rulings := make([]gen.Ruling, 0, len(res.Verdict.Rulings))
	for _, r := range res.Verdict.Rulings {
		rulings = append(rulings, gen.Ruling{Subtype: r.Subtype, Passed: r.Passed, Detail: r.Detail})
	}
	return gen.FinalizeCaseFile200JSONResponse{
		CaseFileId: httpx.ParseUUIDOrZero(res.CaseFileID),
		Verdict: gen.Verdict{
			Outcome:     gen.VerdictOutcome(res.Verdict.Outcome),
			Rulings:     &rulings,
			EvaluatedAt: res.Verdict.EvaluatedAt,
		},
		Counters: gen.AnalysisCounters{
			FindingsCount:   int32(res.Counters.FindingsCount),
			CoveragePercent: res.Counters.CoveragePercent,
			NewCount:        int32(res.Counters.NewCount),
			ExistingCount:   int32(res.Counters.ExistingCount),
			ResolvedCount:   int32(res.Counters.ResolvedCount),
			HasTracking:     res.Counters.HasTracking,
		},
	}, nil
}

func mapFinalizeError(err error) (gen.FinalizeCaseFileResponseObject, error) {
	switch apperr.Of(err) {
	case apperr.NotFound:
		return gen.FinalizeCaseFile404JSONResponse{NotFoundJSONResponse: httpx.NotFound("case file not found")}, nil
	case apperr.Conflict:
		return gen.FinalizeCaseFile409JSONResponse(httpx.NewProblem(http.StatusConflict, "case file already judged")), nil
	case apperr.Validation:
		return gen.FinalizeCaseFile400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	default:
		return nil, err
	}
}

func wireEvidenceToDTO(input gen.IngestEvidenceRequest) (evidencedto.Evidence, error) {
	dto := evidencedto.Evidence{
		Subtype:     string(input.Subtype),
		Source:      input.Source,
		CollectedAt: input.CollectedAt,
	}
	if input.Id != nil {
		dto.ID = *input.Id
	}
	if input.Findings != nil {
		dto.Findings = make([]evidencedto.Finding, 0, len(*input.Findings))
		for _, finding := range *input.Findings {
			dto.Findings = append(dto.Findings, evidencedto.Finding{
				Tool:          finding.Tool,
				RuleID:        finding.RuleId,
				Severity:      finding.Severity,
				FilePath:      finding.FilePath,
				Line:          int(finding.Line),
				Message:       finding.Message,
				FingerprintID: finding.Fingerprint,
			})
		}
	}
	if input.Coverage != nil {
		dto.Coverage = &evidencedto.Coverage{
			TotalLines:   int(input.Coverage.TotalLines),
			CoveredLines: int(input.Coverage.CoveredLines),
		}
		if input.Coverage.ByLanguage != nil {
			dto.Coverage.ByLanguage = make([]evidencedto.LanguageStats, 0, len(*input.Coverage.ByLanguage))
			for _, l := range *input.Coverage.ByLanguage {
				dto.Coverage.ByLanguage = append(dto.Coverage.ByLanguage, evidencedto.LanguageStats{
					Language:     l.Language,
					TotalLines:   int(l.TotalLines),
					CoveredLines: int(l.CoveredLines),
				})
			}
		}
		if input.Coverage.ByFile != nil {
			dto.Coverage.ByFile = make([]evidencedto.FileCoverage, 0, len(*input.Coverage.ByFile))
			for _, fc := range *input.Coverage.ByFile {
				dto.Coverage.ByFile = append(dto.Coverage.ByFile, evidencedto.FileCoverage{
					FilePath:  fc.FilePath,
					Covered:   toIntSlice(fc.CoveredLines),
					Uncovered: toIntSlice(fc.UncoveredLines),
				})
			}
		}
	}
	if input.NewCodeCoverage != nil {
		dto.NewCodeCoverage = &evidencedto.NewCodeCoverage{
			CoveredLines:   int(input.NewCodeCoverage.CoveredLines),
			CoverableLines: int(input.NewCodeCoverage.CoverableLines),
		}
	}
	if input.License != nil {
		deps := make([]evidencedto.Dependency, 0, len(input.License.Dependencies))
		for _, d := range input.License.Dependencies {
			deps = append(deps, evidencedto.Dependency{
				Name:    d.Name,
				Version: d.Version,
				License: d.License,
			})
		}
		dto.License = &evidencedto.License{Dependencies: deps}
	}
	if input.Architecture != nil {
		violations := make([]evidencedto.Violation, 0, len(input.Architecture.Violations))
		for _, v := range input.Architecture.Violations {
			violations = append(violations, evidencedto.Violation{
				Rule:      v.Rule,
				SourcePkg: v.SourcePkg,
				TargetPkg: v.TargetPkg,
				Message:   v.Message,
			})
		}
		dto.Architecture = &evidencedto.Architecture{Violations: violations}
	}

	if dto.Findings == nil && dto.Coverage == nil && dto.NewCodeCoverage == nil && dto.License == nil && dto.Architecture == nil {
		return evidencedto.Evidence{}, fmt.Errorf("evidence content missing: populate one of findings, coverage, new_code_coverage, license, architecture")
	}
	return dto, nil
}

func toIntSlice(input []int32) []int {
	out := make([]int, len(input))
	for i, v := range input {
		out[i] = int(v)
	}
	return out
}
