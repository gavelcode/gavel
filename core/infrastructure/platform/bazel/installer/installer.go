package installer

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"text/template"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

//go:embed bazelrc.tmpl
var bazelrcTmplStr string

//go:embed module.bazel.tmpl
var moduleTmplStr string

var bazelrcTmpl = template.Must(template.New("bazelrc").Parse(bazelrcTmplStr))
var moduleTmpl = template.Must(template.New("module.bazel").Parse(moduleTmplStr))

const gavelDir = ".gavel"
const bazelrcInclude = "try-import %workspace%/.gavel/gavel.bazelrc"
const moduleInclude = `include("//:.gavel/gavel.MODULE.bazel")`

const GavelBazelrc = ".gavel/gavel.bazelrc"
const GavelModule = ".gavel/gavel.MODULE.bazel"
const gavelToolsVersion = "0.3.8"
const gavelDepLine = `bazel_dep(name = "gavel_tools", version = "` + gavelToolsVersion + `")`

type Installer struct{}

func NewInstaller() *Installer {
	return &Installer{}
}

func (i *Installer) Catalog(tooling []string) ([]string, []string) {
	return catalog.AspectNames(tooling), catalog.BinaryNames(tooling)
}

func (i *Installer) ToolsForLanguage(language string) []string {
	return catalog.ToolNamesForLanguage(language)
}

type bazelrcData struct {
	Go         bool
	Java       bool
	Python     bool
	TypeScript bool
	Rust       bool
	deps       catalog.Deps
}

// Label is the aspect label for one catalog tool, resolved as judge resolves it.
func (d bazelrcData) Label(language, tool string) (string, error) {
	aspect, ok := catalog.AspectFor(language, tool, d.deps)
	if !ok {
		return "", fmt.Errorf("catalog has no %s tool %q", language, tool)
	}
	return aspect.Path, nil
}

// Flags are the extra build flags the resolved tool needs.
func (d bazelrcData) Flags(language, tool string) []string {
	aspect, _ := catalog.AspectFor(language, tool, d.deps)
	return aspect.BuildFlags
}

type moduleData struct {
	Version string
	// Modules are the non-core modules the selected tools resolved to.
	Modules []string
}

func toBazelrcData(tooling []string, deps catalog.Deps) bazelrcData {
	set := toToolingSet(tooling)
	return bazelrcData{
		Go:         set["go"],
		Java:       set["java"],
		Python:     set["python"],
		TypeScript: set["typescript"],
		Rust:       set["rust"],
		deps:       deps,
	}
}

// The core module's dep is written once into the root MODULE.bazel (gavelDepLine), so it is not repeated here.
func toModuleData(tooling []string, deps catalog.Deps) moduleData {
	seen := make(map[string]bool)
	var modules []string
	for _, aspect := range catalog.AllAspectsForLanguages(tooling, deps) {
		if aspect.Module == catalog.CoreModule || seen[aspect.Module] {
			continue
		}
		seen[aspect.Module] = true
		modules = append(modules, aspect.Module)
	}
	sort.Strings(modules)
	return moduleData{Version: gavelToolsVersion, Modules: modules}
}

func toToolingSet(tooling []string) map[string]bool {
	set := make(map[string]bool, len(tooling))
	for _, t := range tooling {
		set[t] = true
	}
	return set
}

func renderTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", tmpl.Name(), err)
	}
	return buf.String(), nil
}
