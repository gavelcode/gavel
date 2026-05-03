package architecture

import (
	"fmt"
	"strings"
)

type Violation struct {
	rule      string
	sourcePkg string
	targetPkg string
	message   string
}

func NewViolation(rule, sourcePkg, targetPkg, message string) (Violation, error) {
	if strings.TrimSpace(rule) == "" {
		return Violation{}, fmt.Errorf("%w: rule must not be empty", ErrInvalidViolation)
	}
	if strings.TrimSpace(sourcePkg) == "" {
		return Violation{}, fmt.Errorf("%w: sourcePkg must not be empty", ErrInvalidViolation)
	}
	if strings.TrimSpace(message) == "" {
		return Violation{}, fmt.Errorf("%w: message must not be empty", ErrInvalidViolation)
	}

	return Violation{
		rule:      rule,
		sourcePkg: sourcePkg,
		targetPkg: targetPkg,
		message:   message,
	}, nil
}

func (av Violation) Rule() string      { return av.rule }
func (av Violation) SourcePkg() string { return av.sourcePkg }
func (av Violation) TargetPkg() string { return av.targetPkg }
func (av Violation) Message() string   { return av.message }
