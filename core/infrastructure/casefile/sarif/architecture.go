package sarif

import (
	"encoding/json"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type archDocument struct {
	Runs []archRun `json:"runs"`
}

type archRun struct {
	Results []archResult `json:"results"`
}

type archResult struct {
	RuleID     string         `json:"ruleId"`
	Message    message        `json:"message"`
	Properties archProperties `json:"properties"`
}

type archProperties struct {
	SourcePkg string `json:"sourcePkg"`
	TargetPkg string `json:"targetPkg"`
}

func ParseArchitectureViolations(data []byte) ([]evidencedto.Violation, error) {
	var doc archDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	var violations []evidencedto.Violation
	for _, run := range doc.Runs {
		for _, r := range run.Results {
			violations = append(violations, evidencedto.Violation{
				Rule:      r.RuleID,
				SourcePkg: r.Properties.SourcePkg,
				TargetPkg: r.Properties.TargetPkg,
				Message:   r.Message.Text,
			})
		}
	}
	return violations, nil
}
