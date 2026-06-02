package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/pleading/model"
)

func TestNewPleadingStatus(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      model.Status
		expectErr bool
	}{
		{name: "open", value: "open", want: model.StatusOpen, expectErr: false},
		{name: "merged", value: "merged", want: model.StatusMerged, expectErr: false},
		{name: "closed", value: "closed", want: model.StatusClosed, expectErr: false},
		{name: "empty rejected", value: "", expectErr: true},
		{name: "unknown rejected", value: "approved", expectErr: true},
		{name: "uppercase rejected", value: "OPEN", expectErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := model.NewStatus(tcase.value)
			if tcase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidStatus)
				return
			}
			require.NoError(t, err)
			assert.True(t, tcase.want.Equal(got))
			assert.Equal(t, tcase.value, got.String())
		})
	}
}
