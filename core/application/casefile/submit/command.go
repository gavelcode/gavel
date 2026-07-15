package submit

import (
	"fmt"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
)

type Command struct {
	tenantID         string
	projectID        string
	commitSHA        string
	branch           string
	evidences        []evidencedto.Evidence
	fingerprints     []string
	archIDs          []string
	archDelta        finalize.ArchDeltaInput
	fileCoverage     []evidencedto.FileCoverage
	quick            bool
	absolute         bool
	noBaselineUpdate bool
	startedAt        time.Time
}

func NewCommand(
	tenantID, projectID, commitSHA, branch string,
	evidences []evidencedto.Evidence,
	fingerprints, archIDs []string,
	archDelta finalize.ArchDeltaInput,
	fileCoverage []evidencedto.FileCoverage,
	quick, absolute bool,
	startedAt time.Time,
) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(commitSHA) == "" {
		return Command{}, fmt.Errorf("%w: commitSHA must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(branch) == "" {
		return Command{}, fmt.Errorf("%w: branch must not be empty", ErrInvalidCommand)
	}
	return Command{
		tenantID:     tenantID,
		projectID:    projectID,
		commitSHA:    commitSHA,
		branch:       branch,
		evidences:    evidences,
		fingerprints: fingerprints,
		archIDs:      archIDs,
		archDelta:    archDelta,
		fileCoverage: fileCoverage,
		quick:        quick,
		absolute:     absolute,
		startedAt:    startedAt,
	}, nil
}

func (c Command) TenantID() string                         { return c.tenantID }
func (c Command) ProjectID() string                        { return c.projectID }
func (c Command) CommitSHA() string                        { return c.commitSHA }
func (c Command) Branch() string                           { return c.branch }
func (c Command) Evidences() []evidencedto.Evidence        { return c.evidences }
func (c Command) Fingerprints() []string                   { return c.fingerprints }
func (c Command) ArchIDs() []string                        { return c.archIDs }
func (c Command) ArchDelta() finalize.ArchDeltaInput       { return c.archDelta }
func (c Command) FileCoverage() []evidencedto.FileCoverage { return c.fileCoverage }
func (c Command) Quick() bool                              { return c.quick }
func (c Command) Absolute() bool                           { return c.absolute }
func (c Command) NoBaselineUpdate() bool                   { return c.noBaselineUpdate }

func (c Command) WithoutBaselineUpdate(skip bool) Command {
	c.noBaselineUpdate = skip
	return c
}
func (c Command) StartedAt() time.Time { return c.startedAt }
