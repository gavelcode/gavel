package analyzetarget

import "time"

type Finding struct {
	Tool        string
	RuleID      string
	Severity    string
	FilePath    string
	Line        int
	Message     string
	Fingerprint string
}

type Result struct {
	Target   string
	Findings []Finding
	Duration time.Duration
}
