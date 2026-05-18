package sarif

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

const (
	unknownFilePath = "unknown"

	levelError   = "error"
	levelWarning = "warning"
	levelNote    = "note"
	levelNone    = "none"
)

type SourceReader interface {
	ReadLine(filePath string, line int) (string, error)
}

type ParserOption func(*Parser)

func WithSourceReader(r SourceReader) ParserOption {
	return func(p *Parser) {
		p.source = r
	}
}

type Parser struct {
	source SourceReader
}

var _ ingestfindings.Parser = (*Parser)(nil)

func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Parser) Parse(_ context.Context, data []byte) ([]ingestfindings.Parsed, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeSARIF, err)
	}

	var out []ingestfindings.Parsed
	seen := make(map[string]int)
	for _, run := range doc.Runs {
		toolName := run.Tool.Driver.Name
		ruleLevels := indexRuleLevels(run.Tool.Driver.Rules)
		for i, result := range run.Results {
			parsed, err := toParsed(toolName, ruleLevels, result, p.source)
			if err != nil {
				return nil, fmt.Errorf("%w: result %d: %w", ErrInvalidResult, i, err)
			}
			fingerprint := parsed.FingerprintID.Value()
			if n := seen[fingerprint]; n > 0 {
				deduped, fpErr := finding.NewFingerprintID(fmt.Sprintf("%s:%d", fingerprint, n))
				if fpErr == nil {
					parsed.FingerprintID = deduped
				}
			}
			seen[fingerprint]++
			out = append(out, parsed)
		}
	}
	return out, nil
}

func indexRuleLevels(rules []ruleDescriptor) map[string]string {
	idx := make(map[string]string, len(rules))
	for _, r := range rules {
		if r.ID == "" {
			continue
		}
		idx[r.ID] = r.DefaultConfiguration.Level
	}
	return idx
}

func toParsed(toolName string, ruleLevels map[string]string, result result, src SourceReader) (ingestfindings.Parsed, error) {
	if strings.TrimSpace(result.RuleID) == "" {
		return ingestfindings.Parsed{}, fmt.Errorf("%w: missing ruleId", ErrInvalidResult)
	}

	level := result.Level
	if level == "" {
		level = ruleLevels[result.RuleID]
	}
	severity := mapLevel(level)

	filePath, line := extractLocation(result)
	fpValue := extractFingerprintValue(result, toolName, filePath, line, src)
	fingerprint, err := finding.NewFingerprintID(fpValue)
	if err != nil {
		return ingestfindings.Parsed{}, fmt.Errorf("fingerprint: %w", err)
	}

	return ingestfindings.Parsed{
		RuleID:        result.RuleID,
		Severity:      severity,
		FilePath:      filePath,
		Line:          line,
		Message:       result.Message.Text,
		FingerprintID: fingerprint,
	}, nil
}

func mapLevel(level string) finding.Severity {
	switch level {
	case levelError:
		return finding.SeverityError
	case levelNote, levelNone:
		return finding.SeverityNote
	case levelWarning, "":
		return finding.SeverityWarning
	default:
		return finding.SeverityWarning
	}
}

func extractLocation(r result) (string, int) {
	if len(r.Locations) == 0 {
		return unknownFilePath, 0
	}
	loc := r.Locations[0]
	filePath := loc.PhysicalLocation.ArtifactLocation.URI
	if filePath == "" {
		filePath = unknownFilePath
	}
	filePath = normalizePath(filePath)
	return filePath, loc.PhysicalLocation.Region.StartLine
}

const (
	fileURIPrefix  = "file://"
	execrootMarker = "execroot/_main/"
)

func normalizePath(path string) string {
	path = strings.TrimPrefix(path, fileURIPrefix)
	if idx := strings.Index(path, execrootMarker); idx != -1 {
		path = path[idx+len(execrootMarker):]
	}
	return path
}

func extractFingerprintValue(run result, toolName, filePath string, line int, src SourceReader) string {
	if v := firstNonEmptyByKey(run.Fingerprints); v != "" {
		return v
	}
	if v := firstNonEmptyByKey(run.PartialFingerprints); v != "" {
		return v
	}
	if src != nil {
		if content, err := src.ReadLine(filePath, line); err == nil && content != "" {
			input := fmt.Sprintf("%s:%s:%s:%s", toolName, run.RuleID, filePath, content)
			hash := sha256.Sum256([]byte(input))
			return fmt.Sprintf("%x", hash[:16])
		}
	}
	input := fmt.Sprintf("%s:%s:%s:%d", toolName, run.RuleID, filePath, line)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash[:16])
}

func firstNonEmptyByKey(message map[string]string) string {
	if len(message) == 0 {
		return ""
	}
	keys := make([]string, 0, len(message))
	for k := range message {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := message[k]; v != "" {
			return v
		}
	}
	return ""
}
