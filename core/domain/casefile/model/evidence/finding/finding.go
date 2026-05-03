package finding

import (
	"fmt"
	"strings"
)

type Finding struct {
	id       FingerprintID
	tool     string
	ruleID   string
	severity Severity
	filePath string
	line     int
	message  string
}

func NewFinding(tool, ruleID string, severity Severity, filePath string, line int, message string, fingerprintID FingerprintID) (Finding, error) {
	if strings.TrimSpace(tool) == "" {
		return Finding{}, fmt.Errorf("%w: tool must not be empty", ErrInvalidFinding)
	}
	if strings.TrimSpace(ruleID) == "" {
		return Finding{}, fmt.Errorf("%w: ruleID must not be empty", ErrInvalidFinding)
	}
	if strings.TrimSpace(filePath) == "" {
		return Finding{}, fmt.Errorf("%w: filePath must not be empty", ErrInvalidFinding)
	}
	if line < 0 {
		return Finding{}, fmt.Errorf("%w: line must be >= 0", ErrInvalidFinding)
	}
	return Finding{
		id:       fingerprintID,
		tool:     tool,
		ruleID:   ruleID,
		severity: severity,
		filePath: filePath,
		line:     line,
		message:  message,
	}, nil
}

func (f Finding) ID() FingerprintID  { return f.id }
func (f Finding) Tool() string       { return f.tool }
func (f Finding) RuleID() string     { return f.ruleID }
func (f Finding) Severity() Severity { return f.severity }
func (f Finding) FilePath() string   { return f.filePath }
func (f Finding) Line() int          { return f.line }
func (f Finding) Message() string    { return f.message }

func (f Finding) Equal(other Finding) bool {
	return f.id.Equal(other.id)
}
