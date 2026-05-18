package archconfig

import (
	"path/filepath"

	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

type PolicyLoader struct{}

func NewPolicyLoader() *PolicyLoader {
	return &PolicyLoader{}
}

func (l *PolicyLoader) LoadPolicy(workspace string) (archpolicy.Policy, error) {
	configPath := filepath.Join(workspace, ".gavel", "architecture.yml")
	return ParseFile(configPath)
}
