package project

import (
	"context"
	"net/http"

	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	updatelanguages "github.com/usegavel/gavel/core/application/project/updatelanguages"
	updatequalitygate "github.com/usegavel/gavel/core/application/project/updatequalitygate"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type Deps struct {
	ListProjects             *projectlist.Handler
	CreateProject            *projectcreate.Handler
	GetProject               *projectgetbykey.Handler
	UpdateProjectLanguages   *updatelanguages.Handler
	UpdateProjectQualityGate *updatequalitygate.Handler
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) ListProjects(ctx context.Context, req gen.ListProjectsRequestObject) (gen.ListProjectsResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListProjects401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)

	q, err := projectlist.NewQuery(principal.TenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListProjects.Execute(ctx, q)
	if err != nil {
		return nil, err
	}

	items := make([]gen.Project, 0, len(res.Items))
	for _, p := range res.Items {
		items = append(items, projectFromSummary(p))
	}
	return gen.ListProjects200JSONResponse{Items: items, NextCursor: httpx.NextCursor(offset+len(items), res.Total)}, nil
}

func (h *Handler) CreateProject(ctx context.Context, req gen.CreateProjectRequestObject) (gen.CreateProjectResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.CreateProject401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.CreateProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}
	target := ""
	if req.Body.TargetPattern != nil {
		target = *req.Body.TargetPattern
	}
	cmd, err := projectcreate.NewCommand(principal.TenantID, req.Body.Key, req.Body.Name, target)
	if err != nil {
		return gen.CreateProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.CreateProject.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.Conflict:
			return gen.CreateProject409JSONResponse(httpx.NewProblem(http.StatusConflict, err.Error())), nil
		case apperr.Validation:
			return gen.CreateProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.CreateProject201JSONResponse{ProjectId: httpx.ParseUUIDOrZero(res.ProjectID)}, nil
}

func (h *Handler) GetProject(ctx context.Context, req gen.GetProjectRequestObject) (gen.GetProjectResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetProject401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetProject404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	return gen.GetProject200JSONResponse(ProjectDetailFromView(detail)), nil
}

func (h *Handler) UpdateProjectQualityGate(ctx context.Context, req gen.UpdateProjectQualityGateRequestObject) (gen.UpdateProjectQualityGateResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.UpdateProjectQualityGate401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.UpdateProjectQualityGate400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.UpdateProjectQualityGate404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	input, err := qualityGateInput(*req.Body)
	if err != nil {
		return gen.UpdateProjectQualityGate400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	cmd, err := updatequalitygate.NewCommand(principal.TenantID, detail.ID, input)
	if err != nil {
		return gen.UpdateProjectQualityGate400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.UpdateProjectQualityGate.Execute(ctx, cmd); err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.UpdateProjectQualityGate404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		case apperr.Validation:
			return gen.UpdateProjectQualityGate400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.UpdateProjectQualityGate204Response{}, nil
}

func (h *Handler) UpdateProjectLanguages(ctx context.Context, req gen.UpdateProjectLanguagesRequestObject) (gen.UpdateProjectLanguagesResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.UpdateProjectLanguages401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.UpdateProjectLanguages400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.UpdateProjectLanguages404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	cmd, err := updatelanguages.NewCommand(principal.TenantID, detail.ID, req.Body.Languages)
	if err != nil {
		return gen.UpdateProjectLanguages400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.UpdateProjectLanguages.Execute(ctx, cmd); err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.UpdateProjectLanguages404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		case apperr.Validation:
			return gen.UpdateProjectLanguages400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.UpdateProjectLanguages204Response{}, nil
}

func (h *Handler) fetchProjectDetail(ctx context.Context, tenantID, key string) (*projectgetbykey.ProjectDetail, error) {
	q, err := projectgetbykey.NewQuery(tenantID, key)
	if err != nil {
		return nil, err
	}
	return h.deps.GetProject.Execute(ctx, q)
}
