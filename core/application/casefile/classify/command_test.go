package classify_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func TestNewCommand(t *testing.T) {
	findings := []finding.Finding{mustFinding(t, "fp-1")}

	tests := []struct {
		name      string
		projectID string
		branch    string
		findings  []finding.Finding
		wantErr   bool
	}{
		{name: "valid", projectID: "proj-1", branch: "main", findings: findings},
		{name: "valid empty findings", projectID: "proj-1", branch: "main", findings: nil},
		{name: "empty project id rejected", projectID: "", branch: "main", findings: findings, wantErr: true},
		{name: "blank project id rejected", projectID: "   ", branch: "main", findings: findings, wantErr: true},
		{name: "empty branch rejected", projectID: "proj-1", branch: "", findings: findings, wantErr: true},
		{name: "blank branch rejected", projectID: "proj-1", branch: "   ", findings: findings, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := classify.NewCommand(testCase.projectID, testCase.branch, testCase.findings)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, classify.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
			assert.Equal(t, testCase.branch, cmd.Branch())
			assert.Equal(t, len(testCase.findings), len(cmd.Findings()))
		})
	}
}

func TestNewCommandDefensiveCopy(t *testing.T) {
	findings := []finding.Finding{mustFinding(t, "fp-1")}
	cmd, err := classify.NewCommand("proj-1", "main", findings)
	require.NoError(t, err)

	findings[0] = mustFinding(t, "fp-mutated")

	assert.Equal(t, "fp-1", cmd.Findings()[0].ID().Value(),
		"command takes defensive copy of findings")
}

func mustFinding(t *testing.T, fpValue string) finding.Finding {
	t.Helper()
	fp, err := finding.NewFingerprintID(fpValue)
	require.NoError(t, err)
	f, err := finding.NewFinding("tool", "rule1", finding.SeverityWarning, "file.go", 1, "msg", fp)
	require.NoError(t, err)
	return f
}
