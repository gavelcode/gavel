package toolexecution

import (
	"fmt"
	"strings"
)

type Failure struct {
	tool   string
	reason string
}

func NewFailure(tool, reason string) (Failure, error) {
	if strings.TrimSpace(tool) == "" {
		return Failure{}, fmt.Errorf("%w: tool must not be empty", ErrInvalidFailure)
	}
	if strings.TrimSpace(reason) == "" {
		return Failure{}, fmt.Errorf("%w: reason must not be empty", ErrInvalidFailure)
	}

	return Failure{
		tool:   tool,
		reason: reason,
	}, nil
}

func (f Failure) Tool() string   { return f.tool }
func (f Failure) Reason() string { return f.reason }
