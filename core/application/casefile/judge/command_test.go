package judge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/judge"
)

func TestNewCommand(t *testing.T) {
	tracking := evidencedto.Tracking{}

	tests := []struct {
		name       string
		caseFileID string
		tracking   *evidencedto.Tracking
		wantErr    bool
	}{
		{name: "valid with tracking", caseFileID: "case-1", tracking: &tracking},
		{name: "valid without tracking", caseFileID: "case-1", tracking: nil},
		{name: "empty caseFileID rejected", caseFileID: "", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := judge.NewCommand(testTenant.String(), testCase.caseFileID, testCase.tracking)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, judge.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.caseFileID, cmd.CaseFileID())
			assert.Equal(t, testCase.tracking, cmd.Tracking())
		})
	}
}
