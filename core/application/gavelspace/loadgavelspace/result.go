package loadgavelspace

import (
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Result struct {
	Gavelspace gavelspacemodel.Gavelspace
	Projects   []projectmodel.Project
	View       WorkspaceView
}

type WorkspaceView struct {
	GavelspaceName string
	ServerURL      string
	ServerToken    string
	FindingsSource string
	Projects       []ProjectView
}

type ProjectView struct {
	ID            string
	Name          string
	TargetPattern string
	DefaultBranch string
	Languages     []string
	GateRules     []GateRuleView
	Baseline      BaselineView
	ArchPolicy    *ArchPolicyView
}

type GateRuleView struct {
	Subtype string
}

type BaselineView struct {
	FingerprintCount int
	ArchIDCount      int
	ArchIDs          []string
	HasPrevious      bool
}

type ArchPolicyView struct {
	Layers []ArchLayerView
	Rules  []ArchDenyRuleView
}

type ArchLayerView struct {
	Name     string
	Patterns []string
}

type ArchDenyRuleView struct {
	Name   string
	Source string
	Deny   []string
}

func buildView(gavelspace gavelspacemodel.Gavelspace, projects []projectmodel.Project) WorkspaceView {
	view := WorkspaceView{
		GavelspaceName: gavelspace.ID().String(),
		FindingsSource: gavelspace.FindingsSource(),
	}
	if gavelspace.ServerConfig().IsConfigured() {
		view.ServerURL = gavelspace.ServerConfig().URL()
		view.ServerToken = gavelspace.ServerConfig().Token()
	}
	for _, project := range projects {
		langs := make([]string, 0, len(project.Languages()))
		for _, l := range project.Languages() {
			langs = append(langs, l.String())
		}

		var rules []GateRuleView
		for _, r := range project.Gate().Rules() {
			rules = append(rules, GateRuleView{
				Subtype: r.Subtype().String(),
			})
		}

		bl := project.Baseline(project.DefaultBranch())
		baselineView := BaselineView{
			FingerprintCount: len(bl.Fingerprints()),
			ArchIDCount:      len(bl.ArchIDs()),
			ArchIDs:          bl.ArchIDs(),
			HasPrevious:      bl.HasPrevious(),
		}

		var archPolicyView *ArchPolicyView
		if pol := project.Policy(); pol != nil {
			policyView := ArchPolicyView{}
			for _, l := range pol.Layers() {
				policyView.Layers = append(policyView.Layers, ArchLayerView{Name: l.Name(), Patterns: l.Patterns()})
			}
			for _, r := range pol.DenyRules() {
				policyView.Rules = append(policyView.Rules, ArchDenyRuleView{Name: r.Name(), Source: r.Source(), Deny: r.Deny()})
			}
			archPolicyView = &policyView
		}

		view.Projects = append(view.Projects, ProjectView{
			ID:            project.ID().String(),
			Name:          project.Name(),
			TargetPattern: project.TargetPattern(),
			DefaultBranch: project.DefaultBranch(),
			Languages:     langs,
			GateRules:     rules,
			Baseline:      baselineView,
			ArchPolicy:    archPolicyView,
		})
	}
	return view
}
