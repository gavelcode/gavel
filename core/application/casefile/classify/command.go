package classify

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type Command struct {
	projectID string
	branch    string
	findings  []finding.Finding
}

func NewCommand(projectID, branch string, findings []finding.Finding) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(branch) == "" {
		return Command{}, fmt.Errorf("%w: branch must not be empty", ErrInvalidCommand)
	}
	copied := make([]finding.Finding, len(findings))
	copy(copied, findings)
	return Command{
		projectID: projectID,
		branch:    branch,
		findings:  copied,
	}, nil
}

func (c Command) ProjectID() string { return c.projectID }
func (c Command) Branch() string    { return c.branch }

func (c Command) Findings() []finding.Finding {
	copied := make([]finding.Finding, len(c.findings))
	copy(copied, c.findings)
	return copied
}
