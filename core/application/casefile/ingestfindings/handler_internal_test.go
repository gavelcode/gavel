package ingestfindings

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type stubParser struct {
	result []Parsed
	err    error
}

func (s *stubParser) Parse(_ context.Context, _ []byte) ([]Parsed, error) {
	return s.result, s.err
}

func TestExecuteToDomainFindingsError(t *testing.T) {
	fpID, _ := finding.NewFingerprintID("fp-1")
	handler := &Handler{
		parsers: map[string]Parser{
			"sarif": &stubParser{result: []Parsed{{
				RuleID: "", Severity: finding.SeverityWarning, FilePath: "f.go", Line: 1, Message: "m", FingerprintID: fpID,
			}}},
		},
		now: func() time.Time { return time.Now().UTC() },
	}

	cmd := Command{data: []byte("x"), format: "sarif", source: "test", subtype: evidence.SubtypeCodeQuality}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build findings")
}

func TestExecuteNewContentError(t *testing.T) {
	sev, _ := finding.NewSeverity("warning")
	fpID, _ := finding.NewFingerprintID("fp-1")
	handler := &Handler{
		parsers: map[string]Parser{
			"sarif": &stubParser{result: []Parsed{{
				RuleID: "r", Severity: sev, FilePath: "f.go", Line: 1, Message: "m", FingerprintID: fpID,
			}}},
		},
		now: func() time.Time { return time.Now().UTC() },
	}

	cmd := Command{data: []byte("x"), format: "sarif", source: "test", subtype: evidence.SubtypeCoverage}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build findings content")
}

func TestExecuteNewEvidenceError(t *testing.T) {
	sev, _ := finding.NewSeverity("warning")
	fpID, _ := finding.NewFingerprintID("fp-1")
	handler := &Handler{
		parsers: map[string]Parser{
			"sarif": &stubParser{result: []Parsed{{
				RuleID: "r", Severity: sev, FilePath: "f.go", Line: 1, Message: "m", FingerprintID: fpID,
			}}},
		},
		now: func() time.Time { return time.Time{} },
	}

	cmd := Command{data: []byte("x"), format: "sarif", source: "test", subtype: evidence.SubtypeCodeQuality}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build evidence")
}
