package submit

import (
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/shared/event"
)

type Result struct {
	CaseFileID string
	Verdict    judge.VerdictView
	Counters   finalize.Counters
	Delta      finalize.Delta
	Events     []event.Event
}
