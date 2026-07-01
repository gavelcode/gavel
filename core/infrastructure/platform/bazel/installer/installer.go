package installer

import (
	"bytes"
	_ "embed"
	"fmt"
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
const gavelToolsVersion = "0.3.1"
const gavelDepLine = `bazel_dep(name = "gavel_tools", version = "` + gavelToolsVersion + `")`

type Installer struct{}

func NewInstaller() *Installer {
	return &Installer{}
}

func (i *Installer) Catalog(tooling []string) ([]string, []string) {
	return catalog.AspectNames(tooling), catalog.BinaryNames(tooling)
}

type bazelrcData struct {
	Prefix     string
	Go         bool
	Java       bool
	Python     bool
	TypeScript bool
	Rust       bool
}

type moduleData struct{}

func toBazelrcData(tooling []string) bazelrcData {
	set := toToolingSet(tooling)
	return bazelrcData{
		Prefix:     catalog.ModulePrefix() + "//lint/aspects:defs.bzl%",
		Go:         set["go"],
		Java:       set["java"],
		Python:     set["python"],
		TypeScript: set["typescript"],
		Rust:       set["rust"],
	}
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
