package evidence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

func TestNewEvidenceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected evidence.Type
	}{
		{name: "source_code", input: "source_code", expected: evidence.TypeSourceCode},
		{name: "security", input: "security", expected: evidence.TypeSecurity},
		{name: "supply_chain", input: "supply_chain", expected: evidence.TypeSupplyChain},
		{name: "coverage", input: "coverage", expected: evidence.TypeCoverage},
		{name: "architecture", input: "architecture", expected: evidence.TypeArchitecture},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := evidence.NewType(tcase.input)

			require.NoError(t, err)
			assert.Equal(t, tcase.expected, got)
		})
	}
}

func TestNewEvidenceTypeInvalidValue(t *testing.T) {
	_, err := evidence.NewType("unknown")

	require.Error(t, err)
	assert.ErrorIs(t, err, evidence.ErrInvalidType)
}

func TestEvidenceTypeString(t *testing.T) {
	tests := []struct {
		name     string
		typ      evidence.Type
		expected string
	}{
		{name: "source_code", typ: evidence.TypeSourceCode, expected: "source_code"},
		{name: "security", typ: evidence.TypeSecurity, expected: "security"},
		{name: "supply_chain", typ: evidence.TypeSupplyChain, expected: "supply_chain"},
		{name: "coverage", typ: evidence.TypeCoverage, expected: "coverage"},
		{name: "architecture", typ: evidence.TypeArchitecture, expected: "architecture"},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			assert.Equal(t, tcase.expected, tcase.typ.String())
		})
	}
}

func TestEvidenceTypeEquality(t *testing.T) {
	evidenceType, err := evidence.NewType("security")
	require.NoError(t, err)

	sameType, err := evidence.NewType("security")
	require.NoError(t, err)

	assert.Equal(t, evidenceType, sameType)
}

func TestEvidenceTypeEqual(t *testing.T) {
	assert.True(t, evidence.TypeSecurity.Equal(evidence.TypeSecurity))
	assert.False(t, evidence.TypeSecurity.Equal(evidence.TypeSourceCode))
	assert.False(t, evidence.TypeSecurity.Equal(evidence.Type{}))
}
