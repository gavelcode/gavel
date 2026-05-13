package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
)

type License struct {
	Dependencies []Dependency
}

func fromDomainLicense(c license.Content) License {
	deps := c.Dependencies()
	out := make([]Dependency, 0, len(deps))
	for _, d := range deps {
		out = append(out, Dependency{
			Name:    d.Name(),
			Version: d.Version(),
			License: d.License(),
		})
	}
	return License{Dependencies: out}
}

func toDomainLicense(in License) (license.Content, error) {
	deps := make([]license.Dependency, 0, len(in.Dependencies))
	for i, d := range in.Dependencies {
		dl, err := license.NewDependency(d.Name, d.Version, d.License)
		if err != nil {
			return license.Content{}, fmt.Errorf("dependencies[%d]: %w", i, err)
		}
		deps = append(deps, dl)
	}
	return license.NewContent(deps)
}
