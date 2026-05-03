package coverage

import "fmt"

type LanguageStats struct {
	language     Language
	totalLines   int
	coveredLines int
}

func NewLanguageStats(language Language, totalLines, coveredLines int) (LanguageStats, error) {
	if totalLines < 0 {
		return LanguageStats{}, fmt.Errorf("%w: totalLines must be >= 0", ErrInvalidLanguageStats)
	}
	if coveredLines < 0 {
		return LanguageStats{}, fmt.Errorf("%w: coveredLines must be >= 0", ErrInvalidLanguageStats)
	}
	if coveredLines > totalLines {
		return LanguageStats{}, fmt.Errorf("%w: coveredLines must be <= totalLines", ErrInvalidLanguageStats)
	}

	return LanguageStats{
		language:     language,
		totalLines:   totalLines,
		coveredLines: coveredLines,
	}, nil
}

func (lc LanguageStats) Language() Language { return lc.language }
func (lc LanguageStats) TotalLines() int    { return lc.totalLines }
func (lc LanguageStats) CoveredLines() int  { return lc.coveredLines }

func (lc LanguageStats) Percent() float64 {
	if lc.totalLines == 0 {
		return 0
	}
	return float64(lc.coveredLines) / float64(lc.totalLines) * percentMultiplier
}
