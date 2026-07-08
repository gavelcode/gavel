package pleading

import (
	"context"

	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	pleadingresolve "github.com/usegavel/gavel/core/application/pleading/resolve"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type Deps struct {
	ListPleadings       *pleadinglist.Handler
	GetPleading         *pleadingget.Handler
	FilePleading        *pleadingfile.Handler
	ResolvePleading     *pleadingresolve.Handler
	ResolveProjectByKey *projectgetbykey.Handler
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) ListPleadings(ctx context.Context, req gen.ListPleadingsRequestObject) (gen.ListPleadingsResponseObject, error) {
	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)
	projectID := ""
	if req.Params.ProjectId != nil {
		projectID = req.Params.ProjectId.String()
	}
	gavelspace := ""
	if req.Params.Gavelspace != nil {
		gavelspace = *req.Params.Gavelspace
	}
	status := ""
	if req.Params.Status != nil {
		status = string(*req.Params.Status)
	}
	q, err := pleadinglist.NewQuery(projectID, status, gavelspace, limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListPleadings.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.Pleading, 0, len(res.Items))
	for _, p := range res.Items {
		items = append(items, pleadingFromSummary(p))
	}
	return gen.ListPleadings200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) GetPleading(ctx context.Context, req gen.GetPleadingRequestObject) (gen.GetPleadingResponseObject, error) {
	q, err := pleadingget.NewQuery(req.Id.String())
	if err != nil {
		return nil, err
	}
	detail, err := h.deps.GetPleading.Execute(ctx, q)
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetPleading404JSONResponse{NotFoundJSONResponse: httpx.NotFound("pleading not found")}, nil
		}
		return nil, err
	}
	return gen.GetPleading200JSONResponse(pleadingFromDetail(detail)), nil
}

func (h *Handler) ResolvePleading(ctx context.Context, req gen.ResolvePleadingRequestObject) (gen.ResolvePleadingResponseObject, error) {
	if req.Body == nil {
		return gen.ResolvePleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}
	cmd, err := pleadingresolve.NewCommand(req.Id.String(), string(req.Body.Outcome))
	if err != nil {
		return gen.ResolvePleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.ResolvePleading.Execute(ctx, cmd); err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.ResolvePleading404JSONResponse{NotFoundJSONResponse: httpx.NotFound("pleading not found")}, nil
		case apperr.Validation:
			return gen.ResolvePleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.ResolvePleading204Response{}, nil
}

func (h *Handler) ListProjectPleadings(ctx context.Context, req gen.ListProjectPleadingsRequestObject) (gen.ListProjectPleadingsResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListProjectPleadings401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.ListProjectPleadings404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)
	status := ""
	if req.Params.Status != nil {
		status = string(*req.Params.Status)
	}
	q, err := pleadinglist.NewQuery(detail.ID, status, "", limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListPleadings.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.Pleading, 0, len(res.Items))
	for _, p := range res.Items {
		items = append(items, pleadingFromSummary(p))
	}
	return gen.ListProjectPleadings200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) FileProjectPleading(ctx context.Context, req gen.FileProjectPleadingRequestObject) (gen.FileProjectPleadingResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.FileProjectPleading401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.FileProjectPleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.FileProjectPleading404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	cmd, err := pleadingfile.NewCommand(
		detail.ID,
		int(req.Body.Number),
		req.Body.Title,
		req.Body.Petitioner,
		req.Body.SourceBranch,
		req.Body.TargetBranch,
		req.Body.CommitSha,
	)
	if err != nil {
		return gen.FileProjectPleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.FilePleading.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.Conflict:
			return gen.FileProjectPleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		case apperr.Validation:
			return gen.FileProjectPleading400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.FileProjectPleading201JSONResponse{PleadingId: httpx.ParseUUIDOrZero(res.PleadingID)}, nil
}

func (h *Handler) fetchProjectDetail(ctx context.Context, tenantID, key string) (*projectgetbykey.ProjectDetail, error) {
	q, err := projectgetbykey.NewQuery(tenantID, key)
	if err != nil {
		return nil, err
	}
	return h.deps.ResolveProjectByKey.Execute(ctx, q)
}
