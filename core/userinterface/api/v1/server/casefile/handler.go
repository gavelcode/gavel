package casefile

import (
	"context"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/application/project/getbaseline"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type FileCoverageSaver interface {
	Save(ctx context.Context, caseFileID string, entries []evidencedto.FileCoverage) error
}

type Deps struct {
	ListCaseFiles       *casefilelist.Handler
	GetCaseFile         *casefileget.Handler
	ListFindings        *findinglist.Handler
	GetBaseline         *getbaseline.Handler
	CreateCaseFile      *createcasefile.Handler
	IngestEvidence      *ingestevidence.Handler
	FinalizeCaseFile    *finalize.Handler
	ResolveProjectByKey *projectgetbykey.Handler
	FileCoverage        FileCoverageSaver
	Now                 func() time.Time
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return &Handler{deps: deps}
}

func (h *Handler) ListCaseFiles(ctx context.Context, req gen.ListCaseFilesRequestObject) (gen.ListCaseFilesResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListCaseFiles401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)
	projectID := ""
	if req.Params.ProjectId != nil {
		projectID = req.Params.ProjectId.String()
	}
	gavelspace := ""
	if req.Params.Gavelspace != nil {
		gavelspace = *req.Params.Gavelspace
	}
	if projectID == "" && gavelspace == "" {
		return gen.ListCaseFiles400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("project_id or gavelspace is required")}, nil
	}
	q, err := casefilelist.NewQuery(principal.TenantID, projectID, gavelspace, limit, offset)
	if err != nil {
		return gen.ListCaseFiles400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.ListCaseFiles.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.CaseFile, 0, len(res.Items))
	for _, c := range res.Items {
		items = append(items, caseFileFromSummary(c))
	}
	return gen.ListCaseFiles200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) GetCaseFile(ctx context.Context, req gen.GetCaseFileRequestObject) (gen.GetCaseFileResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetCaseFile401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	q, err := casefileget.NewQuery(principal.TenantID, req.Id.String())
	if err != nil {
		return nil, err
	}
	detail, err := h.deps.GetCaseFile.Execute(ctx, q)
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetCaseFile404JSONResponse{NotFoundJSONResponse: httpx.NotFound("case file not found")}, nil
		}
		return nil, err
	}
	return gen.GetCaseFile200JSONResponse(caseFileDetailFrom(detail)), nil
}

func (h *Handler) ListFindings(ctx context.Context, req gen.ListFindingsRequestObject) (gen.ListFindingsResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListFindings401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)

	filters := findinglist.Filters{
		Tool:       httpx.Deref(req.Params.Tool),
		Severity:   httpx.Deref(req.Params.Severity),
		Status:     httpx.Deref(req.Params.Status),
		FilePath:   httpx.Deref(req.Params.FilePath),
		Gavelspace: httpx.Deref(req.Params.Gavelspace),
	}
	if req.Params.ProjectId != nil {
		filters.ProjectID = req.Params.ProjectId.String()
	}
	if req.Params.CasefileId != nil {
		filters.CaseFileID = req.Params.CasefileId.String()
	}

	q, err := findinglist.NewQuery(principal.TenantID, filters, limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListFindings.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.Finding, 0, len(res.Items))
	for _, f := range res.Items {
		items = append(items, findingFromView(f))
	}
	return gen.ListFindings200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) ListProjectCaseFiles(ctx context.Context, req gen.ListProjectCaseFilesRequestObject) (gen.ListProjectCaseFilesResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListProjectCaseFiles401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.ListProjectCaseFiles404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)
	q, err := casefilelist.NewQuery(principal.TenantID, detail.ID, "", limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListCaseFiles.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.CaseFile, 0, len(res.Items))
	for _, c := range res.Items {
		items = append(items, caseFileFromSummary(c))
	}
	return gen.ListProjectCaseFiles200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) GetProjectBaseline(ctx context.Context, req gen.GetProjectBaselineRequestObject) (gen.GetProjectBaselineResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetProjectBaseline401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	branch := ""
	if req.Params.Branch != nil {
		branch = *req.Params.Branch
	}
	q, err := getbaseline.NewQuery(principal.TenantID, string(req.Key), branch)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.GetBaseline.Execute(ctx, q)
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetProjectBaseline404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	return gen.GetProjectBaseline200JSONResponse{
		Fingerprints:    res.Fingerprints,
		ArchitectureIds: res.ArchIDs,
		HasPrevious:     res.HasPrevious,
	}, nil
}

func (h *Handler) fetchProjectDetail(ctx context.Context, tenantID, key string) (*projectgetbykey.ProjectDetail, error) {
	q, err := projectgetbykey.NewQuery(tenantID, key)
	if err != nil {
		return nil, err
	}
	return h.deps.ResolveProjectByKey.Execute(ctx, q)
}
