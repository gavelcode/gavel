package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestNewProjectRef(t *testing.T) {
	validID := projectmodel.NewProjectID(uuid.New())

	tests := []struct {
		name          string
		id            projectmodel.ProjectID
		targetPattern string
		wantError     bool
	}{
		{name: "valid", id: validID, targetPattern: "//src/..."},
		{name: "empty target pattern rejected", id: validID, targetPattern: "", wantError: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			ref, err := model.NewProjectRef(tcase.id, tcase.targetPattern)

			if tcase.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidGavelspace)
				return
			}

			require.NoError(t, err)
			assert.True(t, tcase.id.Equal(ref.ID()))
			assert.Equal(t, tcase.targetPattern, ref.TargetPattern())
		})
	}
}
