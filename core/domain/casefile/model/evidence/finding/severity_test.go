package finding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func TestNewSeverity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected finding.Severity
	}{
		{name: "error", input: "error", expected: finding.SeverityError},
		{name: "warning", input: "warning", expected: finding.SeverityWarning},
		{name: "note", input: "note", expected: finding.SeverityNote},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := finding.NewSeverity(tcase.input)

			require.NoError(t, err)
			assert.Equal(t, tcase.expected, got)
		})
	}
}

func TestNewSeverityInvalidValue(t *testing.T) {
	_, err := finding.NewSeverity("critical")

	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInvalidSeverity)
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity finding.Severity
		expected string
	}{
		{name: "error", severity: finding.SeverityError, expected: "error"},
		{name: "warning", severity: finding.SeverityWarning, expected: "warning"},
		{name: "note", severity: finding.SeverityNote, expected: "note"},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			assert.Equal(t, tcase.expected, tcase.severity.String())
		})
	}
}

func TestSeverityEquality(t *testing.T) {
	severity, err := finding.NewSeverity("error")
	require.NoError(t, err)

	sameSeverity, err := finding.NewSeverity("error")
	require.NoError(t, err)

	assert.Equal(t, severity, sameSeverity)
}

func TestSeverityEqual(t *testing.T) {
	assert.True(t, finding.SeverityError.Equal(finding.SeverityError))
	assert.False(t, finding.SeverityError.Equal(finding.SeverityWarning))
	assert.False(t, finding.SeverityWarning.Equal(finding.Severity{}))
}
