package judge

import "time"

type VerdictView struct {
	Outcome     string
	Rulings     []RulingView
	EvaluatedAt time.Time
}
