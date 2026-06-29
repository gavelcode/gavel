package toolexecution

import (
	"fmt"
	"strings"
)

// Failure records that one analyzer could not run to completion: which tool and
// why. It is the unit of tool-execution evidence — the verdict cannot be trusted
// for a target whose analysis was incomplete.
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
