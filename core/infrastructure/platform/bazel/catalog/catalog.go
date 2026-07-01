package catalog

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

type Tool struct {
	Name        string
	Aspect      string
	SARIFSuffix string
	BuildFlags  []string
	Binary      string
}

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

func (c *Catalog) aspects(languages []string, keep func(Tool) bool) []Aspect {
	seen := make(map[string]bool)
	var selected []Aspect
	for _, language := range languages {
		for _, tool := range c.languages[language] {
			if seen[tool.Aspect] || !keep(tool) {
				continue
			}
			seen[tool.Aspect] = true
			selected = append(selected, c.aspectFor(tool))
		}
	}
	return selected
}

func (c *Catalog) aspectFor(tool Tool) Aspect {
	return Aspect{
		Name:        tool.Aspect,
		Path:        modulePrefix + c.aspectsBzl + "%" + tool.Aspect,
		SARIFSuffix: tool.SARIFSuffix,
		BuildFlags:  tool.BuildFlags,
	}
}

func (c *Catalog) selectedAspects(selection map[string][]string) ([]Aspect, error) {
	languages := make([]string, 0, len(selection))
	for language := range selection {
		languages = append(languages, language)
	}
	sort.Strings(languages)

	selected := make([]Aspect, 0)
	for _, language := range languages {
		byName := make(map[string]Tool, len(c.languages[language]))
		for _, tool := range c.languages[language] {
			byName[tool.Name] = tool
		}
		for _, toolName := range selection[language] {
			tool, exists := byName[toolName]
			if !exists {
				return nil, fmt.Errorf("language %q: unknown tool %q", language, toolName)
			}
			selected = append(selected, c.aspectFor(tool))
		}
	}
	return selected, nil
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
