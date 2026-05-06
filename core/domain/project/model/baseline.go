package model

import (
	"errors"
	"sort"
)

var ErrEmptyFilePath = errors.New("file path must not be empty")

type FileCoverageEntry struct {
	filePath  string
	covered   []int
	uncovered []int
}

func NewFileCoverageEntry(filePath string, covered, uncovered []int) (FileCoverageEntry, error) {
	if filePath == "" {
		return FileCoverageEntry{}, ErrEmptyFilePath
	}
	return FileCoverageEntry{
		filePath:  filePath,
		covered:   copyInts(covered),
		uncovered: copyInts(uncovered),
	}, nil
}

func (e FileCoverageEntry) FilePath() string { return e.filePath }

func (e FileCoverageEntry) Covered() []int { return copyInts(e.covered) }

func (e FileCoverageEntry) Uncovered() []int { return copyInts(e.uncovered) }

func copyInts(src []int) []int {
	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

type Baseline struct {
	fingerprints    []string
	archIDs         []string
	coveragePercent *float64
	fileCoverage    []FileCoverageEntry
}

func NewBaseline(fingerprints, archIDs []string, coveragePercent *float64, fileCoverage []FileCoverageEntry) Baseline {
	var copied *float64
	if coveragePercent != nil {
		v := *coveragePercent
		copied = &v
	}
	return Baseline{
		fingerprints:    sortedUnique(fingerprints),
		archIDs:         sortedUnique(archIDs),
		coveragePercent: copied,
		fileCoverage:    copyFileCoverage(fileCoverage),
	}
}

func (b Baseline) Fingerprints() []string {
	if b.fingerprints == nil {
		return nil
	}
	copied := make([]string, len(b.fingerprints))
	copy(copied, b.fingerprints)
	return copied
}

func (b Baseline) ArchIDs() []string {
	if b.archIDs == nil {
		return nil
	}
	copied := make([]string, len(b.archIDs))
	copy(copied, b.archIDs)
	return copied
}

func (b Baseline) CoveragePercent() *float64 {
	if b.coveragePercent == nil {
		return nil
	}
	v := *b.coveragePercent
	return &v
}

func (b Baseline) FileCoverage() []FileCoverageEntry {
	return copyFileCoverage(b.fileCoverage)
}

func (b Baseline) HasPrevious() bool {
	return len(b.fingerprints) > 0 || len(b.archIDs) > 0 || b.coveragePercent != nil
}

func copyFileCoverage(src []FileCoverageEntry) []FileCoverageEntry {
	if src == nil {
		return nil
	}
	dst := make([]FileCoverageEntry, len(src))
	copy(dst, src)
	return dst
}

func (b Baseline) Ratchet(currentFingerprints, currentArchIDs []string) Baseline {
	return NewBaseline(
		intersect(b.fingerprints, currentFingerprints),
		intersect(b.archIDs, currentArchIDs),
		b.CoveragePercent(),
		b.FileCoverage(),
	)
}

func intersect(baseline, current []string) []string {
	currentSet := make(map[string]struct{}, len(current))
	for _, identifier := range current {
		currentSet[identifier] = struct{}{}
	}
	var result []string
	for _, identifier := range baseline {
		if _, ok := currentSet[identifier]; ok {
			result = append(result, identifier)
		}
	}
	return result
}

func sortedUnique(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	copied := make([]string, len(s))
	copy(copied, s)
	sort.Strings(copied)
	writeIndex := 0
	for readIndex := range copied {
		if readIndex == 0 || copied[readIndex] != copied[readIndex-1] {
			copied[writeIndex] = copied[readIndex]
			writeIndex++
		}
	}
	return copied[:writeIndex]
}
