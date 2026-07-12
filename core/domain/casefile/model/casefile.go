package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type CaseFile struct {
	id                CaseFileID
	tenantID          tenant.TenantID
	projectID         projectmodel.ProjectID
	commitSHA         string
	branch            string
	startedAt         time.Time
	evidences         []evidence.Evidence
	verdict           *verdict.Result
	isFreshEvaluation bool
	events            []event.DomainEvent
}

func NewCaseFile(tenantID tenant.TenantID, projectID projectmodel.ProjectID, commitSHA, branch string, startedAt, createdAt time.Time) (CaseFile, error) {
	if tenantID.IsZero() {
		return CaseFile{}, fmt.Errorf("%w: tenantID must not be zero", ErrInvalidCaseFile)
	}
	if err := validateCaseFileFields(commitSHA, branch, startedAt); err != nil {
		return CaseFile{}, err
	}
	if createdAt.IsZero() {
		return CaseFile{}, fmt.Errorf("%w: createdAt must not be zero", ErrInvalidCaseFile)
	}
	caseFileID := NewCaseFileID(uuid.New())
	return CaseFile{
		id:        caseFileID,
		tenantID:  tenantID,
		projectID: projectID,
		commitSHA: commitSHA,
		branch:    branch,
		startedAt: startedAt,
		events:    []event.DomainEvent{NewCaseFileOpened(caseFileID, projectID, commitSHA, branch, createdAt)},
	}, nil
}

func ReconstituteCaseFile(id CaseFileID, tenantID tenant.TenantID, projectID projectmodel.ProjectID, commitSHA, branch string, startedAt time.Time, evidences []evidence.Evidence, verdict *verdict.Result, isFreshEvaluation bool) (CaseFile, error) {
	if err := validateCaseFileFields(commitSHA, branch, startedAt); err != nil {
		return CaseFile{}, err
	}
	return CaseFile{
		id:                id,
		tenantID:          tenantID,
		projectID:         projectID,
		commitSHA:         commitSHA,
		branch:            branch,
		startedAt:         startedAt,
		evidences:         copyEvidences(evidences),
		verdict:           verdict,
		isFreshEvaluation: isFreshEvaluation,
	}, nil
}

func validateCaseFileFields(commitSHA, branch string, startedAt time.Time) error {
	if strings.TrimSpace(commitSHA) == "" {
		return fmt.Errorf("%w: commitSHA must not be empty", ErrInvalidCaseFile)
	}
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("%w: branch must not be empty", ErrInvalidCaseFile)
	}
	if startedAt.IsZero() {
		return fmt.Errorf("%w: startedAt must not be zero", ErrInvalidCaseFile)
	}
	return nil
}

func (cf *CaseFile) AddEvidence(ev evidence.Evidence, occurredAt time.Time) error {
	if cf.verdict != nil {
		return fmt.Errorf("%w: cannot add evidence after judgment", ErrAlreadyJudged)
	}
	cf.evidences = append(cf.evidences, ev)
	cf.events = append(cf.events, NewEvidenceCollected(cf.id, cf.projectID, ev.Subtype().String(), ev.Source(), occurredAt))
	return nil
}

func (cf *CaseFile) Judge(qualityGate qualitygate.Gate, tracking *tracking.Result, evaluatedAt time.Time, delta *DeltaInput) (verdict.Result, error) {
	if cf.verdict != nil {
		return verdict.Result{}, fmt.Errorf("%w: cannot judge twice", ErrAlreadyJudged)
	}

	grouped := cf.groupEvidenceBySubtype()
	cf.applyTrackingFilter(grouped, tracking)
	rulings := cf.evaluateRules(qualityGate, grouped, delta)
	rulings = append(rulings, cf.toolExecutionRuling(grouped))

	result, err := verdict.Compose(rulings, evaluatedAt)
	if err != nil {
		return verdict.Result{}, fmt.Errorf("compose verdict: %w", err)
	}
	cf.verdict = &result

	cf.emitVerdictEvents(result, rulings, evaluatedAt)

	return result, nil
}

func (cf *CaseFile) toolExecutionRuling(grouped map[evidence.Subtype]evidence.Content) verdict.Ruling {
	var hard, degraded []toolexecution.Failure
	for _, failed := range toolExecutionFailures(grouped) {
		if failed.Degraded() {
			degraded = append(degraded, failed)
		} else {
			hard = append(hard, failed)
		}
	}
	if len(hard) > 0 {
		return verdict.NewRuling(evidence.SubtypeToolExecution, false, toolExecutionReasons(hard))
	}
	if len(degraded) > 0 {
		return verdict.NewRuling(evidence.SubtypeToolExecution, true, "incomplete analysis — "+toolExecutionReasons(degraded))
	}
	return verdict.NewRuling(evidence.SubtypeToolExecution, true, "")
}

func toolExecutionReasons(failures []toolexecution.Failure) string {
	reasons := make([]string, 0, len(failures))
	for _, failed := range failures {
		reasons = append(reasons, fmt.Sprintf("%s: %s", failed.Tool(), failed.Reason()))
	}
	return strings.Join(reasons, "; ")
}

func toolExecutionFailures(grouped map[evidence.Subtype]evidence.Content) []toolexecution.Failure {
	content, ok := grouped[evidence.SubtypeToolExecution]
	if !ok {
		return nil
	}
	executions, ok := content.(toolexecution.Content)
	if !ok {
		return nil
	}
	return executions.Failures()
}

func (cf *CaseFile) RecordVerdict(v verdict.Result) error {
	if cf.verdict != nil {
		return fmt.Errorf("%w: cannot record verdict after judgment", ErrAlreadyJudged)
	}
	cf.verdict = &v
	cf.events = append(cf.events, NewVerdictRendered(cf.id, cf.projectID, v.Outcome().String(), v.EvaluatedAt()))
	return nil
}

func (cf *CaseFile) groupEvidenceBySubtype() map[evidence.Subtype]evidence.Content {
	grouped := make(map[evidence.Subtype]evidence.Content)
	for _, currentEvidence := range cf.evidences {
		sub := currentEvidence.Subtype()
		existing, ok := grouped[sub]
		if !ok {
			grouped[sub] = currentEvidence.Content()
			continue
		}
		merged, _ := existing.Merge(currentEvidence.Content())
		grouped[sub] = merged
	}
	return grouped
}

func (cf *CaseFile) applyTrackingFilter(grouped map[evidence.Subtype]evidence.Content, tracking *tracking.Result) {
	if tracking == nil {
		return
	}

	newFPs := buildNewFingerprintSet(tracking)

	for sub, content := range grouped {
		fc, ok := content.(finding.Content)
		if !ok {
			continue
		}
		grouped[sub] = filterToNewFindings(sub, fc, newFPs)
	}
}

func buildNewFingerprintSet(tracking *tracking.Result) map[string]bool {
	newFindings := tracking.NewFindings()
	set := make(map[string]bool, len(newFindings))
	for _, f := range newFindings {
		set[f.ID().Value()] = true
	}
	return set
}

func filterToNewFindings(sub evidence.Subtype, fc finding.Content, newFPs map[string]bool) evidence.Content {
	var filtered []finding.Finding
	for _, f := range fc.Findings() {
		if newFPs[f.ID().Value()] {
			filtered = append(filtered, f)
		}
	}
	content, _ := finding.NewContent(sub, filtered)
	return content
}

func (cf *CaseFile) evaluateRules(qg qualitygate.Gate, grouped map[evidence.Subtype]evidence.Content, delta *DeltaInput) []verdict.Ruling {
	rules := qg.Rules()
	rulings := make([]verdict.Ruling, 0, len(rules))

	for _, rule := range rules {
		content, hasEvidence := grouped[rule.Subtype()]
		absolutePassed := true
		absoluteDetail := ""
		if hasEvidence {
			outcome := rule.Strategy().Evaluate(content)
			absolutePassed = outcome.Passed()
			absoluteDetail = outcome.Detail()
		}

		deltaPassed, deltaDetail := evaluateDeltaCondition(rule, delta)

		passed := absolutePassed && deltaPassed
		detail := combineDetails(absoluteDetail, deltaDetail)

		rulings = append(rulings, verdict.NewRuling(rule.Subtype(), passed, detail))
	}

	return rulings
}

func evaluateDeltaCondition(rule qualitygate.Rule, delta *DeltaInput) (bool, string) {
	if delta == nil {
		return true, ""
	}

	if mr := rule.MinResolved(); mr != nil {
		resolved := resolvedCountForSubtype(rule.Subtype(), delta)
		if resolved < *mr {
			return false, fmt.Sprintf("resolved %d (min %d)", resolved, *mr)
		}
	}

	if minDelta := rule.MinDelta(); minDelta != nil {
		if delta.PreviousCoverage == nil {
			return true, ""
		}
		actual := delta.CurrentCoverage - *delta.PreviousCoverage
		if actual < *minDelta {
			return false, fmt.Sprintf("coverage delta %.1f%% (min %.1f%%)", actual, *minDelta)
		}
	}

	return true, ""
}

func resolvedCountForSubtype(subtype evidence.Subtype, delta *DeltaInput) int {
	if subtype == evidence.SubtypeArchitecture {
		return delta.ArchResolved
	}
	return delta.FindingsResolved
}

func combineDetails(absolute, delta string) string {
	if absolute == "" && delta == "" {
		return ""
	}
	if absolute == "" {
		return delta
	}
	if delta == "" {
		return absolute
	}
	return absolute + "; " + delta
}

func (cf *CaseFile) emitVerdictEvents(result verdict.Result, rulings []verdict.Ruling, occurredAt time.Time) {
	cf.events = append(cf.events, NewVerdictRendered(cf.id, cf.projectID, result.Outcome().String(), occurredAt))

	if result.Outcome() == verdict.OutcomeFail {
		cf.events = append(cf.events, NewQualityGateFailed(cf.id, cf.projectID, collectFailingSubtypes(rulings), occurredAt))
	}
}

func collectFailingSubtypes(rulings []verdict.Ruling) []string {
	var subtypes []string
	for _, r := range rulings {
		if !r.Passed() {
			subtypes = append(subtypes, r.Subtype().String())
		}
	}
	return subtypes
}

func (cf *CaseFile) ID() CaseFileID                    { return cf.id }
func (cf *CaseFile) TenantID() tenant.TenantID         { return cf.tenantID }
func (cf *CaseFile) ProjectID() projectmodel.ProjectID { return cf.projectID }
func (cf *CaseFile) CommitSHA() string                 { return cf.commitSHA }
func (cf *CaseFile) Branch() string                    { return cf.branch }
func (cf *CaseFile) StartedAt() time.Time              { return cf.startedAt }

func (cf *CaseFile) Evidences() []evidence.Evidence {
	return copyEvidences(cf.evidences)
}

func (cf *CaseFile) Verdict() (verdict.Result, bool) {
	if cf.verdict == nil {
		return verdict.Result{}, false
	}
	return *cf.verdict, true
}

func (cf *CaseFile) IsJudged() bool { return cf.verdict != nil }

func (cf *CaseFile) IsFreshEvaluation() bool { return cf.isFreshEvaluation }

func (cf *CaseFile) MarkFreshEvaluation() { cf.isFreshEvaluation = true }

func (cf *CaseFile) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(cf.events))
	copy(copied, cf.events)
	return copied
}

func (cf *CaseFile) ClearEvents() {
	cf.events = nil
}

func copyEvidences(evidences []evidence.Evidence) []evidence.Evidence {
	if evidences == nil {
		return nil
	}
	copied := make([]evidence.Evidence, len(evidences))
	copy(copied, evidences)
	return copied
}
