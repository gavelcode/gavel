package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Coverage struct {
	TotalLines   int
	CoveredLines int
	ByLanguage   []LanguageStats
	ByFile       []FileCoverage
}

func fromDomainCoverage(content coverage.Content) Coverage {
	langs := content.ByLanguage()
	out := make([]LanguageStats, 0, len(langs))
	for _, l := range langs {
		out = append(out, LanguageStats{
			Language:     l.Language().String(),
			TotalLines:   l.TotalLines(),
			CoveredLines: l.CoveredLines(),
		})
	}
	return Coverage{
		TotalLines:   content.TotalLines(),
		CoveredLines: content.CoveredLines(),
		ByLanguage:   out,
	}
}

func toDomainCoverage(input Coverage) (coverage.Content, error) {
	langs := make([]coverage.LanguageStats, 0, len(input.ByLanguage))
	for index, l := range input.ByLanguage {
		lang, err := coverage.NewLanguage(l.Language)
		if err != nil {
			return coverage.Content{}, fmt.Errorf("byLanguage[%d] language: %w", index, err)
		}
		lc, err := coverage.NewLanguageStats(lang, l.TotalLines, l.CoveredLines)
		if err != nil {
			return coverage.Content{}, fmt.Errorf("byLanguage[%d]: %w", index, err)
		}
		langs = append(langs, lc)
	}
	return coverage.NewContent(input.TotalLines, input.CoveredLines, langs)
}
