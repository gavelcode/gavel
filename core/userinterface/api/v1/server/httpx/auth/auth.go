package auth

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

type Middleware struct {
	Resolve *resolveprincipal.Handler
	Cookie  SessionCookie
	Now     Clock
}

func NewMiddleware(resolve *resolveprincipal.Handler, cookie SessionCookie, now Clock) *Middleware {
	if resolve == nil {
		panic("middleware/auth: resolveprincipal handler must not be nil")
	}
	if now == nil {
		now = time.Now
	}
	return &Middleware{Resolve: resolve, Cookie: cookie, Now: now}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		principal, err := m.resolve(request)
		if err != nil {
			httpx.WriteProblem(writer, http.StatusUnauthorized, "unauthenticated")
			return
		}
		ctx := WithPrincipal(request.Context(), principal)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			principal, ok := PrincipalFromContext(request.Context())
			if !ok {
				httpx.WriteProblem(writer, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if !slices.Contains(roles, principal.Role) {
				httpx.WriteProblem(writer, http.StatusForbidden, "insufficient role")
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func (m *Middleware) RequireScopeOrRole(scope, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			principal, ok := PrincipalFromContext(request.Context())
			if !ok {
				httpx.WriteProblem(writer, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if principal.ViaAPIToken {
				if !slices.Contains(principal.Scopes, scope) {
					httpx.WriteProblem(writer, http.StatusForbidden, "missing scope: "+scope)
					return
				}
			} else {
				if principal.Role != role {
					httpx.WriteProblem(writer, http.StatusForbidden, "requires role: "+role)
					return
				}
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func (m *Middleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			principal, ok := PrincipalFromContext(request.Context())
			if !ok {
				httpx.WriteProblem(writer, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if principal.ViaAPIToken && !slices.Contains(principal.Scopes, scope) {
				httpx.WriteProblem(writer, http.StatusForbidden, "missing scope: "+scope)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func (m *Middleware) resolve(request *http.Request) (*Principal, error) {
	bearer := extractBearer(request.Header.Get("Authorization"))
	cookie := m.Cookie.Read(request)
	if bearer == "" && cookie == "" {
		return nil, errors.New("no credentials")
	}

	q, err := resolveprincipal.NewQuery(cookie, bearer, m.Now())
	if err != nil {
		return nil, err
	}
	res, err := m.Resolve.Execute(request.Context(), q)
	if err != nil {
		return nil, err
	}
	return &Principal{
		UserID:             res.UserID,
		TenantID:           res.TenantID,
		Email:              res.Email,
		DisplayName:        res.DisplayName,
		Role:               res.Role,
		MustChangePassword: res.MustChangePassword,
		ViaAPIToken:        res.ViaAPIToken,
		APITokenID:         res.APITokenID,
		Scopes:             res.Scopes,
	}, nil
}

func extractBearer(authz string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authz, prefix) {
		return ""
	}
	return strings.TrimSpace(authz[len(prefix):])
}
