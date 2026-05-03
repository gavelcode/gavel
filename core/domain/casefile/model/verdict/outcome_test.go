package verdict_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
)

func TestNewOutcome(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantStr string
		wantErr bool
	}{
		{
			name:    "pass",
			input:   "pass",
			wantStr: "pass",
		},
		{
			name:    "fail",
			input:   "fail",
			wantStr: "fail",
		},
		{
			name:    "invalid outcome rejected",
			input:   "unknown",
			wantErr: true,
		},
		{
			name:    "empty string rejected",
			input:   "",
			wantErr: true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			outcome, err := verdict.NewOutcome(tcase.input)

			if tcase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.wantStr, outcome.String())
		})
	}
}

func TestOutcomeEqual(t *testing.T) {
	assert.True(t, verdict.OutcomePass.Equal(verdict.OutcomePass))
	assert.False(t, verdict.OutcomePass.Equal(verdict.OutcomeFail))
}

func TestShouldRecordAsBaseline(t *testing.T) {
	assert.True(t, verdict.OutcomePass.ShouldRecordAsBaseline())
	assert.False(t, verdict.OutcomeFail.ShouldRecordAsBaseline())
}
