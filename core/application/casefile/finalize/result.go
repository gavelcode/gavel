package finalize

import (
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/shared/event"
)

type Result struct {
	CaseFileID string
	Verdict    judge.VerdictView
	Counters   Counters
	Delta      Delta
	Events     []event.Event
}
