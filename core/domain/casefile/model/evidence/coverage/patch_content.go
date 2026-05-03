package coverage

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type PatchContent struct {
	coveredLines   int
	coverableLines int
}

func NewPatchContent(coveredLines, coverableLines int) (PatchContent, error) {
	if coveredLines < 0 {
		return PatchContent{}, fmt.Errorf("%w: coveredLines must be >= 0", ErrInvalidPatchContent)
	}
	if coverableLines < 0 {
		return PatchContent{}, fmt.Errorf("%w: coverableLines must be >= 0", ErrInvalidPatchContent)
	}
	if coveredLines > coverableLines {
		return PatchContent{}, fmt.Errorf("%w: coveredLines must be <= coverableLines", ErrInvalidPatchContent)
	}
	return PatchContent{
		coveredLines:   coveredLines,
		coverableLines: coverableLines,
	}, nil
}

func NewPatchContentFromChanges(changedLines map[string][]int, perLineCoverage map[string]map[int]int) (PatchContent, error) {
	var covered, coverable int
	for file, lines := range changedLines {
		fileCov, exists := perLineCoverage[file]
		if !exists {
			continue
		}
		for _, lineNum := range lines {
			hitCount, isCoverable := fileCov[lineNum]
			if !isCoverable {
				continue
			}
			coverable++
			if hitCount > 0 {
				covered++
			}
		}
	}
	return NewPatchContent(covered, coverable)
}

func (c PatchContent) Type() evidence.Type {
	return evidence.TypeCoverage
}

func (c PatchContent) Subtype() evidence.Subtype {
	return evidence.SubtypeNewCodeCoverage
}

func (c PatchContent) CoveredLines() int {
	return c.coveredLines
}

func (c PatchContent) CoverableLines() int {
	return c.coverableLines
}

func (c PatchContent) Percent() float64 {
	if c.coverableLines == 0 {
		return 0
	}
	return float64(c.coveredLines) / float64(c.coverableLines) * percentMultiplier
}

func (c PatchContent) Merge(other evidence.Content) (evidence.Content, error) {
	o, ok := other.(PatchContent)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge patch content with %T", ErrInvalidPatchContent, other)
	}
	return NewPatchContent(c.coveredLines+o.coveredLines, c.coverableLines+o.coverableLines)
}
