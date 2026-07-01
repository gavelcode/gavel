package model

import "time"

const (
	EventNameQualityGateUpdated        = "project.quality_gate_updated"
	EventNameLanguagesUpdated          = "project.languages_updated"
	EventNameArchitecturePolicyUpdated = "project.architecture_policy_updated"
	EventNameTargetPatternUpdated      = "project.target_pattern_updated"
	EventNameExcludePatternsUpdated    = "project.exclude_patterns_updated"
	EventNameToolSelectionUpdated      = "project.tool_selection_updated"
)

type QualityGateUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewQualityGateUpdated(projectID ProjectID, occurredAt time.Time) QualityGateUpdated {
	return QualityGateUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e QualityGateUpdated) EventName() string     { return EventNameQualityGateUpdated }
func (e QualityGateUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e QualityGateUpdated) ProjectID() ProjectID  { return e.projectID }

type LanguagesUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewLanguagesUpdated(projectID ProjectID, occurredAt time.Time) LanguagesUpdated {
	return LanguagesUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e LanguagesUpdated) EventName() string     { return EventNameLanguagesUpdated }
func (e LanguagesUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e LanguagesUpdated) ProjectID() ProjectID  { return e.projectID }

type ToolSelectionUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewToolSelectionUpdated(projectID ProjectID, occurredAt time.Time) ToolSelectionUpdated {
	return ToolSelectionUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e ToolSelectionUpdated) EventName() string     { return EventNameToolSelectionUpdated }
func (e ToolSelectionUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e ToolSelectionUpdated) ProjectID() ProjectID  { return e.projectID }

type TargetPatternUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewTargetPatternUpdated(projectID ProjectID, occurredAt time.Time) TargetPatternUpdated {
	return TargetPatternUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e TargetPatternUpdated) EventName() string     { return EventNameTargetPatternUpdated }
func (e TargetPatternUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e TargetPatternUpdated) ProjectID() ProjectID  { return e.projectID }

type ExcludePatternsUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewExcludePatternsUpdated(projectID ProjectID, occurredAt time.Time) ExcludePatternsUpdated {
	return ExcludePatternsUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e ExcludePatternsUpdated) EventName() string     { return EventNameExcludePatternsUpdated }
func (e ExcludePatternsUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e ExcludePatternsUpdated) ProjectID() ProjectID  { return e.projectID }

type ArchitecturePolicyUpdated struct {
	projectID  ProjectID
	occurredAt time.Time
}

func NewArchitecturePolicyUpdated(projectID ProjectID, occurredAt time.Time) ArchitecturePolicyUpdated {
	return ArchitecturePolicyUpdated{projectID: projectID, occurredAt: occurredAt}
}

func (e ArchitecturePolicyUpdated) EventName() string     { return EventNameArchitecturePolicyUpdated }
func (e ArchitecturePolicyUpdated) OccurredAt() time.Time { return e.occurredAt }
func (e ArchitecturePolicyUpdated) ProjectID() ProjectID  { return e.projectID }
