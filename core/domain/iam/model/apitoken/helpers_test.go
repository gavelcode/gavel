package apitoken_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var iamTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func mustTenantID(t *testing.T) tenant.TenantID {
	t.Helper()
	return tenant.NewTenantID(uuid.New())
}

func mustUserID(t *testing.T) user.UserID {
	t.Helper()
	return user.NewUserID(uuid.New())
}
