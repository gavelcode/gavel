package ingestfindings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func TestHandlerExecuteSuccessfulIngest(t *testing.T) {
	parser := &fakeParser{parsed: []ingestfindings.Parsed{validParsed(t)}}
	handler := ingestfindings.NewHandler(map[string]ingestfindings.Parser{"sarif": parser})

	cmd := mustCommand(t, []byte("{}"), "sarif", "spotbugs", evidence.SubtypeCodeQuality.String())
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, parser.calls)
	assert.Equal(t, evidence.SubtypeCodeQuality.String(), result.Evidence.Subtype)
	assert.Equal(t, "spotbugs", result.Evidence.Source)

	require.Len(t, result.Evidence.Findings, 1)
	assert.Equal(t, "rule1", result.Evidence.Findings[0].RuleID)
	assert.Equal(t, "spotbugs", result.Evidence.Findings[0].Tool)
}

func TestHandlerExecuteUnknownFormat(t *testing.T) {
	parser := &fakeParser{parsed: nil}
	handler := ingestfindings.NewHandler(map[string]ingestfindings.Parser{"sarif": parser})

	cmd := mustCommand(t, []byte("{}"), "spotbugs-xml", "spotbugs", evidence.SubtypeCodeQuality.String())
	_, err := handler.Execute(context.Background(), cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, ingestfindings.ErrUnknownFormat)
	assert.Equal(t, 0, parser.calls)
}

func TestHandlerExecuteParserError(t *testing.T) {
	boom := errors.New("boom")
	parser := &fakeParser{err: boom}
	handler := ingestfindings.NewHandler(map[string]ingestfindings.Parser{"sarif": parser})

	cmd := mustCommand(t, []byte("{}"), "sarif", "spotbugs", evidence.SubtypeCodeQuality.String())
	_, err := handler.Execute(context.Background(), cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, ingestfindings.ErrParseFailed)
	assert.ErrorIs(t, err, boom)
}

func TestHandlerExecuteEmptyFindingsListProducesValidEvidence(t *testing.T) {
	parser := &fakeParser{parsed: nil}
	handler := ingestfindings.NewHandler(map[string]ingestfindings.Parser{"sarif": parser})

	cmd := mustCommand(t, []byte("{}"), "sarif", "spotbugs", evidence.SubtypeCodeQuality.String())
	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Empty(t, result.Evidence.Findings)
}

func TestHandlerExecuteDispatchesByFormat(t *testing.T) {
	sarifParser := &fakeParser{parsed: []ingestfindings.Parsed{validParsed(t)}}
	xmlParser := &fakeParser{parsed: nil}
	handler := ingestfindings.NewHandler(map[string]ingestfindings.Parser{
		"sarif":        sarifParser,
		"spotbugs-xml": xmlParser,
	})

	cmd := mustCommand(t, []byte("{}"), "sarif", "spotbugs", evidence.SubtypeCodeQuality.String())
	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, sarifParser.calls)
	assert.Equal(t, 0, xmlParser.calls)
}

func TestNewHandlerRejectsEmptyParsers(t *testing.T) {
	assert.Panics(t, func() {
		ingestfindings.NewHandler(nil)
	})
	assert.Panics(t, func() {
		ingestfindings.NewHandler(map[string]ingestfindings.Parser{})
	})
}

func mustCommand(t *testing.T, data []byte, format, source, subtype string) ingestfindings.Command {
	t.Helper()
	cmd, err := ingestfindings.NewCommand(data, format, source, subtype)
	require.NoError(t, err)
	return cmd
}

func validParsed(t *testing.T) ingestfindings.Parsed {
	t.Helper()
	fingerprint, err := finding.NewFingerprintID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	require.NoError(t, err)
	return ingestfindings.Parsed{
		RuleID:        "rule1",
		Severity:      finding.SeverityError,
		FilePath:      "file.go",
		Line:          10,
		Message:       "issue",
		FingerprintID: fingerprint,
	}
}
