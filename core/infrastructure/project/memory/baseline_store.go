package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/usegavel/gavel/core/domain/project/model"
)

const (
	dirPermission  = 0o755
	filePermission = 0o644
)

type BaselineStore struct {
	workspace string
}

func NewBaselineStore(workspace string) *BaselineStore {
	return &BaselineStore{workspace: workspace}
}

func (s *BaselineStore) Load(projectName string) model.Baseline {
	dir := filepath.Join(s.workspace, ".gavel", "baseline", projectName)
	fps := readLines(filepath.Join(dir, "findings"))
	archIDs := readLines(filepath.Join(dir, "architecture"))
	cov, fileCov := readCoverageJSON(filepath.Join(dir, "coverage.json"))
	if cov == nil {
		cov = readCoverage(filepath.Join(dir, "coverage"))
	}
	if fps == nil && archIDs == nil && cov == nil {
		return model.Baseline{}
	}
	return model.NewBaseline(fps, archIDs, cov, fileCov)
}

func (s *BaselineStore) Save(projectName string, baseline model.Baseline) error {
	dir := filepath.Join(s.workspace, ".gavel", "baseline", projectName)
	if err := os.MkdirAll(dir, dirPermission); err != nil {
		return err
	}
	if err := writeOrClearLines(filepath.Join(dir, "findings"), baseline.Fingerprints()); err != nil {
		return err
	}
	if err := writeOrClearLines(filepath.Join(dir, "architecture"), baseline.ArchIDs()); err != nil {
		return err
	}
	if err := writeOrClearCoverage(filepath.Join(dir, "coverage"), baseline.CoveragePercent()); err != nil {
		return err
	}
	if err := writeOrClearFileCoverage(filepath.Join(dir, "coverage.json"), baseline); err != nil {
		return err
	}
	return nil
}

func writeOrClearLines(path string, values []string) error {
	if len(values) == 0 {
		return removeIfExists(path)
	}
	sorted := make([]string, len(values))
	copy(sorted, values)
	sort.Strings(sorted)
	content := strings.Join(sorted, "\n") + "\n"
	return os.WriteFile(path, []byte(content), filePermission)
}

func writeOrClearCoverage(path string, coverage *float64) error {
	if coverage == nil {
		return removeIfExists(path)
	}
	content := fmt.Sprintf("%g\n", *coverage)
	return os.WriteFile(path, []byte(content), filePermission)
}

func writeOrClearFileCoverage(path string, baseline model.Baseline) error {
	entries := baseline.FileCoverage()
	if len(entries) == 0 {
		return removeIfExists(path)
	}
	pct := 0.0
	if cp := baseline.CoveragePercent(); cp != nil {
		pct = *cp
	}
	return writeFileCoverageJSON(path, pct, entries)
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

type fileCoverageJSON struct {
	FilePath  string `json:"file_path"`
	Covered   []int  `json:"covered"`
	Uncovered []int  `json:"uncovered,omitempty"`
}

type coverageJSON struct {
	Percent float64            `json:"percent"`
	ByFile  []fileCoverageJSON `json:"by_file"`
}

func writeFileCoverageJSON(path string, percent float64, entries []model.FileCoverageEntry) error {
	byFile := make([]fileCoverageJSON, 0, len(entries))
	for _, e := range entries {
		byFile = append(byFile, fileCoverageJSON{
			FilePath:  e.FilePath(),
			Covered:   e.Covered(),
			Uncovered: e.Uncovered(),
		})
	}
	data, err := json.Marshal(coverageJSON{Percent: percent, ByFile: byFile})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, filePermission)
}

func readCoverageJSON(path string) (*float64, []model.FileCoverageEntry) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	var covJSON coverageJSON
	if err := json.Unmarshal(data, &covJSON); err != nil {
		return nil, nil
	}
	entries := make([]model.FileCoverageEntry, 0, len(covJSON.ByFile))
	for _, f := range covJSON.ByFile {
		entry, err := model.NewFileCoverageEntry(f.FilePath, f.Covered, f.Uncovered)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	pct := covJSON.Percent
	if len(entries) == 0 {
		return &pct, nil
	}
	return &pct, entries
}

func readCoverage(path string) *float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil
	}
	v, err := strconv.ParseFloat(content, 64)
	if err != nil {
		return nil
	}
	return &v
}

func readLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}
