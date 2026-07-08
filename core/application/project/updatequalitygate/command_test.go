package updatequalitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatequalitygate"
)

func TestNewCommand(t *testing.T) {
	minDelta := 5.0
	tests := []struct {
		name      string
		projectID string
		input     updatequalitygate.Input
		wantErr   bool
	}{
		{name: "valid empty gate", projectID: "proj-1", input: updatequalitygate.Input{}},
		{
			name:      "valid with rule",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "code_quality",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeZeroTolerance},
					},
				},
			},
		},
		{
			name:      "valid with minDelta option",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "code_quality",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeZeroTolerance},
						MinDelta: &minDelta,
					},
				},
			},
		},
		{
			name:      "valid forbiddenList strategy",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype: "license",
						Strategy: updatequalitygate.StrategyInput{
							Type:          updatequalitygate.StrategyTypeForbiddenList,
							ForbiddenList: &updatequalitygate.ForbiddenList{Forbidden: []string{"GPL-3.0"}},
						},
					},
				},
			},
		},
		{name: "empty project id rejected", projectID: "", wantErr: true},
		{name: "blank project id rejected", projectID: "   ", wantErr: true},
		{
			name:      "invalid subtype rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "INVALID",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeZeroTolerance},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "unknown strategy rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "code_quality",
						Strategy: updatequalitygate.StrategyInput{Type: "nonsense"},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "missing countBySeverity payload rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "code_quality",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeCountBySeverity},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "missing minPercentage payload rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "coverage",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeMinPercentage},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "missing forbiddenList payload rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "code_quality",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeForbiddenList},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "missing maxViolations payload rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "architecture",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeMaxViolations},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "missing minNewCodeCoverage payload rejected",
			projectID: "proj-1",
			input: updatequalitygate.Input{
				Rules: []updatequalitygate.RuleInput{
					{
						Subtype:  "new_code_coverage",
						Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeMinNewCodeCoverage},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := updatequalitygate.NewCommand(testTenant.String(), testCase.projectID, testCase.input)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, updatequalitygate.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
			assert.Equal(t, len(testCase.input.Rules), len(cmd.Gate().Rules()))
		})
	}
}
