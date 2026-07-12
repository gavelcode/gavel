package toolexecution

import (
	"fmt"
	"strings"
)

type Failure struct {
	tool     string
	reason   string
	degraded bool
}

func NewFailure(tool, reason string) (Failure, error) {
	return newFailure(tool, reason, false)
}

// NewDegradedFailure records a tool that ran but could not analyze completely
// (e.g. Error Prone whose javac hit unresolved symbols because the target uses
// annotation processors). The analysis is partial, not broken, so it must not
// fail the verdict — only surface honestly that coverage was incomplete.
func NewDegradedFailure(tool, reason string) (Failure, error) {
	return newFailure(tool, reason, true)
}

func newFailure(tool, reason string, degraded bool) (Failure, error) {
	if strings.TrimSpace(tool) == "" {
		return Failure{}, fmt.Errorf("%w: tool must not be empty", ErrInvalidFailure)
	}
	if strings.TrimSpace(reason) == "" {
		return Failure{}, fmt.Errorf("%w: reason must not be empty", ErrInvalidFailure)
	}

	return Failure{
		tool:     tool,
		reason:   reason,
		degraded: degraded,
	}, nil
}

func (f Failure) Tool() string   { return f.tool }
func (f Failure) Reason() string { return f.reason }
func (f Failure) Degraded() bool { return f.degraded }
