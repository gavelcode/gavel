package project

import (
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	updatequalitygate "github.com/usegavel/gavel/core/application/project/updatequalitygate"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func projectFromSummary(summary projectlist.ProjectSummary) gen.Project {
	return gen.Project{
		Id:            httpx.ParseUUIDOrZero(summary.ID),
		Key:           summary.Key,
		Name:          summary.Name,
		DefaultBranch: summary.DefaultBranch,
		LatestVerdict: summary.LatestVerdict,
		TotalFindings: int32(summary.TotalFindings),
		CreatedAt:     summary.CreatedAt,
	}
}

func ProjectDetailFromView(detail *projectgetbykey.ProjectDetail) gen.ProjectDetail {
	rules := make([]gen.ProjectRuleView, 0, len(detail.QualityGateRules))
	for _, rule := range detail.QualityGateRules {
		rules = append(rules, gen.ProjectRuleView{
			Subtype:      rule.Subtype,
			StrategyType: rule.StrategyType,
		})
	}
	counts := make(map[string]int32, len(detail.SeverityCounts))
	for k, v := range detail.SeverityCounts {
		counts[k] = int32(v)
	}
	return gen.ProjectDetail{
		Id:               httpx.ParseUUIDOrZero(detail.ID),
		Key:              detail.Key,
		Name:             detail.Name,
		DefaultBranch:    detail.DefaultBranch,
		LatestVerdict:    detail.LatestVerdict,
		TotalFindings:    int32(detail.TotalFindings),
		CreatedAt:        detail.CreatedAt,
		TargetPattern:    detail.TargetPattern,
		Languages:        nonNilStrings(detail.Languages),
		QualityGateRules: rules,
		SeverityCounts:   counts,
	}
}

func nonNilStrings(strategy []string) []string {
	if strategy == nil {
		return []string{}
	}
	return strategy
}

func qualityGateInput(gate gen.QualityGate) (updatequalitygate.Input, error) {
	rules := make([]updatequalitygate.RuleInput, 0, len(gate.Rules))
	for _, rule := range gate.Rules {
		strategy, err := strategyInput(rule.Strategy)
		if err != nil {
			return updatequalitygate.Input{}, err
		}
		ruleInput := updatequalitygate.RuleInput{
			Subtype:  rule.Subtype,
			Strategy: strategy,
		}
		if rule.MinResolved != nil {
			v := int(*rule.MinResolved)
			ruleInput.MinResolved = &v
		}
		if rule.MinDelta != nil {
			ruleInput.MinDelta = rule.MinDelta
		}
		rules = append(rules, ruleInput)
	}
	return updatequalitygate.Input{Rules: rules}, nil
}

func strategyInput(strategy gen.QualityGateStrategy) (updatequalitygate.StrategyInput, error) {
	out := updatequalitygate.StrategyInput{Type: string(strategy.Type)}
	if strategy.CountBySeverity != nil {
		out.CountBySeverity = &updatequalitygate.CountBySeverity{
			MaxError:   int(strategy.CountBySeverity.MaxError),
			MaxWarning: int(strategy.CountBySeverity.MaxWarning),
			MaxNote:    int(strategy.CountBySeverity.MaxNote),
		}
	}
	if strategy.MinPercentage != nil {
		out.MinPercentage = &updatequalitygate.MinPercentage{Min: strategy.MinPercentage.Min}
	}
	if strategy.ForbiddenList != nil {
		out.ForbiddenList = &updatequalitygate.ForbiddenList{Forbidden: strategy.ForbiddenList.Forbidden}
	}
	if strategy.MaxViolations != nil {
		out.MaxViolations = &updatequalitygate.MaxViolations{Max: int(strategy.MaxViolations.Max)}
	}
	if strategy.MinNewCodeCoverage != nil {
		out.MinNewCodeCoverage = &updatequalitygate.MinNewCodeCoverage{Min: strategy.MinNewCodeCoverage.Min}
	}
	return out, nil
}
