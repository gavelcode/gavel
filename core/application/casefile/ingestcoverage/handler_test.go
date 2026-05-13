package ingestcoverage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingest "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

func TestHandlerExecuteSuccessfulIngest(t *testing.T) {
	parser := &fakeParser{parsed: ingest.Parsed{
		TotalLines:   100,
		CoveredLines: 80,
		ByLanguage:   []coverage.LanguageStats{validLanguageCoverage(t, "go", 100, 80)},
	}}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, parser.calls)
	assert.Equal(t, evidence.SubtypeCoverage.String(), result.Evidence.Subtype)
	assert.Equal(t, "go-test", result.Evidence.Source)

	require.NotNil(t, result.Evidence.Coverage)
	assert.Equal(t, 100, result.Evidence.Coverage.TotalLines)
	assert.Equal(t, 80, result.Evidence.Coverage.CoveredLines)
	require.Len(t, result.Evidence.Coverage.ByLanguage, 1)
}

func TestHandlerExecuteUnknownFormat(t *testing.T) {
	parser := &fakeParser{}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("X"), "cobertura", "go-test")
	_, err := handler.Execute(context.Background(), cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, ingest.ErrUnknownFormat)
	assert.Equal(t, 0, parser.calls)
}

func TestHandlerExecuteParserError(t *testing.T) {
	boom := errors.New("boom")
	parser := &fakeParser{err: boom}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	_, err := handler.Execute(context.Background(), cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, ingest.ErrParseFailed)
	assert.ErrorIs(t, err, boom)
}

func TestHandlerExecuteInvalidCoverageRejected(t *testing.T) {
	parser := &fakeParser{parsed: ingest.Parsed{
		TotalLines:   10,
		CoveredLines: 50,
	}}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	_, err := handler.Execute(context.Background(), cmd)

	require.Error(t, err)
}

func TestHandlerExecuteEmptyCoverageProducesValidEvidence(t *testing.T) {
	parser := &fakeParser{parsed: ingest.Parsed{}}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	require.NotNil(t, result.Evidence.Coverage)
	assert.Equal(t, 0, result.Evidence.Coverage.TotalLines)
	assert.Equal(t, 0, result.Evidence.Coverage.CoveredLines)
}

func TestHandlerExecuteDispatchesByFormat(t *testing.T) {
	lcovParser := &fakeParser{parsed: ingest.Parsed{TotalLines: 1, CoveredLines: 1}}
	cobParser := &fakeParser{}
	handler := ingest.NewHandler(map[string]ingest.Parser{
		"lcov":      lcovParser,
		"cobertura": cobParser,
	})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, lcovParser.calls)
	assert.Equal(t, 0, cobParser.calls)
}

func TestHandlerExecuteByFileFlowsThrough(t *testing.T) {
	parser := &fakeParser{parsed: ingest.Parsed{
		TotalLines:   100,
		CoveredLines: 80,
		ByLanguage:   []coverage.LanguageStats{validLanguageCoverage(t, "go", 100, 80)},
		ByFile: []evidencedto.FileCoverage{
			{FilePath: "src/main.go", Covered: []int{1, 3, 5}, Uncovered: []int{2, 4}},
			{FilePath: "src/util.go", Covered: []int{10}, Uncovered: []int{11}},
		},
	}}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	require.NotNil(t, result.Evidence.Coverage)
	require.Len(t, result.Evidence.Coverage.ByFile, 2)
	assert.Equal(t, "src/main.go", result.Evidence.Coverage.ByFile[0].FilePath)
	assert.Equal(t, []int{1, 3, 5}, result.Evidence.Coverage.ByFile[0].Covered)
	assert.Equal(t, []int{2, 4}, result.Evidence.Coverage.ByFile[0].Uncovered)
	assert.Equal(t, "src/util.go", result.Evidence.Coverage.ByFile[1].FilePath)
}

func TestHandlerExecuteByFileEmptyWhenParserOmits(t *testing.T) {
	parser := &fakeParser{parsed: ingest.Parsed{
		TotalLines:   100,
		CoveredLines: 80,
		ByLanguage:   []coverage.LanguageStats{validLanguageCoverage(t, "go", 100, 80)},
	}}
	handler := ingest.NewHandler(map[string]ingest.Parser{"lcov": parser})

	cmd := mustCommand(t, []byte("TN:\n"), "lcov", "go-test")
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	require.NotNil(t, result.Evidence.Coverage)
	assert.Empty(t, result.Evidence.Coverage.ByFile)
}

func TestNewHandlerRejectsEmptyParsers(t *testing.T) {
	assert.Panics(t, func() {
		ingest.NewHandler(nil)
	})
	assert.Panics(t, func() {
		ingest.NewHandler(map[string]ingest.Parser{})
	})
}

func mustCommand(t *testing.T, data []byte, format, source string) ingest.Command {
	t.Helper()
	cmd, err := ingest.NewCommand(data, format, source)
	require.NoError(t, err)
	return cmd
}

func validLanguageCoverage(t *testing.T, langName string, total, covered int) coverage.LanguageStats {
	t.Helper()
	lang, err := coverage.NewLanguage(langName)
	require.NoError(t, err)
	lc, err := coverage.NewLanguageStats(lang, total, covered)
	require.NoError(t, err)
	return lc
}
