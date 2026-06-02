package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/pleading/resolve"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name       string
		pleadingID string
		outcome    string
		expectErr  bool
	}{
		{name: "merged valid", pleadingID: "p1", outcome: "merged", expectErr: false},
		{name: "closed valid", pleadingID: "p1", outcome: "closed", expectErr: false},
		{name: "open rejected", pleadingID: "p1", outcome: "open", expectErr: true},
		{name: "unknown outcome rejected", pleadingID: "p1", outcome: "approved", expectErr: true},
		{name: "empty outcome rejected", pleadingID: "p1", outcome: "", expectErr: true},
		{name: "empty pleadingID rejected", pleadingID: "", outcome: "merged", expectErr: true},
		{name: "whitespace pleadingID rejected", pleadingID: "   ", outcome: "merged", expectErr: true},
		{name: "uppercase outcome rejected", pleadingID: "p1", outcome: "MERGED", expectErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := resolve.NewCommand(testCase.pleadingID, testCase.outcome)
			if testCase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, resolve.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.pleadingID, cmd.PleadingID())
			assert.Equal(t, testCase.outcome, cmd.Outcome())
		})
	}
}
