package gavelspace

import (
	"context"
	"net/http"

	gscreate "github.com/usegavel/gavel/core/application/gavelspace/create"
	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	gsregisterproject "github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	gsremoveproject "github.com/usegavel/gavel/core/application/gavelspace/removeproject"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type Deps struct {
	ListGavelspaces           *gslist.Handler
	CreateGavelspace          *gscreate.Handler
	GetGavelspace             *gsget.Handler
	RegisterGavelspaceProject *gsregisterproject.Handler
	RemoveGavelspaceProject   *gsremoveproject.Handler
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) ListGavelspaces(ctx context.Context, req gen.ListGavelspacesRequestObject) (gen.ListGavelspacesResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.ListGavelspaces401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	limit, offset := httpx.PageFromCursor(req.Params.Limit, req.Params.Cursor)

	q, err := gslist.NewQuery(principal.TenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.ListGavelspaces.Execute(ctx, q)
	if err != nil {
		return nil, err
	}

	items := make([]gen.Gavelspace, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, gavelspaceFromSummary(item))
	}
	return gen.ListGavelspaces200JSONResponse{
		Items:      items,
		NextCursor: httpx.NextCursor(offset+len(items), res.Total),
	}, nil
}

func (h *Handler) CreateGavelspace(ctx context.Context, req gen.CreateGavelspaceRequestObject) (gen.CreateGavelspaceResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.CreateGavelspace401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.CreateGavelspace400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	cmd, err := gscreate.NewCommand(principal.TenantID, req.Body.Name)
	if err != nil {
		return gen.CreateGavelspace400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	res, err := h.deps.CreateGavelspace.Execute(ctx, cmd)
	if err != nil {
		switch apperr.Of(err) {
		case apperr.Conflict:
			return gen.CreateGavelspace409JSONResponse(httpx.NewProblem(http.StatusConflict, err.Error())), nil
		case apperr.Validation:
			return gen.CreateGavelspace400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.CreateGavelspace201JSONResponse{Name: res.Name}, nil
}

func (h *Handler) GetGavelspace(ctx context.Context, req gen.GetGavelspaceRequestObject) (gen.GetGavelspaceResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetGavelspace401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	q, err := gsget.NewQuery(principal.TenantID, string(req.Name))
	if err != nil {
		return nil, err
	}
	detail, err := h.deps.GetGavelspace.Execute(ctx, q)
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetGavelspace404JSONResponse{NotFoundJSONResponse: httpx.NotFound("gavelspace not found")}, nil
		}
		return nil, err
	}

	projects := make([]gen.GavelspaceProject, 0, len(detail.Projects))
	for _, proj := range detail.Projects {
		projects = append(projects, projectRefFromView(proj))
	}
	return gen.GetGavelspace200JSONResponse{
		Name:      detail.Name,
		Projects:  projects,
		CreatedAt: detail.CreatedAt,
	}, nil
}

func (h *Handler) RegisterGavelspaceProject(ctx context.Context, req gen.RegisterGavelspaceProjectRequestObject) (gen.RegisterGavelspaceProjectResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.RegisterGavelspaceProject401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.RegisterGavelspaceProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("missing body")}, nil
	}

	cmd, err := gsregisterproject.NewCommand(principal.TenantID, string(req.Name), req.Body.ProjectId.String(), req.Body.TargetPattern)
	if err != nil {
		return gen.RegisterGavelspaceProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.RegisterGavelspaceProject.Execute(ctx, cmd); err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.RegisterGavelspaceProject404JSONResponse{NotFoundJSONResponse: httpx.NotFound("gavelspace not found")}, nil
		case apperr.Validation:
			return gen.RegisterGavelspaceProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		case apperr.Conflict:
			return gen.RegisterGavelspaceProject409JSONResponse(httpx.NewProblem(http.StatusConflict, err.Error())), nil
		default:
			return nil, err
		}
	}
	return gen.RegisterGavelspaceProject204Response{}, nil
}

func (h *Handler) RemoveGavelspaceProject(ctx context.Context, req gen.RemoveGavelspaceProjectRequestObject) (gen.RemoveGavelspaceProjectResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.RemoveGavelspaceProject401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}

	cmd, err := gsremoveproject.NewCommand(principal.TenantID, string(req.Name), req.ProjectId.String())
	if err != nil {
		return gen.RemoveGavelspaceProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
	}
	if _, err := h.deps.RemoveGavelspaceProject.Execute(ctx, cmd); err != nil {
		switch apperr.Of(err) {
		case apperr.NotFound:
			return gen.RemoveGavelspaceProject404JSONResponse{NotFoundJSONResponse: httpx.NotFound("gavelspace or project not found")}, nil
		case apperr.Validation:
			return gen.RemoveGavelspaceProject400JSONResponse{BadRequestJSONResponse: httpx.BadRequest(err.Error())}, nil
		default:
			return nil, err
		}
	}
	return gen.RemoveGavelspaceProject204Response{}, nil
}
