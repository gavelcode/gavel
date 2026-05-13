package judge

import (
	"fmt"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
)

type Command struct {
	caseFileID string
	tracking   *evidencedto.Tracking
	deltaInput *casefile.DeltaInput
}

func NewCommand(caseFileID string, tracking *evidencedto.Tracking, opts ...CommandOption) (Command, error) {
	if caseFileID == "" {
		return Command{}, fmt.Errorf("%w: caseFileID must not be empty", ErrInvalidCommand)
	}
	cmd := Command{
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

func (c Command) CaseFileID() string                { return c.caseFileID }
func (c Command) Tracking() *evidencedto.Tracking   { return c.tracking }
func (c Command) DeltaInput() *casefile.DeltaInput { return c.deltaInput }
