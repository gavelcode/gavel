package failure

import "errors"

type Kind int

const (
	Internal Kind = iota

	Validation

	NotFound

	Conflict
)

type sentinel struct {
	msg  string
	kind Kind
}

func (s *sentinel) Error() string { return s.msg }

func (s *sentinel) Kind() Kind { return s.kind }

func New(msg string, kind Kind) error {
	return &sentinel{msg: msg, kind: kind}
}

func Of(err error) Kind {
	if err == nil {
		return Internal
	}
	var k interface{ Kind() Kind }
	if errors.As(err, &k) {
		return k.Kind()
	}
	return Internal
}
