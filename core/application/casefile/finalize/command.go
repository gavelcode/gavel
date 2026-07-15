package finalize

import (
	"fmt"
	"strings"
	"time"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type PrecomputedVerdict struct {
	Outcome     string
	Rulings     []RulingInput
	EvaluatedAt time.Time
}

type RulingInput struct {
	Subtype string
	Passed  bool
	Detail  string
}

type Command struct {
	tenantID           string
	caseFileID         string
	fingerprints       []string
	archIDs            []string
	archDelta          ArchDeltaInput
	quick              bool
	absolute           bool
	noBaselineUpdate   bool
	precomputedVerdict *PrecomputedVerdict
	fileCoverage       []projectmodel.FileCoverageEntry
}

func NewCommand(tenantID, caseFileID string, opts ...Option) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(caseFileID) == "" {
		return Command{}, fmt.Errorf("%w: caseFileID must not be empty", ErrInvalidCommand)
	}
	cmd := Command{tenantID: tenantID, caseFileID: caseFileID}
	for _, opt := range opts {
		opt(&cmd)
	}
	if pv := cmd.precomputedVerdict; pv != nil {
		if strings.TrimSpace(pv.Outcome) == "" {
			return Command{}, fmt.Errorf("%w: precomputed verdict outcome must not be empty", ErrInvalidCommand)
		}
		if pv.EvaluatedAt.IsZero() {
			return Command{}, fmt.Errorf("%w: precomputed verdict evaluatedAt must not be zero", ErrInvalidCommand)
		}
	}
	return cmd, nil
}

type Option func(*Command)

func WithFingerprints(fps []string) Option {
	return func(c *Command) { c.fingerprints = fps }
}

func WithArchIDs(ids []string) Option {
	return func(c *Command) { c.archIDs = ids }
}

func WithArchDelta(d ArchDeltaInput) Option {
	return func(c *Command) { c.archDelta = d }
}

func WithQuick(q bool) Option {
	return func(c *Command) { c.quick = q }
}

func WithAbsolute(a bool) Option {
	return func(c *Command) { c.absolute = a }
}

func WithoutBaselineUpdate(skip bool) Option {
	return func(c *Command) { c.noBaselineUpdate = skip }
}

func WithFileCoverage(fc []projectmodel.FileCoverageEntry) Option {
	return func(c *Command) { c.fileCoverage = fc }
}

func WithPrecomputedVerdict(v PrecomputedVerdict) Option {
	return func(c *Command) { c.precomputedVerdict = &v }
}

func (c Command) TenantID() string                               { return c.tenantID }
func (c Command) CaseFileID() string                             { return c.caseFileID }
func (c Command) Fingerprints() []string                         { return c.fingerprints }
func (c Command) ArchIDs() []string                              { return c.archIDs }
func (c Command) ArchDelta() ArchDeltaInput                      { return c.archDelta }
func (c Command) Quick() bool                                    { return c.quick }
func (c Command) Absolute() bool                                 { return c.absolute }
func (c Command) NoBaselineUpdate() bool                         { return c.noBaselineUpdate }
func (c Command) FileCoverage() []projectmodel.FileCoverageEntry { return c.fileCoverage }
func (c Command) PrecomputedVerdict() *PrecomputedVerdict        { return c.precomputedVerdict }
