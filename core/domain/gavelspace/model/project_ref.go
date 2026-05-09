package model

import (
	"fmt"
	"strings"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type ProjectRef struct {
	id            projectmodel.ProjectID
	targetPattern string
}

func NewProjectRef(id projectmodel.ProjectID, targetPattern string) (ProjectRef, error) {
	if strings.TrimSpace(targetPattern) == "" {
		return ProjectRef{}, fmt.Errorf("%w: target pattern must not be empty", ErrInvalidGavelspace)
	}
	return ProjectRef{id: id, targetPattern: targetPattern}, nil
}

func (r ProjectRef) ID() projectmodel.ProjectID { return r.id }
func (r ProjectRef) TargetPattern() string      { return r.targetPattern }
