package iam

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

var _ service.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	mu         sync.RWMutex
	byHash     map[string]session.Session
	hashByUser map[string][]string
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		byHash:     make(map[string]session.Session),
		hashByUser: make(map[string][]string),
	}
}

func (r *SessionRepository) Save(_ context.Context, sess session.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	hash := sess.TokenHash().String()
	userID := sess.UserID().String()
	if _, existed := r.byHash[hash]; !existed {
		r.hashByUser[userID] = append(r.hashByUser[userID], hash)
	}
	r.byHash[hash] = sess
	return nil
}

func (r *SessionRepository) ByTokenHash(_ context.Context, hash session.TokenHash) (session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sess, ok := r.byHash[hash.String()]
	if !ok {
		return session.Session{}, fmt.Errorf("%w: %s", session.ErrNotFound, hash.String())
	}
	return sess, nil
}

func (r *SessionRepository) DeleteAllForUser(_ context.Context, userID user.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	hashes, ok := r.hashByUser[userID.String()]
	if !ok {
		return nil
	}
	for _, hash := range hashes {
		delete(r.byHash, hash)
	}
	delete(r.hashByUser, userID.String())
	return nil
}

func (r *SessionRepository) DeleteExpired(_ context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var removed int64
	for hash, sess := range r.byHash {
		if sess.IsExpired(before) {
			delete(r.byHash, hash)
			removeFromIndex(r.hashByUser, sess.UserID().String(), hash)
			removed++
		}
	}
	return removed, nil
}

func removeFromIndex(index map[string][]string, userID, hash string) {
	hashes := index[userID]
	for i, h := range hashes {
		if h == hash {
			index[userID] = append(hashes[:i], hashes[i+1:]...)
			break
		}
	}
	if len(index[userID]) == 0 {
		delete(index, userID)
	}
}
