package archpolicy

import "fmt"

type Policy struct {
	layers       []Layer
	denyRules    []DenyRule
	detectCycles bool
}

func NewPolicy(layers []Layer, denyRules []DenyRule, detectCycles bool) (Policy, error) {
	if len(layers) == 0 {
		return Policy{}, fmt.Errorf("%w: at least one layer required", ErrInvalidPolicy)
	}
	if err := validateNoDuplicateLayerNames(layers); err != nil {
		return Policy{}, err
	}
	layerNames := buildLayerNameSet(layers)
	if err := validateRulesReferenceExistingLayers(denyRules, layerNames); err != nil {
		return Policy{}, err
	}

	copiedLayers := make([]Layer, len(layers))
	copy(copiedLayers, layers)
	copiedRules := make([]DenyRule, len(denyRules))
	copy(copiedRules, denyRules)

	return Policy{
		layers:       copiedLayers,
		denyRules:    copiedRules,
		detectCycles: detectCycles,
	}, nil
}

func (p Policy) Layers() []Layer {
	copied := make([]Layer, len(p.layers))
	copy(copied, p.layers)
	return copied
}

func (p Policy) DenyRules() []DenyRule {
	copied := make([]DenyRule, len(p.denyRules))
	copy(copied, p.denyRules)
	return copied
}

func (p Policy) DetectCycles() bool {
	return p.detectCycles
}

func validateNoDuplicateLayerNames(layers []Layer) error {
	seen := make(map[string]bool, len(layers))
	for _, l := range layers {
		if seen[l.name] {
			return fmt.Errorf("%w: duplicate layer name %q", ErrInvalidPolicy, l.name)
		}
		seen[l.name] = true
	}
	return nil
}

func buildLayerNameSet(layers []Layer) map[string]bool {
	set := make(map[string]bool, len(layers))
	for _, l := range layers {
		set[l.name] = true
	}
	return set
}

func validateRulesReferenceExistingLayers(rules []DenyRule, layerNames map[string]bool) error {
	for _, denyRule := range rules {
		if !layerNames[denyRule.source] {
			return fmt.Errorf("%w: rule %q references unknown source layer %q", ErrInvalidPolicy, denyRule.name, denyRule.source)
		}
		for _, d := range denyRule.deny {
			if !layerNames[d] {
				return fmt.Errorf("%w: rule %q denies unknown layer %q", ErrInvalidPolicy, denyRule.name, d)
			}
			if d == denyRule.source {
				return fmt.Errorf("%w: rule %q denies its own source layer %q", ErrInvalidPolicy, denyRule.name, d)
			}
		}
	}
	return nil
}
