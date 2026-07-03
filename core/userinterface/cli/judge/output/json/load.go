package json

import (
	encjson "encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	gavelDir    = ".gavel"
	resultsDir  = "results"
	verdictFile = "verdict.json"
)

// ErrNoResults signals that no judge results were found under the workspace:
// `gavel judge` has not run, so there is nothing for a consumer to deliver.
var ErrNoResults = errors.New("no judge results found under .gavel/results; run `gavel judge` first")

// Verdict is one project's judged result, read back from
// .gavel/results/<project>/verdict.json. It is the read view of the same
// on-disk contract WriteCache produces, exposing what a downstream consumer
// (the report command) delivers to an external sink.
type Verdict struct {
	Name            string             `json:"name"`
	Verdict         string             `json:"verdict"`
	CommitSHA       string             `json:"commit_sha"`
	Branch          string             `json:"branch"`
	CoveragePercent *float64           `json:"coverage_percent"`
	Findings        []VerdictFinding   `json:"findings"`
	Violations      []VerdictViolation `json:"violations"`
	Delta           *VerdictDelta      `json:"delta"`
}

type VerdictFinding struct {
	Tool          string `json:"tool"`
	RuleID        string `json:"rule_id"`
	Severity      string `json:"severity"`
	FilePath      string `json:"file_path"`
	Line          int    `json:"line"`
	Message       string `json:"message"`
	FingerprintID string `json:"fingerprint"`
	Status        string `json:"status"`
}

type VerdictViolation struct {
	Rule      string `json:"rule"`
	SourcePkg string `json:"source_pkg"`
	TargetPkg string `json:"target_pkg"`
	Message   string `json:"message"`
	Status    string `json:"status"`
}

type VerdictDelta struct {
	HasPrevious   bool `json:"has_previous"`
	NewCount      int  `json:"new_count"`
	FixedCount    int  `json:"fixed_count"`
	ExistingCount int  `json:"existing_count"`
}

// Load reads every .gavel/results/<project>/verdict.json cached by WriteCache
// under the workspace. It returns ErrNoResults when the results directory is
// absent or holds no verdict files.
func Load(workspace string) ([]Verdict, error) {
	dir := filepath.Join(workspace, gavelDir, resultsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoResults
		}
		return nil, fmt.Errorf("read results dir: %w", err)
	}

	var verdicts []Verdict
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name(), verdictFile)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var verdict Verdict
		if err := encjson.Unmarshal(data, &verdict); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		verdicts = append(verdicts, verdict)
	}

	if len(verdicts) == 0 {
		return nil, ErrNoResults
	}
	return verdicts, nil
}
