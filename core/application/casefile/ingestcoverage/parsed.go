package ingestcoverage

import (
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Parsed struct {
	TotalLines   int
	CoveredLines int
	ByLanguage   []coverage.LanguageStats
	ByFile       []evidencedto.FileCoverage
}
