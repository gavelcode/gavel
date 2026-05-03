package verdict

import "fmt"

type Outcome struct {
	value string
}

var (
	OutcomePass = Outcome{value: "pass"}
	OutcomeFail = Outcome{value: "fail"}
)

var validOutcomes = map[string]Outcome{
	"pass": OutcomePass,
	"fail": OutcomeFail,
}

func NewOutcome(s string) (Outcome, error) {
	o, ok := validOutcomes[s]
	if !ok {
		return Outcome{}, fmt.Errorf("%w: invalid outcome %q", ErrInvalidVerdict, s)
	}
	return o, nil
}

func (o Outcome) String() string {
	return o.value
}

func (o Outcome) Equal(other Outcome) bool {
	return o.value == other.value
}

func (o Outcome) ShouldRecordAsBaseline() bool {
	return o.Equal(OutcomePass)
}
