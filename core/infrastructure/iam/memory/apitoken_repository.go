package iam

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

var _ service.APITokenRepository = (*APITokenRepository)(nil)

type APITokenRepository struct {
	mu       sync.RWMutex
	byID     map[string]apitoken.APIToken
	idByHash map[string]string
}

func NewAPITokenRepository() *APITokenRepository {
	return &APITokenRepository{
		byID:     make(map[string]apitoken.APIToken),
		idByHash: make(map[string]string),
	}
}

func (r *APITokenRepository) Save(_ context.Context, token apitoken.APIToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := token.ID().String()
	hash := token.TokenHash().String()
	r.byID[id] = token
	r.idByHash[hash] = id
	return nil
}

func (r *APITokenRepository) ByID(_ context.Context, id apitoken.APITokenID) (apitoken.APIToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tok, ok := r.byID[id.String()]
	if !ok {
		return apitoken.APIToken{}, fmt.Errorf("%w: %s", apitoken.ErrNotFound, id.String())
	}
	return tok, nil
}

func (r *APITokenRepository) ByTokenHash(_ context.Context, hash apitoken.SecretHash) (apitoken.APIToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idByHash[hash.String()]
	if !ok {
		return apitoken.APIToken{}, fmt.Errorf("%w: %s", apitoken.ErrNotFound, hash.String())
	}
	return r.byID[id], nil
}

func (r *APITokenRepository) ListByUser(_ context.Context, userID user.UserID) ([]apitoken.APIToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []apitoken.APIToken
	for _, tok := range r.byID {
		if tok.UserID().Equal(userID) {
			out = append(out, tok)
		}
	}
	return out, nil
}
