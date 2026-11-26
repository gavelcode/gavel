package catalog

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// CoreModule is the module that publishes a tool's aspect when the catalog does not say otherwise.
const CoreModule = "gavel_tools"

type Tool struct {
	Name string
	// Module and AspectsBzl are the effective values: the tool's own, else the defaults.
	Module      string
	AspectsBzl  string
	Aspect      string
	SARIFSuffix string
	BuildFlags  []string
	Binary      string
	Variants    []Variant
}

// Variant is a ruleset-specific build of a tool, chosen when the consumer's
// root module already depends on WhenBazelDep. Gavel never imposes a ruleset.
type Variant struct {
	WhenBazelDep string
	Module       string
	AspectsBzl   string
	Aspect       string
	BuildFlags   []string
}

// Deps is the set of bazel_dep module names the consumer's root module declares.
type Deps map[string]bool

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
	Name        string       `yaml:"name"`
	Module      string       `yaml:"module"`
	AspectsBzl  string       `yaml:"aspects_bzl"`
	Aspect      string       `yaml:"aspect"`
	SARIFSuffix string       `yaml:"sarif_suffix"`
	BuildFlags  []string     `yaml:"build_flags"`
	Binary      string       `yaml:"binary"`
	Variants    []variantDTO `yaml:"variants"`
}

type variantDTO struct {
	WhenBazelDep string   `yaml:"when_bazel_dep"`
	Module       string   `yaml:"module"`
	AspectsBzl   string   `yaml:"aspects_bzl"`
	Aspect       string   `yaml:"aspect"`
	BuildFlags   []string `yaml:"build_flags"`
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
			variants, err := convertVariants(language, tool)
			if err != nil {
				return nil, err
			}
			converted = append(converted, Tool{
				Name:        tool.Name,
				Module:      moduleOrCore(tool.Module),
				AspectsBzl:  firstNonEmpty(tool.AspectsBzl, dto.AspectsBzl),
				Aspect:      tool.Aspect,
				SARIFSuffix: tool.SARIFSuffix,
				BuildFlags:  tool.BuildFlags,
				Binary:      tool.Binary,
				Variants:    variants,
			})
		}
		languages[language] = converted
	}
	return &Catalog{aspectsBzl: dto.AspectsBzl, languages: languages}, nil
}

// Explicit, never implicit: a variant that does not say what it is gets rejected rather than guessed.
func convertVariants(language string, tool toolDTO) ([]Variant, error) {
	variants := make([]Variant, 0, len(tool.Variants))
	for index, variant := range tool.Variants {
		for field, value := range map[string]string{
			"when_bazel_dep": variant.WhenBazelDep,
			"module":         variant.Module,
			"aspect":         variant.Aspect,
		} {
			if value == "" {
				return nil, fmt.Errorf("parse catalog: variant %d of tool %q in %q has no %s", index, tool.Name, language, field)
			}
		}
		variants = append(variants, Variant(variant))
	}
	return variants, nil
}

func moduleOrCore(module string) string {
	return firstNonEmpty(module, CoreModule)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (c *Catalog) aspects(languages []string, deps Deps, keep func(Tool) bool) []Aspect {
	seen := make(map[string]bool)
	var selected []Aspect
	for _, language := range languages {
		for _, tool := range c.languages[language] {
			if seen[tool.Aspect] || !keep(tool) {
				continue
			}
			seen[tool.Aspect] = true
			selected = append(selected, c.aspectFor(tool, deps))
		}
	}
	return selected
}

// aspectFor resolves the effective definition of a tool: the first variant whose
// when_bazel_dep the consumer satisfies, else the base entry. The label is
// @<module><aspects_bzl>%<aspect>.
func (c *Catalog) aspectFor(tool Tool, deps Deps) Aspect {
	for _, variant := range tool.Variants {
		if !deps[variant.WhenBazelDep] {
			continue
		}
		flags := tool.BuildFlags
		if variant.BuildFlags != nil {
			flags = variant.BuildFlags
		}
		return newAspect(variant.Module, firstNonEmpty(variant.AspectsBzl, tool.AspectsBzl), variant.Aspect, tool.SARIFSuffix, flags)
	}
	return newAspect(tool.Module, tool.AspectsBzl, tool.Aspect, tool.SARIFSuffix, tool.BuildFlags)
}

func newAspect(module, aspectsBzl, aspect, sarifSuffix string, flags []string) Aspect {
	return Aspect{
		Name:        aspect,
		Path:        "@" + module + aspectsBzl + "%" + aspect,
		SARIFSuffix: sarifSuffix,
		BuildFlags:  flags,
		Module:      module,
	}
}

func (c *Catalog) selectedAspects(selection map[string][]string, deps Deps) ([]Aspect, error) {
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
			selected = append(selected, c.aspectFor(tool, deps))
		}
	}
	return selected, nil
}

func (c *Catalog) aspectForTool(language, name string, deps Deps) (Aspect, bool) {
	for _, tool := range c.languages[language] {
		if tool.Name == name {
			return c.aspectFor(tool, deps), true
		}
	}
	return Aspect{}, false
}

func (c *Catalog) toolNames(language string) []string {
	names := make([]string, 0, len(c.languages[language]))
	for _, tool := range c.languages[language] {
		names = append(names, tool.Name)
	}
	return names
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
