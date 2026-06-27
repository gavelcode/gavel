package judge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeTargetsToPattern(t *testing.T) {
	affected := []string{"//core/domain:model", "//apps/server:server", "//core/application:app"}

	scoped := scopeTargetsToPattern(affected, "//core/...", nil)

	assert.Equal(t, []string{"//core/domain:model", "//core/application:app"}, scoped)
}

func TestScopeTargetsToPattern_NoMatches(t *testing.T) {
	affected := []string{"//apps/server:server", "//apps/cli:cli"}

	scoped := scopeTargetsToPattern(affected, "//core/...", nil)

	assert.Empty(t, scoped)
}

func TestScopeTargetsToPattern_AllMatch(t *testing.T) {
	affected := []string{"//core/domain:model", "//core/application:app"}

	scoped := scopeTargetsToPattern(affected, "//core/...", nil)

	assert.Equal(t, affected, scoped)
}

func TestScopeTargetsToPattern_ExactTarget(t *testing.T) {
	affected := []string{"//core/domain:model", "//core/domain:model_test"}

	scoped := scopeTargetsToPattern(affected, "//core/domain:model", nil)

	assert.Equal(t, []string{"//core/domain:model"}, scoped)
}

func TestScopeTargetsToPattern_DropsExcluded(t *testing.T) {
	affected := []string{"//core/domain:model", "//core/gen:api", "//core/gen/sub:x"}

	scoped := scopeTargetsToPattern(affected, "//core/...", []string{"//core/gen/..."})

	assert.Equal(t, []string{"//core/domain:model"}, scoped)
}
