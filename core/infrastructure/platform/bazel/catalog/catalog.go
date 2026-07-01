package catalog

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Tool is one linter gavel-tools can run for a language, as published in its
// catalog. It is the identity gavel used to hardcode: which aspect runs it, what
// SARIF file it emits, the build flags it needs, and the tool-binary repo it
// depends on (both optional).
type Tool struct {
	Name        string
	Aspect      string
	SARIFSuffix string
	BuildFlags  []string
	Binary      string
}

// Catalog is the parsed gavel-tools menu: the tools available per language, plus
// the module-relative label where the aspects live.
type Catalog struct {
	aspectsBzl string
	languages  map[string][]Tool
}

type catalogDTO struct {
	Version    int                  `yaml:"version"`
	AspectsBzl string               `yaml:"aspects_bzl"`
	Languages  map[string][]toolDTO `yaml:"languages"`
}

type toolDTO struct {
	Name        string   `yaml:"name"`
	Aspect      string   `yaml:"aspect"`
	SARIFSuffix string   `yaml:"sarif_suffix"`
	BuildFlags  []string `yaml:"build_flags"`
	Binary      string   `yaml:"binary"`
}

// ParseCatalog decodes the gavel-tools catalog.yaml. It is pure so the selection
// logic is fully testable without runfiles, and lets a caller load a catalog
// from a source other than the default runfiles location.
func ParseCatalog(data []byte) (*Catalog, error) {
	var dto catalogDTO
	if err := yaml.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	if dto.AspectsBzl == "" {
		return nil, fmt.Errorf("parse catalog: missing aspects_bzl")
	}
	if len(dto.Languages) == 0 {
		return nil, fmt.Errorf("parse catalog: no languages")
	}

	languages := make(map[string][]Tool, len(dto.Languages))
	for language, tools := range dto.Languages {
		converted := make([]Tool, 0, len(tools))
		for _, tool := range tools {
			if tool.Aspect == "" {
				return nil, fmt.Errorf("parse catalog: tool %q in %q has no aspect", tool.Name, language)
			}
			converted = append(converted, Tool(tool))
		}
		languages[language] = converted
	}
	return &Catalog{aspectsBzl: dto.AspectsBzl, languages: languages}, nil
}

// aspects selects, in the caller's language order and de-duplicated, the aspects
// whose tool passes keep, resolving each to a full aspect path.
func (c *Catalog) aspects(languages []string, keep func(Tool) bool) []Aspect {
	seen := make(map[string]bool)
	var selected []Aspect
	for _, language := range languages {
		for _, tool := range c.languages[language] {
			if seen[tool.Aspect] || !keep(tool) {
				continue
			}
			seen[tool.Aspect] = true
			selected = append(selected, Aspect{
				Name:        tool.Aspect,
				Path:        modulePrefix + c.aspectsBzl + "%" + tool.Aspect,
				SARIFSuffix: tool.SARIFSuffix,
				BuildFlags:  tool.BuildFlags,
			})
		}
	}
	return selected
}

func (c *Catalog) aspectNames(languages []string) []string {
	var names []string
	for _, language := range languages {
		for _, tool := range c.languages[language] {
			names = append(names, tool.Aspect)
		}
	}
	return names
}

func (c *Catalog) binaryNames(languages []string) []string {
	seen := make(map[string]bool)
	var names []string
	for _, language := range languages {
		for _, tool := range c.languages[language] {
			if tool.Binary == "" || seen[tool.Binary] {
				continue
			}
			seen[tool.Binary] = true
			names = append(names, tool.Binary)
		}
	}
	return names
}
