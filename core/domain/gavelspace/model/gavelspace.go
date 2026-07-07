package model

import (
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type Gavelspace struct {
	id             GavelspaceID
	tenantID       tenant.TenantID
	projects       []ProjectRef
	serverConfig   ServerConfig
	findingsSource string
	events         []event.DomainEvent
}

func NewGavelspace(tenantID tenant.TenantID, name string) (Gavelspace, error) {
	id, err := NewGavelspaceID(name)
	if err != nil {
		return Gavelspace{}, err
	}
	return Gavelspace{id: id, tenantID: tenantID}, nil
}

func ReconstituteGavelspace(gavelspaceID GavelspaceID, tenantID tenant.TenantID, projects []ProjectRef) (Gavelspace, error) {
	seen := make(map[string]bool, len(projects))
	for _, p := range projects {
		if seen[p.targetPattern] {
			return Gavelspace{}, fmt.Errorf("%w: %s", ErrDuplicateTargetPattern, p.targetPattern)
		}
		seen[p.targetPattern] = true
	}
	copied := make([]ProjectRef, len(projects))
	copy(copied, projects)
	return Gavelspace{id: gavelspaceID, tenantID: tenantID, projects: copied}, nil
}

func (g *Gavelspace) ID() GavelspaceID           { return g.id }
func (g *Gavelspace) TenantID() tenant.TenantID  { return g.tenantID }
func (g *Gavelspace) ServerConfig() ServerConfig { return g.serverConfig }
func (g *Gavelspace) FindingsSource() string     { return g.findingsSource }

func (g *Gavelspace) SetServerConfig(sc ServerConfig) { g.serverConfig = sc }
func (g *Gavelspace) SetFindingsSource(source string) { g.findingsSource = source }

func (g *Gavelspace) Projects() []ProjectRef {
	copied := make([]ProjectRef, len(g.projects))
	copy(copied, g.projects)
	return copied
}

func (g *Gavelspace) AddProject(ref ProjectRef, occurredAt time.Time) error {
	for _, p := range g.projects {
		if p.targetPattern == ref.targetPattern {
			return fmt.Errorf("%w: %s", ErrDuplicateTargetPattern, ref.targetPattern)
		}
	}
	g.projects = append(g.projects, ref)
	g.events = append(g.events, NewProjectAdded(g.id, ref.id, ref.targetPattern, occurredAt))
	return nil
}

func (g *Gavelspace) RemoveProject(projectID projectmodel.ProjectID, occurredAt time.Time) error {
	for i, p := range g.projects {
		if p.id.Equal(projectID) {
			g.projects = append(g.projects[:i], g.projects[i+1:]...)
			g.events = append(g.events, NewProjectRemoved(g.id, projectID, occurredAt))
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
}

func (g *Gavelspace) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(g.events))
	copy(copied, g.events)
	return copied
}

func (g *Gavelspace) ClearEvents() {
	g.events = nil
}
