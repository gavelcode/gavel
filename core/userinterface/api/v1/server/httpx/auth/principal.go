package auth

import "context"

type Principal struct {
	UserID             string
	TenantID           string
	Email              string
	DisplayName        string
	Role               string
	MustChangePassword bool
	ViaAPIToken        bool
	APITokenID         string
	Scopes             []string
}

type principalCtxKey struct{}

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalCtxKey{}).(*Principal)
	return p, ok
}
