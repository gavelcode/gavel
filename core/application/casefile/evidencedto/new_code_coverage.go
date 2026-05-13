package evidencedto

import "github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"

type NewCodeCoverage struct {
	CoveredLines   int
	CoverableLines int
}

func fromDomainNewCodeCoverage(c coverage.PatchContent) NewCodeCoverage {
	return NewCodeCoverage{
		CoveredLines:   c.CoveredLines(),
		CoverableLines: c.CoverableLines(),
	}
}

func toDomainNewCodeCoverage(in NewCodeCoverage) (coverage.PatchContent, error) {
	return coverage.NewPatchContent(in.CoveredLines, in.CoverableLines)
}
