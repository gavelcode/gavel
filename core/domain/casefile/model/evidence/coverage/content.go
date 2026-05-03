package coverage

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

const percentMultiplier = 100

type Content struct {
	totalLines   int
	coveredLines int
	byLanguage   []LanguageStats
}

func NewContent(totalLines, coveredLines int, byLanguage []LanguageStats) (Content, error) {
	if totalLines < 0 {
		return Content{}, fmt.Errorf("%w: totalLines must be >= 0", ErrInvalidContent)
	}
	if coveredLines < 0 {
		return Content{}, fmt.Errorf("%w: coveredLines must be >= 0", ErrInvalidContent)
	}
	if coveredLines > totalLines {
		return Content{}, fmt.Errorf("%w: coveredLines must be <= totalLines", ErrInvalidContent)
	}

	copied := make([]LanguageStats, len(byLanguage))
	copy(copied, byLanguage)

	return Content{
		totalLines:   totalLines,
		coveredLines: coveredLines,
		byLanguage:   copied,
	}, nil
}

func (cc Content) Type() evidence.Type {
	return evidence.TypeCoverage
}

func (cc Content) Subtype() evidence.Subtype {
	return evidence.SubtypeCoverage
}

func (cc Content) TotalLines() int {
	return cc.totalLines
}

func (cc Content) CoveredLines() int {
	return cc.coveredLines
}

func (cc Content) Percent() float64 {
	if cc.totalLines == 0 {
		return 0
	}
	return float64(cc.coveredLines) / float64(cc.totalLines) * percentMultiplier
}

func (cc Content) ByLanguage() []LanguageStats {
	copied := make([]LanguageStats, len(cc.byLanguage))
	copy(copied, cc.byLanguage)
	return copied
}

func (cc Content) Merge(other evidence.Content) (evidence.Content, error) {
	otherCoverage, ok := other.(Content)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge coverage content with %T", ErrInvalidContent, other)
	}
	merged := make([]LanguageStats, 0, len(cc.byLanguage)+len(otherCoverage.byLanguage))
	merged = append(merged, cc.byLanguage...)
	merged = append(merged, otherCoverage.byLanguage...)
	return NewContent(cc.totalLines+otherCoverage.totalLines, cc.coveredLines+otherCoverage.coveredLines, merged)
}
