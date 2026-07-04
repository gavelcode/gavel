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

// Verdict is the read view of a cached judge result. Its fields are mapped from
// the write-side projectDTO, which owns the on-disk JSON contract (the json
// tags) — so a single tagged struct set defines the format and the read and
// write sides cannot drift.
type Verdict struct {
	Name            string
	Verdict         string
	CommitSHA       string
	Branch          string
	CoveragePercent *float64
	Findings        []VerdictFinding
	Violations      []VerdictViolation
	Delta           *VerdictDelta
}

type VerdictFinding struct {
	Tool          string
	RuleID        string
	Severity      string
	FilePath      string
	Line          int
	Message       string
	FingerprintID string
	Status        string
}

type VerdictViolation struct {
	Rule      string
	SourcePkg string
	TargetPkg string
	Message   string
	Status    string
}

type VerdictDelta struct {
	HasPrevious   bool
	NewCount      int
	FixedCount    int
	ExistingCount int
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
		var dto projectDTO
		if err := encjson.Unmarshal(data, &dto); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		verdicts = append(verdicts, toVerdict(dto))
	}

	if len(verdicts) == 0 {
		return nil, ErrNoResults
	}
	return verdicts, nil
}

func toVerdict(dto projectDTO) Verdict {
	verdict := Verdict{
		Name:            dto.Name,
		Verdict:         dto.Verdict,
		CommitSHA:       dto.CommitSHA,
		Branch:          dto.Branch,
		CoveragePercent: dto.CoveragePercent,
	}
	// Direct struct conversions (fields are identical): this also guards the
	// contract — if a write-side DTO field changes, the conversion stops
	// compiling, forcing the read view to be updated rather than drifting.
	for _, finding := range dto.Findings {
		verdict.Findings = append(verdict.Findings, VerdictFinding(finding))
	}
	for _, violation := range dto.Violations {
		verdict.Violations = append(verdict.Violations, VerdictViolation(violation))
	}
	if dto.Delta != nil {
		verdict.Delta = &VerdictDelta{
			HasPrevious:   dto.Delta.HasPrevious,
			NewCount:      dto.Delta.NewCount,
			FixedCount:    dto.Delta.FixedCount,
			ExistingCount: dto.Delta.ExistingCount,
		}
	}
	return verdict
}
