package createcasefile

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	projectID       string
	commitSHA       string
	branch          string
	startedAt       time.Time
	freshEvaluation bool
}

func NewCommand(projectID, commitSHA, branch string, startedAt time.Time, opts ...Option) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(commitSHA) == "" {
		return Command{}, fmt.Errorf("%w: commitSHA must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(branch) == "" {
		return Command{}, fmt.Errorf("%w: branch must not be empty", ErrInvalidCommand)
	}
	if startedAt.IsZero() {
		return Command{}, fmt.Errorf("%w: startedAt must not be zero", ErrInvalidCommand)
	}
	cmd := Command{
		projectID: projectID,
		commitSHA: commitSHA,
		branch:    branch,
		startedAt: startedAt,
	}
	for _, opt := range opts {
		opt(&cmd)
	}
	return cmd, nil
}

func (c Command) ProjectID() string     { return c.projectID }
func (c Command) CommitSHA() string     { return c.commitSHA }
func (c Command) Branch() string        { return c.branch }
func (c Command) StartedAt() time.Time  { return c.startedAt }
func (c Command) FreshEvaluation() bool { return c.freshEvaluation }
