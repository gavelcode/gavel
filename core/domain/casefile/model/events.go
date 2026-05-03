package model

import (
	"time"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

const (
	EventNameCaseFileOpened    = "casefile.opened"
	EventNameEvidenceCollected = "casefile.evidence_collected"
	EventNameVerdictRendered   = "casefile.verdict_rendered"
	EventNameQualityGateFailed = "casefile.quality_gate_failed"
)

type CaseFileOpened struct {
	caseFileID CaseFileID
	projectID  projectmodel.ProjectID
	commitSHA  string
	branch     string
	occurredAt time.Time
}

func NewCaseFileOpened(caseFileID CaseFileID, projectID projectmodel.ProjectID, commitSHA, branch string, occurredAt time.Time) CaseFileOpened {
	return CaseFileOpened{
		caseFileID: caseFileID,
		projectID:  projectID,
		commitSHA:  commitSHA,
		branch:     branch,
		occurredAt: occurredAt,
	}
}

func (e CaseFileOpened) EventName() string                 { return EventNameCaseFileOpened }
func (e CaseFileOpened) OccurredAt() time.Time             { return e.occurredAt }
func (e CaseFileOpened) CaseFileID() CaseFileID            { return e.caseFileID }
func (e CaseFileOpened) ProjectID() projectmodel.ProjectID { return e.projectID }
func (e CaseFileOpened) CommitSHA() string                 { return e.commitSHA }
func (e CaseFileOpened) Branch() string                    { return e.branch }

type EvidenceCollected struct {
	caseFileID CaseFileID
	projectID  projectmodel.ProjectID
	subtype    string
	source     string
	occurredAt time.Time
}

func NewEvidenceCollected(caseFileID CaseFileID, projectID projectmodel.ProjectID, subtype, source string, occurredAt time.Time) EvidenceCollected {
	return EvidenceCollected{
		caseFileID: caseFileID,
		projectID:  projectID,
		subtype:    subtype,
		source:     source,
		occurredAt: occurredAt,
	}
}

func (e EvidenceCollected) EventName() string                 { return EventNameEvidenceCollected }
func (e EvidenceCollected) OccurredAt() time.Time             { return e.occurredAt }
func (e EvidenceCollected) CaseFileID() CaseFileID            { return e.caseFileID }
func (e EvidenceCollected) ProjectID() projectmodel.ProjectID { return e.projectID }
func (e EvidenceCollected) Subtype() string                   { return e.subtype }
func (e EvidenceCollected) Source() string                    { return e.source }

type VerdictRendered struct {
	caseFileID CaseFileID
	projectID  projectmodel.ProjectID
	outcome    string
	occurredAt time.Time
}

func NewVerdictRendered(caseFileID CaseFileID, projectID projectmodel.ProjectID, outcome string, occurredAt time.Time) VerdictRendered {
	return VerdictRendered{
		caseFileID: caseFileID,
		projectID:  projectID,
		outcome:    outcome,
		occurredAt: occurredAt,
	}
}

func (e VerdictRendered) EventName() string                 { return EventNameVerdictRendered }
func (e VerdictRendered) OccurredAt() time.Time             { return e.occurredAt }
func (e VerdictRendered) CaseFileID() CaseFileID            { return e.caseFileID }
func (e VerdictRendered) ProjectID() projectmodel.ProjectID { return e.projectID }
func (e VerdictRendered) Outcome() string                   { return e.outcome }

type QualityGateFailed struct {
	caseFileID      CaseFileID
	projectID       projectmodel.ProjectID
	failingSubtypes []string
	occurredAt      time.Time
}

func NewQualityGateFailed(caseFileID CaseFileID, projectID projectmodel.ProjectID, failingSubtypes []string, occurredAt time.Time) QualityGateFailed {
	copied := make([]string, len(failingSubtypes))
	copy(copied, failingSubtypes)
	return QualityGateFailed{
		caseFileID:      caseFileID,
		projectID:       projectID,
		failingSubtypes: copied,
		occurredAt:      occurredAt,
	}
}

func (e QualityGateFailed) EventName() string                 { return EventNameQualityGateFailed }
func (e QualityGateFailed) OccurredAt() time.Time             { return e.occurredAt }
func (e QualityGateFailed) CaseFileID() CaseFileID            { return e.caseFileID }
func (e QualityGateFailed) ProjectID() projectmodel.ProjectID { return e.projectID }

func (e QualityGateFailed) FailingSubtypes() []string {
	copied := make([]string, len(e.failingSubtypes))
	copy(copied, e.failingSubtypes)
	return copied
}
