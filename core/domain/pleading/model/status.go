package model

import "fmt"

type Status struct {
	value string
}

var (
	StatusOpen   = Status{value: "open"}
	StatusMerged = Status{value: "merged"}
	StatusClosed = Status{value: "closed"}
)

var validStatuses = map[string]Status{
	"open":   StatusOpen,
	"merged": StatusMerged,
	"closed": StatusClosed,
}

func NewStatus(s string) (Status, error) {
	status, ok := validStatuses[s]
	if !ok {
		return Status{}, fmt.Errorf("%w: %q", ErrInvalidStatus, s)
	}
	return status, nil
}

func (s Status) String() string {
	return s.value
}

func (s Status) Equal(other Status) bool {
	return s.value == other.value
}

func (s Status) IsTerminal() bool {
	return s.Equal(StatusMerged) || s.Equal(StatusClosed)
}
