package listfindings

type FindingView struct {
	Tool          string
	RuleID        string
	Severity      string
	FilePath      string
	Line          int
	Message       string
	FingerprintID string
	Status        string
	Source        string
	CommitSHA     string
	ProjectKey    string
	CaseFileID    string
}
