package ingestevidence

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type Command struct {
	caseFileID string
	evidences  []evidencedto.Evidence
}

func NewCommand(caseFileID string, evidences []evidencedto.Evidence) (Command, error) {
	if strings.TrimSpace(caseFileID) == "" {
		return Command{}, fmt.Errorf("%w: caseFileID must not be empty", ErrInvalidCommand)
	}
	if len(evidences) == 0 {
		return Command{}, fmt.Errorf("%w: at least one evidence is required", ErrInvalidCommand)
	}
	copied := make([]evidencedto.Evidence, len(evidences))
	copy(copied, evidences)
	return Command{caseFileID: caseFileID, evidences: copied}, nil
}

func (c Command) CaseFileID() string { return c.caseFileID }

func (c Command) Evidences() []evidencedto.Evidence {
	copied := make([]evidencedto.Evidence, len(c.evidences))
	copy(copied, c.evidences)
	return copied
}
