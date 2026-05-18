package archconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

func ParseFile(path string) (archpolicy.Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return archpolicy.Policy{}, fmt.Errorf("%w: %s: %w", ErrReadConfig, path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (archpolicy.Policy, error) {
	var dto configDTO
	if err := yaml.Unmarshal(data, &dto); err != nil {
		return archpolicy.Policy{}, fmt.Errorf("%w: %w", ErrParseConfig, err)
	}
	policy, err := mapToDomain(dto)
	if err != nil {

		return archpolicy.Policy{}, fmt.Errorf("%w: %w", ErrParseConfig, err)
	}
	return policy, nil
}

func mapToDomain(dto configDTO) (archpolicy.Policy, error) {
	layers, err := buildLayers(dto.Layers)
	if err != nil {
		return archpolicy.Policy{}, err
	}

	rules, err := buildRules(dto.Rules)
	if err != nil {
		return archpolicy.Policy{}, err
	}

	detectCycles := dto.DetectCycles || (dto.Generic != nil && dto.Generic.NoCircularDeps)

	return archpolicy.NewPolicy(layers, rules, detectCycles)
}

func buildLayers(raw map[string][]string) ([]archpolicy.Layer, error) {
	layers := make([]archpolicy.Layer, 0, len(raw))
	for name, patterns := range raw {
		layer, err := archpolicy.NewLayer(name, patterns)
		if err != nil {
			return nil, fmt.Errorf("layer %q: %w", name, err)
		}
		layers = append(layers, layer)
	}
	return layers, nil
}

func buildRules(dtos []ruleDTO) ([]archpolicy.DenyRule, error) {
	rules := make([]archpolicy.DenyRule, 0, len(dtos))
	for i, dto := range dtos {
		rule, err := archpolicy.NewDenyRule(dto.Name, dto.Source, dto.Deny)
		if err != nil {
			return nil, fmt.Errorf("rules[%d]: %w", i, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
