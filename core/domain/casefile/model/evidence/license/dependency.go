package license

import (
	"fmt"
	"strings"
)

type Dependency struct {
	name    string
	version string
	license string
}

func NewDependency(name, version, license string) (Dependency, error) {
	if strings.TrimSpace(name) == "" {
		return Dependency{}, fmt.Errorf("%w: name must not be empty", ErrInvalidDependency)
	}
	if strings.TrimSpace(version) == "" {
		return Dependency{}, fmt.Errorf("%w: version must not be empty", ErrInvalidDependency)
	}
	if strings.TrimSpace(license) == "" {
		return Dependency{}, fmt.Errorf("%w: license must not be empty", ErrInvalidDependency)
	}

	return Dependency{
		name:    name,
		version: version,
		license: license,
	}, nil
}

func (dl Dependency) Name() string    { return dl.name }
func (dl Dependency) Version() string { return dl.version }
func (dl Dependency) License() string { return dl.license }
