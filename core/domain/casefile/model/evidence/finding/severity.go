package finding

import "fmt"

type Severity struct {
	value string
}

var (
	SeverityError   = Severity{value: "error"}
	SeverityWarning = Severity{value: "warning"}
	SeverityNote    = Severity{value: "note"}
)

var validSeverities = map[string]Severity{
	"error":   SeverityError,
	"warning": SeverityWarning,
	"note":    SeverityNote,
}

func NewSeverity(s string) (Severity, error) {
	sev, ok := validSeverities[s]
	if !ok {
		return Severity{}, fmt.Errorf("%w: %q", ErrInvalidSeverity, s)
	}
	return sev, nil
}

func (s Severity) String() string {
	return s.value
}

func (s Severity) Equal(other Severity) bool {
	return s.value == other.value
}
