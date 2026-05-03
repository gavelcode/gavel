package architecture_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
)

func TestNewArchitectureViolation(t *testing.T) {
	tests := []struct {
		name      string
		rule      string
		sourcePkg string
		targetPkg string
		message   string
		wantErr   bool
	}{
		{
			name:      "shouldCreateWithAllFields",
			rule:      "layer-dependency",
			sourcePkg: "internal/domain/foo",
			targetPkg: "internal/infrastructure/bar",
			message:   "domain imports infrastructure",
		},
		{
			name:      "shouldCreateWithEmptyTargetPkg",
			rule:      "circular-dependency",
			sourcePkg: "internal/domain",
			targetPkg: "",
			message:   "circular dependency detected",
		},
		{
			name:      "shouldRejectEmptyRule",
			rule:      "",
			sourcePkg: "internal/domain",
			targetPkg: "internal/infra",
			message:   "msg",
			wantErr:   true,
		},
		{
			name:      "shouldRejectBlankRule",
			rule:      "   ",
			sourcePkg: "internal/domain",
			targetPkg: "internal/infra",
			message:   "msg",
			wantErr:   true,
		},
		{
			name:      "shouldRejectEmptySourcePkg",
			rule:      "layer-dependency",
			sourcePkg: "",
			targetPkg: "internal/infra",
			message:   "msg",
			wantErr:   true,
		},
		{
			name:      "shouldRejectEmptyMessage",
			rule:      "layer-dependency",
			sourcePkg: "internal/domain",
			targetPkg: "internal/infra",
			message:   "",
			wantErr:   true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			archViol, err := architecture.NewViolation(tcase.rule, tcase.sourcePkg, tcase.targetPkg, tcase.message)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, architecture.ErrInvalidViolation)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.rule, archViol.Rule())
			assert.Equal(t, tcase.sourcePkg, archViol.SourcePkg())
			assert.Equal(t, tcase.targetPkg, archViol.TargetPkg())
			assert.Equal(t, tcase.message, archViol.Message())
		})
	}
}
