package lcov_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/infrastructure/casefile/lcov"
)

func TestParseSingleLanguage(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "single_lang.lcov"))

	require.NoError(t, err)
	assert.Equal(t, 4, result.TotalLines)
	assert.Equal(t, 3, result.CoveredLines)

	require.Len(t, result.ByLanguage, 1)
	lc := result.ByLanguage[0]
	assert.Equal(t, "go", lc.Language().String())
	assert.Equal(t, 4, lc.TotalLines())
	assert.Equal(t, 3, lc.CoveredLines())
}

func TestParseMultipleLanguages(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "multi_lang.lcov"))

	require.NoError(t, err)
	assert.Equal(t, 35, result.TotalLines)
	assert.Equal(t, 28, result.CoveredLines)

	languages := languageNames(result.ByLanguage)
	assert.ElementsMatch(t, []string{"go", "java", "typescript"}, languages)
}

func TestParseUnknownExtensionFallsBackToOther(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "unknown_extension.lcov"))

	require.NoError(t, err)
	require.Len(t, result.ByLanguage, 1)
	assert.Equal(t, "other", result.ByLanguage[0].Language().String())
	assert.Equal(t, 4, result.ByLanguage[0].TotalLines())
	assert.Equal(t, 2, result.ByLanguage[0].CoveredLines())
}

func TestParseMalformedCountRejected(t *testing.T) {
	parser := lcov.NewParser()
	_, err := parser.Parse(context.Background(), readFixture(t, "malformed_count.lcov"))

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrInvalidLine)
}

func TestParseScannerError(t *testing.T) {
	parser := lcov.NewParser()
	longLine := make([]byte, 70000)
	for i := range longLine {
		longLine[i] = 'x'
	}
	data := append([]byte("SF:src/foo.go\n"), longLine...)

	_, err := parser.Parse(context.Background(), data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrScanLCOV)
}

func TestParseCoveredExceedsTotalReturnsError(t *testing.T) {
	parser := lcov.NewParser()
	data := []byte("SF:src/foo.go\nLF:2\nLH:5\nend_of_record\n")

	_, err := parser.Parse(context.Background(), data)

	require.Error(t, err)
}

func TestParseMalformedHitCountRejected(t *testing.T) {
	parser := lcov.NewParser()
	data := []byte("SF:src/foo.go\nLF:10\nLH:not-a-number\nend_of_record\n")

	_, err := parser.Parse(context.Background(), data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrInvalidLine)
}

func TestParseEmptyDataReturnsZero(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), nil)

	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalLines)
	assert.Equal(t, 0, result.CoveredLines)
	assert.Empty(t, result.ByLanguage)
}

func TestParseAggregatesAcrossFilesSameLanguage(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "single_lang.lcov"))

	require.NoError(t, err)
	require.Len(t, result.ByLanguage, 1)
	assert.Equal(t, 4, result.ByLanguage[0].TotalLines(), "two .go files should aggregate into a single language entry")
}

func TestParseTotalsMatchSumOfLanguages(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "multi_lang.lcov"))
	require.NoError(t, err)

	var total, covered int
	for _, lc := range result.ByLanguage {
		total += lc.TotalLines()
		covered += lc.CoveredLines()
	}
	assert.Equal(t, result.TotalLines, total)
	assert.Equal(t, result.CoveredLines, covered)
}

func TestParsePopulatesByFile(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "single_lang.lcov"))

	require.NoError(t, err)
	require.NotEmpty(t, result.ByFile)

	fileMap := make(map[string]struct{}, len(result.ByFile))
	for _, fc := range result.ByFile {
		fileMap[fc.FilePath] = struct{}{}
		assert.NotEmpty(t, fc.Covered, "file %s should have covered lines", fc.FilePath)
	}
	assert.Contains(t, fileMap, "src/foo.go")
	assert.Contains(t, fileMap, "src/bar.go")
}

func TestParseEmptyDataReturnsEmptyByFile(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, result.ByFile)
}

func TestParseWithoutSFRecordsCountsGlobally(t *testing.T) {
	parser := lcov.NewParser()
	result, err := parser.Parse(context.Background(), readFixture(t, "no_sf_records.lcov"))

	require.NoError(t, err)
	assert.Equal(t, 10, result.TotalLines)
	assert.Equal(t, 5, result.CoveredLines)
	assert.Empty(t, result.ByLanguage, "without SF, no language entries are created")
}

func TestParseDuplicateFileDeduplicatesTotals(t *testing.T) {
	parser := lcov.NewParser()
	data := []byte(
		"SF:shared.go\nDA:1,10\nDA:2,5\nDA:3,0\nLF:3\nLH:2\nend_of_record\n" +
			"SF:shared.go\nDA:1,0\nDA:2,8\nDA:3,0\nLF:3\nLH:1\nend_of_record\n" +
			"SF:unique.go\nDA:1,1\nLF:1\nLH:1\nend_of_record\n",
	)

	result, err := parser.Parse(context.Background(), data)

	require.NoError(t, err)
	assert.Equal(t, 4, result.TotalLines, "shared.go (3 lines) + unique.go (1 line), deduplicated")
	assert.Equal(t, 3, result.CoveredLines, "shared.go lines 1+2 covered (max merge), unique.go line 1 covered")
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}

func languageNames(coverages []coverage.LanguageStats) []string {
	out := make([]string, 0, len(coverages))
	for _, lc := range coverages {
		out = append(out, lc.Language().String())
	}
	return out
}
