package evidencedto

import (
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

type Evidence struct {
	ID              string
	Subtype         string
	Source          string
	CollectedAt     time.Time
	Findings        []Finding
	Coverage        *Coverage
	NewCodeCoverage *NewCodeCoverage
	License         *License
	Architecture    *Architecture
	ToolExecution   *ToolExecution
}

func EvidenceFromDomain(domainEvidence evidence.Evidence) Evidence {
	out := Evidence{
		ID:          domainEvidence.ID().String(),
		Subtype:     domainEvidence.Subtype().String(),
		Source:      domainEvidence.Source(),
		CollectedAt: domainEvidence.CollectedAt(),
	}
	switch content := domainEvidence.Content().(type) {
	case finding.Content:
		out.Findings = fromDomainFindings(content.Findings())
	case coverage.Content:
		cov := fromDomainCoverage(content)
		out.Coverage = &cov
	case license.Content:
		lic := fromDomainLicense(content)
		out.License = &lic
	case coverage.PatchContent:
		ncc := fromDomainNewCodeCoverage(content)
		out.NewCodeCoverage = &ncc
	case architecture.Content:
		arch := fromDomainArchitecture(content)
		out.Architecture = &arch
	case toolexecution.Content:
		exec := fromDomainToolExecution(content)
		out.ToolExecution = &exec
	}
	return out
}

func EvidenceToDomain(input Evidence) (evidence.Evidence, error) {
	subtype, err := evidence.NewSubtype(input.Subtype)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("evidence subtype: %w", err)
	}
	content, err := toDomainContent(subtype, input)
	if err != nil {
		return evidence.Evidence{}, err
	}
	if input.ID == "" {
		return evidence.NewEvidence(subtype, input.Source, content, input.CollectedAt)
	}
	id, err := evidence.ParseEvidenceID(input.ID)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("evidence id: %w", err)
	}
	return evidence.ReconstituteEvidence(id, subtype, input.Source, content, input.CollectedAt)
}

func FilterNewViolations(evidence *Evidence, newIDs map[string]bool) *Evidence {
	if evidence == nil || evidence.Architecture == nil {
		return evidence
	}
	var filtered []Violation
	for _, v := range evidence.Architecture.Violations {
		id := v.Rule + ":" + v.SourcePkg + ":" + v.TargetPkg
		if newIDs[id] {
			filtered = append(filtered, v)
		}
	}
	result := *evidence
	result.Architecture = &Architecture{Violations: filtered}
	return &result
}

func ReplaceArchEvidence(evidences []Evidence, archEvidence *Evidence) []Evidence {
	if archEvidence == nil {
		var result []Evidence
		for _, ev := range evidences {
			if ev.Subtype != "architecture" {
				result = append(result, ev)
			}
		}
		return result
	}
	result := make([]Evidence, 0, len(evidences))
	replaced := false
	for _, ev := range evidences {
		if ev.Subtype == "architecture" {
			result = append(result, *archEvidence)
			replaced = true
		} else {
			result = append(result, ev)
		}
	}
	if !replaced {
		result = append(result, *archEvidence)
	}
	return result
}

func toDomainContent(subtype evidence.Subtype, input Evidence) (evidence.Content, error) {
	switch subtype {
	case evidence.SubtypeCoverage:
		if input.Coverage == nil {
			return nil, fmt.Errorf("%w: coverage subtype requires coverage payload", ErrIncompatibleEvidence)
		}
		return toDomainCoverage(*input.Coverage)
	case evidence.SubtypeLicense:
		if input.License == nil {
			return nil, fmt.Errorf("%w: license subtype requires license payload", ErrIncompatibleEvidence)
		}
		return toDomainLicense(*input.License)
	case evidence.SubtypeNewCodeCoverage:
		if input.NewCodeCoverage == nil {
			return nil, fmt.Errorf("%w: new_code_coverage subtype requires new_code_coverage payload", ErrIncompatibleEvidence)
		}
		return toDomainNewCodeCoverage(*input.NewCodeCoverage)
	case evidence.SubtypeArchitecture:
		if input.Architecture == nil {
			return nil, fmt.Errorf("%w: architecture subtype requires architecture payload", ErrIncompatibleEvidence)
		}
		return toDomainArchitecture(*input.Architecture)
	case evidence.SubtypeToolExecution:
		if input.ToolExecution == nil {
			return nil, fmt.Errorf("%w: tool_execution subtype requires tool_execution payload", ErrIncompatibleEvidence)
		}
		return toDomainToolExecution(*input.ToolExecution)
	default:
		findings, err := toDomainFindings(input.Findings)
		if err != nil {
			return nil, err
		}
		return finding.NewContent(subtype, findings)
	}
}
