package file

import (
	"fmt"
	"strings"
)

type Command struct {
	projectID    string
	number       int
	title        string
	petitioner   string
	sourceBranch string
	targetBranch string
	commitSHA    string
}

func NewCommand(projectID string, number int, title, petitioner, sourceBranch, targetBranch, commitSHA string) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if number <= 0 {
		return Command{}, fmt.Errorf("%w: number must be positive", ErrInvalidCommand)
	}
	if strings.TrimSpace(title) == "" {
		return Command{}, fmt.Errorf("%w: title must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(sourceBranch) == "" {
		return Command{}, fmt.Errorf("%w: sourceBranch must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetBranch) == "" {
		return Command{}, fmt.Errorf("%w: targetBranch must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(commitSHA) == "" {
		return Command{}, fmt.Errorf("%w: commitSHA must not be empty", ErrInvalidCommand)
	}
	return Command{
		projectID:    projectID,
		number:       number,
		title:        title,
		petitioner:   petitioner,
		sourceBranch: sourceBranch,
		targetBranch: targetBranch,
		commitSHA:    commitSHA,
	}, nil
}

func (c Command) ProjectID() string    { return c.projectID }
func (c Command) Number() int          { return c.number }
func (c Command) Title() string        { return c.title }
func (c Command) Petitioner() string   { return c.petitioner }
func (c Command) SourceBranch() string { return c.sourceBranch }
func (c Command) TargetBranch() string { return c.targetBranch }
func (c Command) CommitSHA() string    { return c.commitSHA }
