package sarif_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
)

func TestParseArchitectureViolations_ValidDocument(t *testing.T) {
	data := []byte(`{
		"runs": [{
			"results": [
				{
					"ruleId": "layer_violation",
					"message": {"text": "forbidden dependency"},
					"properties": {"sourcePkg": "com.api", "targetPkg": "com.domain"}
				},
				{
					"ruleId": "dependency_rule",
					"message": {"text": "not allowed"},
					"properties": {"sourcePkg": "com.infra", "targetPkg": "com.app"}
				}
			]
		}]
	}`)

	violations, err := sarif.ParseArchitectureViolations(data)

	require.NoError(t, err)
	require.Len(t, violations, 2)
	assert.Equal(t, "layer_violation", violations[0].Rule)
	assert.Equal(t, "com.api", violations[0].SourcePkg)
	assert.Equal(t, "com.domain", violations[0].TargetPkg)
	assert.Equal(t, "forbidden dependency", violations[0].Message)
	assert.Equal(t, "dependency_rule", violations[1].Rule)
}

func TestParseArchitectureViolations_EmptyResults(t *testing.T) {
	data := []byte(`{"runs": [{"results": []}]}`)

	violations, err := sarif.ParseArchitectureViolations(data)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestParseArchitectureViolations_InvalidJSON(t *testing.T) {
	_, err := sarif.ParseArchitectureViolations([]byte(`not json`))

	assert.Error(t, err)
}

func TestParseArchitectureViolations_MultipleRuns(t *testing.T) {
	data := []byte(`{
		"runs": [
			{"results": [{"ruleId": "r1", "message": {"text": "m1"}, "properties": {"sourcePkg": "a", "targetPkg": "b"}}]},
			{"results": [{"ruleId": "r2", "message": {"text": "m2"}, "properties": {"sourcePkg": "c", "targetPkg": "d"}}]}
		]
	}`)

	violations, err := sarif.ParseArchitectureViolations(data)

	require.NoError(t, err)
	assert.Len(t, violations, 2)
}
