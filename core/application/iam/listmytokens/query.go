package listmytokens

import (
	"fmt"
	"strings"
	"time"
)

type Query struct {
	userID string
	now    time.Time
}

func NewQuery(userID string, now time.Time) (Query, error) {
	if strings.TrimSpace(userID) == "" {
		return Query{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidQuery)
	}
	if now.IsZero() {
		return Query{}, fmt.Errorf("%w: now must not be zero", ErrInvalidQuery)
	}
	return Query{userID: userID, now: now}, nil
}

func (q Query) UserID() string { return q.userID }
func (q Query) Now() time.Time { return q.now }
