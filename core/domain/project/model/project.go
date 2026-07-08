package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

const defaultBranch = "main"

var keyRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

var targetPatternRegex = regexp.MustCompile(`^//([a-zA-Z0-9][a-zA-Z0-9._/-]*(\.\.\.|(:[a-zA-Z0-9._-]+))?|\.\.\.)$`)

const keyMaxLength = 64

type Project struct {
	id                 ProjectID
	tenantID           tenant.TenantID
	key                string
	name               string
	targetPattern      string
	excludePatterns    []string
	languages          []coverage.Language
	toolSelection      map[string][]string
	defaultBranch      string
	qualityGate        qualitygate.Gate
	architecturePolicy *archpolicy.Policy
	baselines          map[string]Baseline
	events             []event.DomainEvent
}

func NewProject(tenantID tenant.TenantID, key, name, targetPattern string) (Project, error) {
	if err := validateKey(key); err != nil {
		return Project{}, err
	}
	if err := validateProjectFields(name, targetPattern); err != nil {
		return Project{}, err
	}
	id := NewProjectID(uuid.New())
	return Project{
		id:            id,
		tenantID:      tenantID,
		key:           key,
		name:          name,
		targetPattern: targetPattern,
		defaultBranch: defaultBranch,
	}, nil
}

func ReconstituteProject(
	projectID ProjectID, tenantID tenant.TenantID, key, name, targetPattern, branch string,
	languages []coverage.Language,
	qualityGate qualitygate.Gate,
	architecturePolicy *archpolicy.Policy,
	baselines map[string]Baseline,
) (Project, error) {
	if err := validateKey(key); err != nil {
		return Project{}, err
	}
	if err := validateProjectFields(name, targetPattern); err != nil {
		return Project{}, err
	}
	if strings.TrimSpace(branch) == "" {
		return Project{}, fmt.Errorf("%w: branch must not be empty", ErrInvalidProject)
	}
	copied := make([]coverage.Language, len(languages))
	copy(copied, languages)
	var blCopy map[string]Baseline
	if len(baselines) > 0 {
		blCopy = make(map[string]Baseline, len(baselines))
		for k, v := range baselines {
			blCopy[k] = v
		}
	}
	return Project{
		id:                 projectID,
		tenantID:           tenantID,
		key:                key,
		name:               name,
		targetPattern:      targetPattern,
		defaultBranch:      branch,
		languages:          copied,
		qualityGate:        qualityGate,
		architecturePolicy: architecturePolicy,
		baselines:          blCopy,
	}, nil
}

func validateKey(key string) error {
	if len(key) < 1 || len(key) > keyMaxLength {
		return fmt.Errorf("%w: key must be 1-%d characters", ErrInvalidProject, keyMaxLength)
	}
	if !keyRegex.MatchString(key) {
		return fmt.Errorf("%w: key must be lowercase alphanumeric with hyphens", ErrInvalidProject)
	}
	return nil
}

func validateProjectFields(name, targetPattern string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidProject)
	}
	return validateTargetPattern(targetPattern)
}

func validateTargetPattern(targetPattern string) error {
	if !targetPatternRegex.MatchString(targetPattern) {
		return fmt.Errorf("%w: targetPattern must be a valid Bazel pattern (e.g. //pkg/...)", ErrInvalidProject)
	}
	return nil
}

func (p *Project) UpdateQualityGate(qg qualitygate.Gate, occurredAt time.Time) {
	p.qualityGate = qg
	p.events = append(p.events, NewQualityGateUpdated(p.id, occurredAt))
}

func (p *Project) UpdateTargetPattern(targetPattern string, occurredAt time.Time) error {
	if err := validateTargetPattern(targetPattern); err != nil {
		return err
	}
	p.targetPattern = targetPattern
	p.events = append(p.events, NewTargetPatternUpdated(p.id, occurredAt))
	return nil
}

func (p *Project) UpdateLanguages(languages []coverage.Language, occurredAt time.Time) {
	copied := make([]coverage.Language, len(languages))
	copy(copied, languages)
	p.languages = copied
	p.events = append(p.events, NewLanguagesUpdated(p.id, occurredAt))
}

func (p *Project) UpdateToolSelection(selection map[string][]string, occurredAt time.Time) {
	p.toolSelection = copyToolSelection(selection)
	p.events = append(p.events, NewToolSelectionUpdated(p.id, occurredAt))
}

func (p Project) ToolSelection() map[string][]string {
	return copyToolSelection(p.toolSelection)
}

func copyToolSelection(selection map[string][]string) map[string][]string {
	if selection == nil {
		return nil
	}
	copied := make(map[string][]string, len(selection))
	for language, tools := range selection {
		toolsCopy := make([]string, len(tools))
		copy(toolsCopy, tools)
		copied[language] = toolsCopy
	}
	return copied
}

func (p *Project) UpdateExcludePatterns(excludePatterns []string, occurredAt time.Time) error {
	if err := validateExcludePatterns(p.targetPattern, excludePatterns); err != nil {
		return err
	}
	copied := make([]string, len(excludePatterns))
	copy(copied, excludePatterns)
	p.excludePatterns = copied
	p.events = append(p.events, NewExcludePatternsUpdated(p.id, occurredAt))
	return nil
}

func validateExcludePatterns(targetPattern string, excludePatterns []string) error {
	scope := strings.TrimSuffix(targetPattern, "...")
	for _, pattern := range excludePatterns {
		if err := validateTargetPattern(pattern); err != nil {
			return fmt.Errorf("%w: exclude %q is not a valid Bazel pattern", ErrInvalidProject, pattern)
		}
		if !strings.HasPrefix(pattern, scope) {
			return fmt.Errorf("%w: exclude %q must resolve within %q", ErrInvalidProject, pattern, targetPattern)
		}
	}
	return nil
}

func (p Project) ID() ProjectID {
	return p.id
}

func (p Project) TenantID() tenant.TenantID {
	return p.tenantID
}

func (p Project) Key() string {
	return p.key
}

func (p Project) Name() string {
	return p.name
}

func (p Project) TargetPattern() string {
	return p.targetPattern
}

func (p Project) ExcludePatterns() []string {
	copied := make([]string, len(p.excludePatterns))
	copy(copied, p.excludePatterns)
	return copied
}

func (p Project) DefaultBranch() string {
	return p.defaultBranch
}

func (p Project) Languages() []coverage.Language {
	copied := make([]coverage.Language, len(p.languages))
	copy(copied, p.languages)
	return copied
}

func (p Project) Gate() qualitygate.Gate {
	return p.qualityGate
}

func (p *Project) UpdateArchitecturePolicy(policy archpolicy.Policy, occurredAt time.Time) {
	p.architecturePolicy = &policy
	p.events = append(p.events, NewArchitecturePolicyUpdated(p.id, occurredAt))
}

func (p Project) Baseline(branch string) Baseline {
	if p.baselines == nil {
		return Baseline{}
	}
	if bl, ok := p.baselines[branch]; ok {
		return bl
	}
	if branch != p.defaultBranch {
		return p.baselines[p.defaultBranch]
	}
	return Baseline{}
}

func (p *Project) SeedBaselineIfAbsent(branch string, fingerprints, archIDs []string, coveragePercent *float64, fileCoverage []FileCoverageEntry) bool {
	if _, ok := p.baselines[branch]; ok {
		return false
	}
	p.UpdateBaseline(branch, fingerprints, archIDs, coveragePercent, fileCoverage)
	return true
}

func (p *Project) UpdateBaseline(branch string, fingerprints, archIDs []string, coveragePercent *float64, fileCoverage []FileCoverageEntry) {
	if p.baselines == nil {
		p.baselines = make(map[string]Baseline)
	}
	p.baselines[branch] = NewBaseline(fingerprints, archIDs, coveragePercent, fileCoverage)
}

func (p *Project) RatchetBaseline(branch string, currentFingerprints, currentArchIDs []string) {
	existing, ok := p.baselines[branch]
	if !ok {
		return
	}
	p.baselines[branch] = existing.Ratchet(currentFingerprints, currentArchIDs)
}

func (p Project) Baselines() map[string]Baseline {
	if p.baselines == nil {
		return nil
	}
	cp := make(map[string]Baseline, len(p.baselines))
	for k, v := range p.baselines {
		cp[k] = v
	}
	return cp
}

func (p Project) Policy() *archpolicy.Policy {
	if p.architecturePolicy == nil {
		return nil
	}
	cp := *p.architecturePolicy
	return &cp
}

func (p Project) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(p.events))
	copy(copied, p.events)
	return copied
}

func (p *Project) ClearEvents() {
	p.events = nil
}
