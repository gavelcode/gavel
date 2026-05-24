package watch

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

type Event struct {
	Event        string   `json:"event"`
	Timestamp    string   `json:"ts"`
	Workspace    string   `json:"workspace,omitempty"`
	BazelVersion string   `json:"bazel_version,omitempty"`
	Files        []string `json:"files,omitempty"`
	Targets      []string `json:"targets,omitempty"`
	Target       string   `json:"target,omitempty"`
	Tool         string   `json:"tool,omitempty"`
	Rule         string   `json:"rule,omitempty"`
	Severity     string   `json:"severity,omitempty"`
	File         string   `json:"file,omitempty"`
	Line         int      `json:"line,omitempty"`
	Message      string   `json:"message,omitempty"`
	Fingerprint  string   `json:"fingerprint,omitempty"`
	Findings     int      `json:"findings,omitempty"`
	DurationMs   int64    `json:"duration_ms,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

type Emitter struct {
	mu  sync.Mutex
	enc *json.Encoder
	now func() time.Time
}

func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{enc: json.NewEncoder(w), now: time.Now}
}

func (e *Emitter) emit(ev Event) error {
	ev.Timestamp = e.now().UTC().Format(time.RFC3339Nano)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enc.Encode(ev)
}

func (e *Emitter) Started(workspace, bazelVersion string) error {
	return e.emit(Event{Event: "started", Workspace: workspace, BazelVersion: bazelVersion})
}

func (e *Emitter) Changed(files []string) error {
	return e.emit(Event{Event: "changed", Files: files})
}

func (e *Emitter) Affected(targets []string) error {
	return e.emit(Event{Event: "affected", Targets: targets})
}

func (e *Emitter) AnalysisStarted(target string) error {
	return e.emit(Event{Event: "analysis_started", Target: target})
}

func (e *Emitter) Finding(target, tool, rule, severity, file string, line int, message, fingerprint string) error {
	return e.emit(Event{
		Event: "finding", Target: target, Tool: tool, Rule: rule,
		Severity: severity, File: file, Line: line, Message: message, Fingerprint: fingerprint,
	})
}

func (e *Emitter) AnalysisDone(target string, findings int, duration time.Duration) error {
	return e.emit(Event{
		Event: "analysis_done", Target: target,
		Findings: findings, DurationMs: duration.Milliseconds(),
	})
}

func (e *Emitter) AnalysisFailed(target, reason string) error {
	return e.emit(Event{Event: "analysis_failed", Target: target, Reason: reason})
}

func (e *Emitter) Stopped(reason string) error {
	return e.emit(Event{Event: "stopped", Reason: reason})
}
