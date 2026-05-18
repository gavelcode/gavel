package sarif_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
)

func TestParseMinimalSARIF(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "minimal.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)

	first := findings[0]
	assert.Equal(t, "NP_NULL_ON_SOME_PATH", first.RuleID)
	assert.Equal(t, finding.SeverityError, first.Severity)
	assert.Equal(t, "src/main/java/Foo.java", first.FilePath)
	assert.Equal(t, 42, first.Line)
	assert.Equal(t, "Possible null dereference", first.Message)
	assert.NotEmpty(t, first.FingerprintID.Value())
}

func TestParseMultipleRuns(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "multi_run.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 3)

	assert.Equal(t, "NP_NULL", findings[0].RuleID)
	assert.Equal(t, finding.SeverityError, findings[0].Severity)

	assert.Equal(t, "UnusedLocalVariable", findings[1].RuleID)
	assert.Equal(t, finding.SeverityWarning, findings[1].Severity)

	assert.Equal(t, "EmptyCatchBlock", findings[2].RuleID)
	assert.Equal(t, finding.SeverityNote, findings[2].Severity)
}

func TestParsePartialFingerprintsOverrideComputed(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "partial_fingerprints.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "abc123stablehash", findings[0].FingerprintID.Value())
}

func TestParsePicksFingerprintAtLexicographicallySmallestKey(t *testing.T) {
	parser := sarif.NewParser()

	for iter := range 50 {
		findings, err := parser.Parse(context.Background(), readFixture(t, "multiple_fingerprints.sarif.json"))
		require.NoError(t, err)
		require.Len(t, findings, 1)
		assert.Equal(t,
			"expected-stable-value",
			findings[0].FingerprintID.Value(),
			"iteration %d: fingerprint must come from the lexicographically smallest key", iter)
	}
}

func TestParseRuleLevelFallback(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "rule_level_fallback.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, finding.SeverityError, findings[0].Severity)
	assert.Equal(t, finding.SeverityNote, findings[1].Severity)
}

func TestParseEmptyDataReturnsNil(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), nil)

	require.NoError(t, err)
	assert.Nil(t, findings)
}

func TestParseEmptyRunsReturnsNil(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "empty_runs.sarif.json"))

	require.NoError(t, err)
	assert.Nil(t, findings)
}

func TestParseNoLocationsFallsBackToUnknown(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "no_locations.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "unknown", findings[0].FilePath)
	assert.Equal(t, 0, findings[0].Line)
}

func TestParseInvalidJSONRejected(t *testing.T) {
	parser := sarif.NewParser()
	_, err := parser.Parse(context.Background(), readFixture(t, "invalid.sarif.json"))

	require.Error(t, err)
	assert.ErrorIs(t, err, sarif.ErrDecodeSARIF)
}

func TestParseMissingRuleIDRejected(t *testing.T) {
	parser := sarif.NewParser()
	_, err := parser.Parse(context.Background(), readFixture(t, "missing_rule_id.sarif.json"))

	require.Error(t, err)
	assert.ErrorIs(t, err, sarif.ErrInvalidResult)
}

func TestParseSandboxPathsNormalized(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "sandbox_paths.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "internal/domain/order/order.go", findings[0].FilePath)
	assert.Equal(t, 96, findings[0].Line)
}

func TestParseFileURIPathsNormalized(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "file_uri_paths.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "src/main/java/com/example/Foo.java", findings[0].FilePath)
	assert.Equal(t, 10, findings[0].Line)
}

func TestParseSandboxPathFingerprintMatchesCleanPath(t *testing.T) {
	parser := sarif.NewParser()

	sandboxData := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[{"ruleId":"errcheck","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"../../../../execroot/_main/internal/order.go"},"region":{"startLine":10}}}]}]}]}`)
	cleanData := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[{"ruleId":"errcheck","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"internal/order.go"},"region":{"startLine":10}}}]}]}]}`)

	sandbox, err := parser.Parse(context.Background(), sandboxData)
	require.NoError(t, err)
	clean, err := parser.Parse(context.Background(), cleanData)
	require.NoError(t, err)

	require.Len(t, sandbox, 1)
	require.Len(t, clean, 1)
	assert.Equal(t, sandbox[0].FingerprintID.Value(), clean[0].FingerprintID.Value())
}

func TestParseCleanPathUnchanged(t *testing.T) {
	parser := sarif.NewParser()
	findings, err := parser.Parse(context.Background(), readFixture(t, "minimal.sarif.json"))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "src/main/java/Foo.java", findings[0].FilePath)
}

func TestParseComputedFingerprintIsDeterministic(t *testing.T) {
	parser := sarif.NewParser()
	data := readFixture(t, "minimal.sarif.json")

	first, err := parser.Parse(context.Background(), data)
	require.NoError(t, err)
	second, err := parser.Parse(context.Background(), data)
	require.NoError(t, err)

	require.Len(t, first, 1)
	require.Len(t, second, 1)
	assert.Equal(t, first[0].FingerprintID.Value(), second[0].FingerprintID.Value())
}

func TestParseContentBasedFingerprintStableAcrossLineShifts(t *testing.T) {
	src := &fakeSourceReader{lines: map[string]map[int]string{
		"internal/order.go": {10: "func (o *Order) Total() Money {", 8: "func (o *Order) Total() Money {"},
	}}
	parser := sarif.NewParser(sarif.WithSourceReader(src))

	atLine10 := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[{"ruleId":"errcheck","level":"error","message":{"text":"unchecked"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"internal/order.go"},"region":{"startLine":10}}}]}]}]}`)
	atLine8 := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[{"ruleId":"errcheck","level":"error","message":{"text":"unchecked"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"internal/order.go"},"region":{"startLine":8}}}]}]}]}`)

	f10, err := parser.Parse(context.Background(), atLine10)
	require.NoError(t, err)
	f8, err := parser.Parse(context.Background(), atLine8)
	require.NoError(t, err)

	require.Len(t, f10, 1)
	require.Len(t, f8, 1)
	assert.Equal(t, f10[0].FingerprintID.Value(), f8[0].FingerprintID.Value(), "same content at different lines must produce same fingerprint")
}

func TestParseContentBasedFingerprintDiffersForDifferentContent(t *testing.T) {
	src := &fakeSourceReader{lines: map[string]map[int]string{
		"file.go": {10: "func Foo() {", 20: "func Bar() {"},
	}}
	parser := sarif.NewParser(sarif.WithSourceReader(src))

	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[
		{"ruleId":"rule1","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":10}}}]},
		{"ruleId":"rule1","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":20}}}]}
	]}]}`)

	findings, err := parser.Parse(context.Background(), data)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.NotEqual(t, findings[0].FingerprintID.Value(), findings[1].FingerprintID.Value(), "different content must produce different fingerprints")
}

func TestParseCollisionCounterDisambiguates(t *testing.T) {
	src := &fakeSourceReader{lines: map[string]map[int]string{
		"file.go": {10: "x := 42", 20: "x := 42"},
	}}
	parser := sarif.NewParser(sarif.WithSourceReader(src))

	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[
		{"ruleId":"mnd","level":"error","message":{"text":"magic"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":10}}}]},
		{"ruleId":"mnd","level":"error","message":{"text":"magic"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":20}}}]}
	]}]}`)

	findings, err := parser.Parse(context.Background(), data)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.NotEqual(t, findings[0].FingerprintID.Value(), findings[1].FingerprintID.Value(), "collisions must be disambiguated")
}

func TestParseFallsBackToLineNumberWithoutSourceReader(t *testing.T) {
	parserWithout := sarif.NewParser()
	parserWith := sarif.NewParser(sarif.WithSourceReader(&fakeSourceReader{lines: map[string]map[int]string{}}))

	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[{"ruleId":"rule1","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":10}}}]}]}]}`)

	without, err := parserWithout.Parse(context.Background(), data)
	require.NoError(t, err)
	withReader, err := parserWith.Parse(context.Background(), data)
	require.NoError(t, err)

	require.Len(t, without, 1)
	require.Len(t, withReader, 1)
	assert.Equal(t, without[0].FingerprintID.Value(), withReader[0].FingerprintID.Value(), "empty source reader must fall back to line-number hash")
}

func TestParseRuleWithEmptyIDIsSkipped(t *testing.T) {
	parser := sarif.NewParser()
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint","rules":[{"id":"","defaultConfiguration":{"level":"error"}},{"id":"rule1","defaultConfiguration":{"level":"note"}}]}},"results":[{"ruleId":"rule1","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":1}}}]}]}]}`)

	findings, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, finding.SeverityNote, findings[0].Severity)
}

func TestParseUnknownLevelDefaultsToWarning(t *testing.T) {
	parser := sarif.NewParser()
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[{"ruleId":"rule1","level":"critical","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":1}}}]}]}]}`)

	findings, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, finding.SeverityWarning, findings[0].Severity)
}

func TestParseEmptyLocationURIFallsBackToUnknown(t *testing.T) {
	parser := sarif.NewParser()
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[{"ruleId":"rule1","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":""},"region":{"startLine":5}}}]}]}]}`)

	findings, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "unknown", findings[0].FilePath)
	assert.Equal(t, 5, findings[0].Line)
}

func TestParseSourceReaderReturnsEmptyContentFallsBackToLineHash(t *testing.T) {
	src := &fakeSourceReader{lines: map[string]map[int]string{
		"file.go": {10: ""},
	}}
	parser := sarif.NewParser(sarif.WithSourceReader(src))

	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[{"ruleId":"rule1","level":"error","message":{"text":"msg"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":10}}}]}]}]}`)

	findings, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.NotEmpty(t, findings[0].FingerprintID.Value())
}

func TestParseFingerprintMapWithAllEmptyValues(t *testing.T) {
	parser := sarif.NewParser()
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"lint"}},"results":[{"ruleId":"rule1","level":"error","message":{"text":"msg"},"fingerprints":{"key1":"","key2":""},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"file.go"},"region":{"startLine":1}}}]}]}]}`)

	findings, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.NotEmpty(t, findings[0].FingerprintID.Value())
}

type fakeSourceReader struct {
	lines map[string]map[int]string
}

func (f *fakeSourceReader) ReadLine(filePath string, line int) (string, error) {
	fileLines, ok := f.lines[filePath]
	if !ok {
		return "", fmt.Errorf("file not found: %s", filePath)
	}
	content, ok := fileLines[line]
	if !ok {
		return "", fmt.Errorf("line %d not found in %s", line, filePath)
	}
	return content, nil
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}
