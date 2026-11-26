package model

import (
	"fmt"
	"regexp"
	"strings"
)

const sarifExtension = ".sarif"

var bridgeNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Bridge is a linter aspect the project provides itself. Gavel builds it and
// judges its SARIF like a catalog tool, but does not guarantee it: the project
// owns the aspect, its hermeticity and its SARIF contract.
type Bridge struct {
	name        string
	aspect      string
	sarifSuffix string
	buildFlags  []string
}

func NewBridge(name, aspect, sarifSuffix string, buildFlags []string) (Bridge, error) {
	if err := validateBridgeName(name); err != nil {
		return Bridge{}, err
	}
	if err := validateBridgeAspect(aspect); err != nil {
		return Bridge{}, err
	}
	if err := validateBridgeSuffix(sarifSuffix); err != nil {
		return Bridge{}, err
	}
	for _, flag := range buildFlags {
		if !strings.HasPrefix(flag, "-") {
			return Bridge{}, fmt.Errorf("%w: build_flags entry %q must be a bazel flag starting with '-'", ErrInvalidBridge, flag)
		}
	}
	return Bridge{
		name:        name,
		aspect:      aspect,
		sarifSuffix: sarifSuffix,
		buildFlags:  append([]string(nil), buildFlags...),
	}, nil
}

func validateBridgeName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidBridge)
	}
	if !bridgeNameRegex.MatchString(name) {
		return fmt.Errorf("%w: name %q must be letters, digits, '.', '_' or '-' and start with a letter or digit", ErrInvalidBridge, name)
	}
	return nil
}

func validateBridgeAspect(aspect string) error {
	if strings.TrimSpace(aspect) == "" {
		return fmt.Errorf("%w: aspect must not be empty", ErrInvalidBridge)
	}
	file, symbol, found := strings.Cut(aspect, "%")
	if !found || file == "" || symbol == "" || strings.Contains(symbol, "%") {
		return fmt.Errorf("%w: aspect %q must be a label of the form <file label>%%<aspect name>, e.g. //tools/lint:shellcheck.bzl%%shellcheck_gavel", ErrInvalidBridge, aspect)
	}
	return nil
}

func validateBridgeSuffix(suffix string) error {
	if strings.TrimSpace(suffix) == "" {
		return fmt.Errorf("%w: sarif_suffix must not be empty", ErrInvalidBridge)
	}
	if !strings.HasSuffix(suffix, sarifExtension) || suffix == sarifExtension {
		return fmt.Errorf("%w: sarif_suffix %q must end in %s and name the bridge (e.g. .shellcheck.sarif)", ErrInvalidBridge, suffix, sarifExtension)
	}
	return nil
}

func (b Bridge) Name() string        { return b.name }
func (b Bridge) Aspect() string      { return b.aspect }
func (b Bridge) SARIFSuffix() string { return b.sarifSuffix }

func (b Bridge) BuildFlags() []string { return append([]string(nil), b.buildFlags...) }
