package auth

import (
	"context"
	"net/http"
)

type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (*Principal, error)
}
