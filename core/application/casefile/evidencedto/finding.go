package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type Finding struct {
	Tool          string
	RuleID        string
	Severity      string
	FilePath      string
	Line          int
	Message       string
	FingerprintID string
}

func ExtractFindings(evidences []Evidence) []Finding {
	var out []Finding
	seen := make(map[string]struct{})
	for _, ev := range evidences {
		for _, finding := range ev.Findings {
			if finding.FingerprintID != "" {
				if _, dup := seen[finding.FingerprintID]; dup {
					continue
				}
				seen[finding.FingerprintID] = struct{}{}
			}
			out = append(out, finding)
		}
	}
	return out
}

func ExtractFingerprints(findings []Finding) []string {
	out := make([]string, 0, len(findings))
	for _, finding := range findings {
		if finding.FingerprintID != "" {
			out = append(out, finding.FingerprintID)
		}
	}
	return out
}

func fromDomainFindings(findings []finding.Finding) []Finding {
	out := make([]Finding, 0, len(findings))
	for _, domainFinding := range findings {
		out = append(out, Finding{
			Tool:          domainFinding.Tool(),
			RuleID:        domainFinding.RuleID(),
			Severity:      domainFinding.Severity().String(),
			FilePath:      domainFinding.FilePath(),
			Line:          domainFinding.Line(),
			Message:       domainFinding.Message(),
			FingerprintID: domainFinding.ID().Value(),
		})
	}
	return out
}

func toDomainFindings(dtos []Finding) ([]finding.Finding, error) {
	out := make([]finding.Finding, 0, len(dtos))
	for index, input := range dtos {
		severity, err := finding.NewSeverity(input.Severity)
		if err != nil {
			return nil, fmt.Errorf("findings[%d] severity: %w", index, err)
		}
		fingerprint, err := finding.NewFingerprintID(input.FingerprintID)
		if err != nil {
			return nil, fmt.Errorf("findings[%d] fingerprint: %w", index, err)
		}
		domainFinding, err := finding.NewFinding(input.Tool, input.RuleID, severity, input.FilePath, input.Line, input.Message, fingerprint)
		if err != nil {
			return nil, fmt.Errorf("findings[%d]: %w", index, err)
		}
		out = append(out, domainFinding)
	}
	return out, nil
}
