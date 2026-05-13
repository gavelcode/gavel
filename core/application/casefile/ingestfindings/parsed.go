package ingestfindings

import "github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"

type Parsed struct {
	RuleID        string
	Severity      finding.Severity
	FilePath      string
	Line          int
	Message       string
	FingerprintID finding.FingerprintID
}
