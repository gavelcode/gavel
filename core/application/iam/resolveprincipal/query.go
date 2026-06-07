package resolveprincipal

import (
	"fmt"
	"time"
)

type Query struct {
	sessionCookie string
	bearerToken   string
	occurredAt    time.Time
}

func NewQuery(sessionCookie, bearerToken string, occurredAt time.Time) (Query, error) {
	if occurredAt.IsZero() {
		return Query{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidQuery)
	}
	return Query{
		sessionCookie: sessionCookie,
		bearerToken:   bearerToken,
		occurredAt:    occurredAt,
	}, nil
}

func (q Query) SessionCookie() string { return q.sessionCookie }
func (q Query) BearerToken() string   { return q.bearerToken }
func (q Query) OccurredAt() time.Time { return q.occurredAt }
