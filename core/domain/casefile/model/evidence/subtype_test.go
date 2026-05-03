package evidence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

func TestNewEvidenceSubtype(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected evidence.Subtype
	}{
		{name: "code_quality", input: "code_quality", expected: evidence.SubtypeCodeQuality},
		{name: "complexity", input: "complexity", expected: evidence.SubtypeComplexity},
		{name: "sast", input: "sast", expected: evidence.SubtypeSAST},
		{name: "secrets", input: "secrets", expected: evidence.SubtypeSecrets},
		{name: "malware", input: "malware", expected: evidence.SubtypeMalware},
		{name: "dast", input: "dast", expected: evidence.SubtypeDAST},
		{name: "sca", input: "sca", expected: evidence.SubtypeSCA},
		{name: "license", input: "license", expected: evidence.SubtypeLicense},
		{name: "coverage", input: "coverage", expected: evidence.SubtypeCoverage},
		{name: "architecture", input: "architecture", expected: evidence.SubtypeArchitecture},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := evidence.NewSubtype(tcase.input)

			require.NoError(t, err)
			assert.Equal(t, tcase.expected, got)
		})
	}
}

func TestNewEvidenceSubtypeInvalidValue(t *testing.T) {
	_, err := evidence.NewSubtype("unknown")

	require.Error(t, err)
	assert.ErrorIs(t, err, evidence.ErrInvalidSubtype)
}

func TestEvidenceSubtypeString(t *testing.T) {
	tests := []struct {
		name     string
		subtype  evidence.Subtype
		expected string
	}{
		{name: "code_quality", subtype: evidence.SubtypeCodeQuality, expected: "code_quality"},
		{name: "complexity", subtype: evidence.SubtypeComplexity, expected: "complexity"},
		{name: "sast", subtype: evidence.SubtypeSAST, expected: "sast"},
		{name: "secrets", subtype: evidence.SubtypeSecrets, expected: "secrets"},
		{name: "malware", subtype: evidence.SubtypeMalware, expected: "malware"},
		{name: "dast", subtype: evidence.SubtypeDAST, expected: "dast"},
		{name: "sca", subtype: evidence.SubtypeSCA, expected: "sca"},
		{name: "license", subtype: evidence.SubtypeLicense, expected: "license"},
		{name: "coverage", subtype: evidence.SubtypeCoverage, expected: "coverage"},
		{name: "architecture", subtype: evidence.SubtypeArchitecture, expected: "architecture"},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			assert.Equal(t, tcase.expected, tcase.subtype.String())
		})
	}
}

func TestEvidenceSubtypeType(t *testing.T) {
	tests := []struct {
		name       string
		subtype    evidence.Subtype
		parentType evidence.Type
	}{
		{name: "code_quality belongs to source_code", subtype: evidence.SubtypeCodeQuality, parentType: evidence.TypeSourceCode},
		{name: "complexity belongs to source_code", subtype: evidence.SubtypeComplexity, parentType: evidence.TypeSourceCode},
		{name: "sast belongs to security", subtype: evidence.SubtypeSAST, parentType: evidence.TypeSecurity},
		{name: "secrets belongs to security", subtype: evidence.SubtypeSecrets, parentType: evidence.TypeSecurity},
		{name: "malware belongs to security", subtype: evidence.SubtypeMalware, parentType: evidence.TypeSecurity},
		{name: "dast belongs to security", subtype: evidence.SubtypeDAST, parentType: evidence.TypeSecurity},
		{name: "sca belongs to supply_chain", subtype: evidence.SubtypeSCA, parentType: evidence.TypeSupplyChain},
		{name: "license belongs to supply_chain", subtype: evidence.SubtypeLicense, parentType: evidence.TypeSupplyChain},
		{name: "coverage belongs to coverage", subtype: evidence.SubtypeCoverage, parentType: evidence.TypeCoverage},
		{name: "architecture belongs to architecture", subtype: evidence.SubtypeArchitecture, parentType: evidence.TypeArchitecture},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			assert.Equal(t, tcase.parentType, tcase.subtype.Type())
		})
	}
}

func TestEvidenceSubtypeEquality(t *testing.T) {
	subtype, err := evidence.NewSubtype("sast")
	require.NoError(t, err)

	sameSubtype, err := evidence.NewSubtype("sast")
	require.NoError(t, err)

	assert.Equal(t, subtype, sameSubtype)
}

func TestEvidenceSubtypeEqual(t *testing.T) {
	assert.True(t, evidence.SubtypeSAST.Equal(evidence.SubtypeSAST))
	assert.False(t, evidence.SubtypeSAST.Equal(evidence.SubtypeSecrets),
		"different subtypes within the same parent type are not equal")
	assert.False(t, evidence.SubtypeSAST.Equal(evidence.Subtype{}))
}

func TestIsSubtypeFindingBased(t *testing.T) {
	tests := []struct {
		name     string
		subtype  evidence.Subtype
		expected bool
	}{
		{name: "code_quality is finding-based", subtype: evidence.SubtypeCodeQuality, expected: true},
		{name: "complexity is finding-based", subtype: evidence.SubtypeComplexity, expected: true},
		{name: "sast is finding-based", subtype: evidence.SubtypeSAST, expected: true},
		{name: "secrets is finding-based", subtype: evidence.SubtypeSecrets, expected: true},
		{name: "malware is finding-based", subtype: evidence.SubtypeMalware, expected: true},
		{name: "dast is finding-based", subtype: evidence.SubtypeDAST, expected: true},
		{name: "sca is finding-based", subtype: evidence.SubtypeSCA, expected: true},
		{name: "license is not finding-based", subtype: evidence.SubtypeLicense, expected: false},
		{name: "coverage is not finding-based", subtype: evidence.SubtypeCoverage, expected: false},
		{name: "architecture is not finding-based", subtype: evidence.SubtypeArchitecture, expected: false},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			assert.Equal(t, tcase.expected, evidence.IsSubtypeFindingBased(tcase.subtype))
		})
	}
}
