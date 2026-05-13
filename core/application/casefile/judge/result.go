package judge

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	CaseFileID string
	Verdict    VerdictView
	Events     []event.Event
}
