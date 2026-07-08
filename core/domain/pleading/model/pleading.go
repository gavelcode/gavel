package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type Pleading struct {
	id           PleadingID
	tenantID     tenant.TenantID
	projectID    projectmodel.ProjectID
	number       int
	title        string
	petitioner   string
	sourceBranch string
	targetBranch string
	commitSHA    string
	status       Status
	events       []event.DomainEvent
}

func FilePleading(
	tenantID tenant.TenantID,
	projectID projectmodel.ProjectID,
	number int,
	title, petitioner, sourceBranch, targetBranch, commitSHA string,
) (Pleading, error) {
	if err := validatePleadingFields(projectID, number, title, sourceBranch, targetBranch, commitSHA); err != nil {
		return Pleading{}, err
	}
	pleadingID := NewPleadingID(uuid.New())
	return Pleading{
		id:           pleadingID,
		tenantID:     tenantID,
		projectID:    projectID,
		number:       number,
		title:        title,
		petitioner:   petitioner,
		sourceBranch: sourceBranch,
		targetBranch: targetBranch,
		commitSHA:    commitSHA,
		status:       StatusOpen,
	}, nil
}

func ReconstitutePleading(
	pleadingID PleadingID,
	tenantID tenant.TenantID,
	projectID projectmodel.ProjectID,
	number int,
	title, petitioner, sourceBranch, targetBranch, commitSHA string,
	status Status,
) (Pleading, error) {
	if err := validatePleadingFields(projectID, number, title, sourceBranch, targetBranch, commitSHA); err != nil {
		return Pleading{}, err
	}
	return Pleading{
		id:           pleadingID,
		tenantID:     tenantID,
		projectID:    projectID,
		number:       number,
		title:        title,
		petitioner:   petitioner,
		sourceBranch: sourceBranch,
		targetBranch: targetBranch,
		commitSHA:    commitSHA,
		status:       status,
	}, nil
}

func validatePleadingFields(projectID projectmodel.ProjectID, number int, title, sourceBranch, targetBranch, commitSHA string) error {
	if number <= 0 {
		return fmt.Errorf("%w: number must be positive", ErrInvalidPleading)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title must not be empty", ErrInvalidPleading)
	}
	if strings.TrimSpace(sourceBranch) == "" {
		return fmt.Errorf("%w: sourceBranch must not be empty", ErrInvalidPleading)
	}
	if strings.TrimSpace(targetBranch) == "" {
		return fmt.Errorf("%w: targetBranch must not be empty", ErrInvalidPleading)
	}
	if strings.TrimSpace(commitSHA) == "" {
		return fmt.Errorf("%w: commitSHA must not be empty", ErrInvalidPleading)
	}
	return nil
}

func (p *Pleading) MarkMerged(occurredAt time.Time) error {
	if p.status.IsTerminal() {
		return fmt.Errorf("%w: cannot merge %s pleading", ErrInvalidTransition, p.status)
	}
	p.status = StatusMerged
	p.events = append(p.events, NewMerged(p.id, occurredAt))
	return nil
}

func (p *Pleading) MarkClosed(occurredAt time.Time) error {
	if p.status.IsTerminal() {
		return fmt.Errorf("%w: cannot close %s pleading", ErrInvalidTransition, p.status)
	}
	p.status = StatusClosed
	p.events = append(p.events, NewClosed(p.id, occurredAt))
	return nil
}

func (p Pleading) ID() PleadingID                    { return p.id }
func (p Pleading) TenantID() tenant.TenantID         { return p.tenantID }
func (p Pleading) ProjectID() projectmodel.ProjectID { return p.projectID }
func (p Pleading) Number() int                       { return p.number }
func (p Pleading) Title() string                     { return p.title }
func (p Pleading) Petitioner() string                { return p.petitioner }
func (p Pleading) SourceBranch() string              { return p.sourceBranch }
func (p Pleading) TargetBranch() string              { return p.targetBranch }
func (p Pleading) CommitSHA() string                 { return p.commitSHA }
func (p Pleading) Status() Status                    { return p.status }

func (p Pleading) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(p.events))
	copy(copied, p.events)
	return copied
}

func (p *Pleading) ClearEvents() {
	p.events = nil
}
