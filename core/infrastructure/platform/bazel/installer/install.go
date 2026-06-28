package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// selfModuleRe matches `module(name = "gavel"...)` or `"gavel_tools"`,
// whether the call is on a single line or split across multiple indented
// lines. Those modules ship the aspects themselves, so a naive
// strings.Contains check would miss the multi-line form and inject a
// self-referential bazel_dep(name = "gavel_tools", ...) that bazel rejects.
const (
	dirPermission  = 0o755
	filePermission = 0o644
)

var selfModuleRe = regexp.MustCompile(`module\s*\(\s*name\s*=\s*"gavel(_tools)?"`)

func (i *Installer) Install(workspace string, tooling []string) (map[string]bool, error) {
	modified := make(map[string]bool)

	if err := os.MkdirAll(filepath.Join(workspace, gavelDir), dirPermission); err != nil {
		return nil, fmt.Errorf("create %s directory: %w", gavelDir, err)
	}

	bazelrcMod, err := i.installBazelrc(workspace, tooling)
	if err != nil {
		return nil, fmt.Errorf("install bazelrc: %w", err)
	}
	for k, v := range bazelrcMod {
		modified[k] = v
	}

	moduleMod, err := i.installModule(workspace)
	if err != nil {
		return nil, fmt.Errorf("install module: %w", err)
	}
	for k, v := range moduleMod {
		modified[k] = v
	}

	buildChanged, err := installLintConfigFilegroup(workspace)
	if err != nil {
		return nil, fmt.Errorf("install lint config filegroup: %w", err)
	}
	if buildChanged {
		modified["BUILD.bazel"] = true
	}

	return modified, nil
}

func (i *Installer) installBazelrc(root string, tooling []string) (map[string]bool, error) {
	block, err := renderTemplate(bazelrcTmpl, toBazelrcData(tooling))
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, GavelBazelrc), []byte(block), filePermission); err != nil {
		return nil, fmt.Errorf("write %s: %w", GavelBazelrc, err)
	}

	registryChanged, err := ensureRegistryLines(root)
	if err != nil {
		return nil, err
	}

	includeChanged, err := ensureIncludeLine(root, ".bazelrc", bazelrcInclude)
	if err != nil {
		return nil, err
	}

	return map[string]bool{
		GavelBazelrc: true,
		".bazelrc":   registryChanged || includeChanged,
	}, nil
}

func (i *Installer) installModule(root string) (map[string]bool, error) {
	block, err := renderTemplate(moduleTmpl, moduleData{})
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, GavelModule), []byte(block), filePermission); err != nil {
		return nil, fmt.Errorf("write %s: %w", GavelModule, err)
	}

	moduleChanged := false
	if !isSelfModule(root) {
		changed, err := ensureIncludeLine(root, "MODULE.bazel", gavelDepLine)
		if err != nil {
			return nil, err
		}
		moduleChanged = changed
	}

	changed, err := ensureIncludeLine(root, "MODULE.bazel", moduleInclude)
	if err != nil {
		return nil, err
	}
	moduleChanged = moduleChanged || changed

	return map[string]bool{
		GavelModule:    true,
		"MODULE.bazel": moduleChanged,
	}, nil
}

func isSelfModule(root string) bool {
	data, err := os.ReadFile(filepath.Join(root, "MODULE.bazel"))
	if err != nil {
		return false
	}
	return selfModuleRe.Match(data)
}

const lintConfigFilegroup = `filegroup(
    name = "gavel_lint_config",
    srcs = glob(
        [
            ".golangci.yml",
            ".golangci.yaml",
            "ruff.toml",
            "pyproject.toml",
            ".bandit",
            "clippy.toml",
            ".clippy.toml",
            "eslint.config.*",
            ".eslintrc.*",
            ".gavel/architecture.yml",
        ],
        allow_empty = True,
    ),
    visibility = ["//visibility:public"],
)`

func installLintConfigFilegroup(root string) (bool, error) {
	path := filepath.Join(root, "BUILD.bazel")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read BUILD.bazel: %w", err)
	}

	if strings.Contains(string(existing), "gavel_lint_config") {
		return false, nil
	}

	var builder strings.Builder
	content := string(existing)
	if content != "" {
		builder.WriteString(strings.TrimRight(content, "\n"))
		builder.WriteString("\n\n")
	}
	builder.WriteString(lintConfigFilegroup)
	builder.WriteString("\n")

	return true, os.WriteFile(path, []byte(builder.String()), filePermission)
}

const (
	gavelRegistryLine = "common --registry=https://gavelcode.github.io/registry"
	bcrRegistryLine   = "common --registry=https://bcr.bazel.build"
)

// Bazel drops its implicit BCR default as soon as any --registry is declared,
// so the gavel registry and BCR must both be listed explicitly.
func ensureRegistryLines(root string) (bool, error) {
	path := filepath.Join(root, ".bazelrc")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read .bazelrc: %w", err)
	}
	content := string(existing)

	missing := make([]string, 0, 2)
	if !strings.Contains(content, gavelRegistryLine) {
		missing = append(missing, gavelRegistryLine)
	}
	if !strings.Contains(content, bcrRegistryLine) {
		missing = append(missing, bcrRegistryLine)
	}
	if len(missing) == 0 {
		return false, nil
	}

	var builder strings.Builder
	if content != "" {
		builder.WriteString(strings.TrimRight(content, "\n"))
		builder.WriteString("\n")
	}
	for _, line := range missing {
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return true, os.WriteFile(path, []byte(builder.String()), filePermission)
}

func ensureIncludeLine(root, filename, line string) (bool, error) {
	path := filepath.Join(root, filename)
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", filename, err)
	}

	content := string(existing)
	if strings.Contains(content, line) {
		return false, nil
	}

	var builder strings.Builder
	if content != "" {
		builder.WriteString(strings.TrimRight(content, "\n"))
		builder.WriteString("\n\n")
	}
	builder.WriteString(line)
	builder.WriteString("\n")

	return true, os.WriteFile(path, []byte(builder.String()), filePermission)
}
