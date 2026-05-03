package tracking

import "github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"

type Result struct {
	newFindings      []finding.Finding
	existingFindings []finding.Finding
	resolvedCount    int
}

func NewResult(newFindings, existingFindings []finding.Finding, resolvedCount int) Result {
	return Result{
		newFindings:      copyFindings(newFindings),
		existingFindings: copyFindings(existingFindings),
		resolvedCount:    resolvedCount,
	}
}

func (tr Result) NewFindings() []finding.Finding {
	return copyFindings(tr.newFindings)
}

func (tr Result) ExistingFindings() []finding.Finding {
	return copyFindings(tr.existingFindings)
}

func (tr Result) ResolvedCount() int {
	return tr.resolvedCount
}

func ClassifyFindings(current []finding.Finding, previousFingerprints []finding.FingerprintID) Result {
	previousSet := buildFingerprintSet(previousFingerprints)
	matchedPrevious := make(map[string]bool, len(previousFingerprints))

	var newFindings, existingFindings []finding.Finding

	for _, currentFinding := range current {
		fpValue := currentFinding.ID().Value()
		if previousSet[fpValue] {
			existingFindings = append(existingFindings, currentFinding)
			matchedPrevious[fpValue] = true
		} else {
			newFindings = append(newFindings, currentFinding)
		}
	}

	resolvedCount := countUnmatched(previousFingerprints, matchedPrevious)

	return NewResult(newFindings, existingFindings, resolvedCount)
}

func buildFingerprintSet(fingerprints []finding.FingerprintID) map[string]bool {
	set := make(map[string]bool, len(fingerprints))
	for _, fp := range fingerprints {
		set[fp.Value()] = true
	}
	return set
}

func countUnmatched(fingerprints []finding.FingerprintID, matched map[string]bool) int {
	count := 0
	for _, fp := range fingerprints {
		if !matched[fp.Value()] {
			count++
		}
	}
	return count
}

func copyFindings(findings []finding.Finding) []finding.Finding {
	if findings == nil {
		return nil
	}
	copied := make([]finding.Finding, len(findings))
	copy(copied, findings)
	return copied
}
