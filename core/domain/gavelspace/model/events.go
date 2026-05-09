package model

import (
	"time"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

const (
	EventNameProjectAdded   = "gavelspace.project_added"
	EventNameProjectRemoved = "gavelspace.project_removed"
)

type ProjectAdded struct {
	gavelspaceName GavelspaceID
	projectID      projectmodel.ProjectID
	targetPattern  string
	occurredAt     time.Time
}

func NewProjectAdded(gavelspaceName GavelspaceID, projectID projectmodel.ProjectID, targetPattern string, occurredAt time.Time) ProjectAdded {
	return ProjectAdded{
		gavelspaceName: gavelspaceName,
		projectID:      projectID,
		targetPattern:  targetPattern,
		occurredAt:     occurredAt,
	}
}

func (e ProjectAdded) EventName() string                 { return EventNameProjectAdded }
func (e ProjectAdded) OccurredAt() time.Time             { return e.occurredAt }
func (e ProjectAdded) GavelspaceID() GavelspaceID        { return e.gavelspaceName }
func (e ProjectAdded) ProjectID() projectmodel.ProjectID { return e.projectID }
func (e ProjectAdded) TargetPattern() string             { return e.targetPattern }

type ProjectRemoved struct {
	gavelspaceName GavelspaceID
	projectID      projectmodel.ProjectID
	occurredAt     time.Time
}

func NewProjectRemoved(gavelspaceName GavelspaceID, projectID projectmodel.ProjectID, occurredAt time.Time) ProjectRemoved {
	return ProjectRemoved{
		gavelspaceName: gavelspaceName,
		projectID:      projectID,
		occurredAt:     occurredAt,
	}
}

func (e ProjectRemoved) EventName() string                 { return EventNameProjectRemoved }
func (e ProjectRemoved) OccurredAt() time.Time             { return e.occurredAt }
func (e ProjectRemoved) GavelspaceID() GavelspaceID        { return e.gavelspaceName }
func (e ProjectRemoved) ProjectID() projectmodel.ProjectID { return e.projectID }
