package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func mustUserID(t *testing.T) user.UserID {
	t.Helper()
	return user.NewUserID(uuid.New())
}
