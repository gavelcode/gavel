package judge

import (
	"fmt"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
)

type Command struct {
	tenantID   string
	caseFileID string
	tracking   *evidencedto.Tracking
	deltaInput *casefile.DeltaInput
}

func NewCommand(tenantID, caseFileID string, tracking *evidencedto.Tracking, opts ...CommandOption) (Command, error) {
	if tenantID == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if caseFileID == "" {
		return Command{}, fmt.Errorf("%w: caseFileID must not be empty", ErrInvalidCommand)
	}
	cmd := Command{
		tenantID:   tenantID,
		caseFileID: caseFileID,
		tracking:   tracking,
	}
	for _, opt := range opts {
		opt(&cmd)
	}
	return cmd, nil
}

type CommandOption func(*Command)

func WithDeltaInput(d *casefile.DeltaInput) CommandOption {
	return func(c *Command) { c.deltaInput = d }
}

func (c Command) TenantID() string                 { return c.tenantID }
func (c Command) CaseFileID() string               { return c.caseFileID }
func (c Command) Tracking() *evidencedto.Tracking  { return c.tracking }
func (c Command) DeltaInput() *casefile.DeltaInput { return c.deltaInput }
