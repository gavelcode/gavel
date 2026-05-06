package qualitygate

type Outcome struct {
	passed bool
	detail string
}

func NewOutcome(passed bool, detail string) Outcome {
	return Outcome{passed: passed, detail: detail}
}

func (r Outcome) Passed() bool {
	return r.passed
}

func (r Outcome) Detail() string {
	return r.detail
}
