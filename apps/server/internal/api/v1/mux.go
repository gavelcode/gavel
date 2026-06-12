package v1

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

const maxIngestBodyBytes = 10 << 20

func NewMux(server gen.StrictServerInterface, authMiddleware *auth.Middleware) http.Handler {
	wrap := &gen.ServerInterfaceWrapper{
		Handler: gen.NewStrictHandlerWithOptions(server, nil, gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		}),
		ErrorHandlerFunc: badRequestProblem,
	}

	router := chi.NewRouter()
	router.Use(auth.ExposeRawRequest)

	router.Group(func(public chi.Router) {
		public.Get("/health", wrap.GetHealth)
		public.Post("/sessions", wrap.CreateSession)
	})

	router.Group(func(authed chi.Router) {
		authed.Use(authMiddleware.Authenticate)

		authed.Delete("/sessions/current", wrap.DeleteCurrentSession)
		authed.Get("/me", wrap.GetMe)
		authed.Post("/me/password", wrap.ChangeMyPassword)
		authed.Get("/me/tokens", wrap.ListMyTokens)
		authed.Post("/me/tokens", wrap.CreateMyToken)
		authed.Delete("/me/tokens/{id}", wrap.DeleteMyToken)

		authed.Get("/gavelspaces", wrap.ListGavelspaces)
		authed.Get("/gavelspaces/{name}", wrap.GetGavelspace)

		authed.Get("/projects", wrap.ListProjects)
		authed.Get("/projects/{key}", wrap.GetProject)
		authed.Get("/projects/{key}/casefiles", wrap.ListProjectCaseFiles)
		authed.Get("/projects/{key}/pleadings", wrap.ListProjectPleadings)
		authed.Get("/projects/{key}/baseline", wrap.GetProjectBaseline)
		authed.Get("/projects/{key}/source", wrap.GetProjectSource)

		authed.Get("/pleadings", wrap.ListPleadings)
		authed.Get("/pleadings/{id}", wrap.GetPleading)
		authed.Get("/casefiles", wrap.ListCaseFiles)
		authed.Get("/casefiles/{id}", wrap.GetCaseFile)
		authed.Get("/findings", wrap.ListFindings)
		authed.Get("/search", wrap.Search)

		authed.Group(func(ingest chi.Router) {
			ingest.Use(authMiddleware.RequireScope(apitoken.ScopeIngest.String()))
			ingest.Use(httpx.MaxBody(maxIngestBodyBytes))
			ingest.Post("/casefiles", wrap.CreateCaseFile)
			ingest.Post("/casefiles/{id}/evidence", wrap.IngestCaseFileEvidence)
			ingest.Post("/casefiles/{id}/finalize", wrap.FinalizeCaseFile)
			ingest.Post("/projects/{key}/source", wrap.UploadProjectSource)
		})

		authed.Group(func(projectSync chi.Router) {
			projectSync.Use(authMiddleware.RequireScopeOrRole(apitoken.ScopeProjectSync.String(), user.RoleAdmin.String()))
			projectSync.Post("/projects", wrap.CreateProject)
		})

		authed.Group(func(admin chi.Router) {
			admin.Use(authMiddleware.RequireRole(user.RoleAdmin.String()))
			admin.Post("/admin/users", wrap.CreateUser)
			admin.Post("/gavelspaces", wrap.CreateGavelspace)
			admin.Post("/gavelspaces/{name}/projects", wrap.RegisterGavelspaceProject)
			admin.Delete("/gavelspaces/{name}/projects/{project_id}", wrap.RemoveGavelspaceProject)
			admin.Put("/projects/{key}/quality-gate", wrap.UpdateProjectQualityGate)
			admin.Put("/projects/{key}/languages", wrap.UpdateProjectLanguages)
			admin.Post("/projects/{key}/pleadings", wrap.FileProjectPleading)
			admin.Patch("/pleadings/{id}", wrap.ResolvePleading)
		})
	})

	return router
}

func badRequestProblem(writer http.ResponseWriter, _ *http.Request, err error) {
	httpx.WriteProblem(writer, http.StatusBadRequest, err.Error())
}

func requestErrorHandler(writer http.ResponseWriter, _ *http.Request, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		httpx.WriteProblem(writer, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	httpx.WriteProblem(writer, http.StatusBadRequest, err.Error())
}

func responseErrorHandler(writer http.ResponseWriter, _ *http.Request, err error) {
	httpx.WriteProblem(writer, http.StatusInternalServerError, err.Error())
}
