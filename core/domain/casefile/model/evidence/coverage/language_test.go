package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

func TestNewLanguage(t *testing.T) {
	lang, err := coverage.NewLanguage("java")

	require.NoError(t, err)
	assert.Equal(t, "java", lang.String())
}

func TestNewLanguageEmptyRejected(t *testing.T) {
	_, err := coverage.NewLanguage("")

	require.Error(t, err)
	assert.ErrorIs(t, err, coverage.ErrInvalidLanguage)
}

func TestNewLanguageNormalizesToLowercase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "capitalized", input: "Java", expected: "java"},
		{name: "uppercase", input: "GO", expected: "go"},
		{name: "mixed case", input: "TypeScript", expected: "typescript"},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			lang, err := coverage.NewLanguage(tcase.input)

			require.NoError(t, err)
			assert.Equal(t, tcase.expected, lang.String())
		})
	}
}

func TestLanguageString(t *testing.T) {
	lang, err := coverage.NewLanguage("python")

	require.NoError(t, err)
	assert.Equal(t, "python", lang.String())
}

func TestLanguageEqualityAfterNormalization(t *testing.T) {
	language, err := coverage.NewLanguage("Java")
	require.NoError(t, err)

	normalized, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	assert.Equal(t, language, normalized)
}

func TestLanguageEqual(t *testing.T) {
	goLanguage, err := coverage.NewLanguage("go")
	require.NoError(t, err)
	goUpperCase, err := coverage.NewLanguage("GO")
	require.NoError(t, err)
	pythonLanguage, err := coverage.NewLanguage("python")
	require.NoError(t, err)

	assert.True(t, goLanguage.Equal(goUpperCase), "equality ignores original casing")
	assert.False(t, goLanguage.Equal(pythonLanguage))
	assert.False(t, goLanguage.Equal(coverage.Language{}))
}
