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

var ErrNoResults = errors.New("no judge results found under .gavel/results; run `gavel judge` first")

type Verdict struct {
	Name            string
	Verdict         string
	CommitSHA       string
	Branch          string
	StartedAt       string
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

func Load(workspace string) ([]Verdict, []string, error) {
	dir := filepath.Join(workspace, gavelDir, resultsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrNoResults
		}
		return nil, nil, fmt.Errorf("read results dir: %w", err)
	}

	var verdicts []Verdict
	var skipped []string
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
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		var dto projectDTO
		if err := encjson.Unmarshal(data, &dto); err != nil {
			skipped = append(skipped, path)
			continue
		}
		verdicts = append(verdicts, toVerdict(dto))
	}

	if len(verdicts) == 0 {
		return nil, skipped, ErrNoResults
	}
	return verdicts, skipped, nil
}

func toVerdict(dto projectDTO) Verdict {
	verdict := Verdict{
		Name:            dto.Name,
		Verdict:         dto.Verdict,
		CommitSHA:       dto.CommitSHA,
		Branch:          dto.Branch,
		StartedAt:       dto.StartedAt,
		CoveragePercent: dto.CoveragePercent,
	}
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
