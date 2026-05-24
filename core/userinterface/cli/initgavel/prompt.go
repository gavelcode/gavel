package initgavel

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

func promptProjectName(defaultName string) (string, error) {
	name := defaultName
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Identifies this workspace in Gavel reports and the dashboard.").
				Value(&name).
				Placeholder(defaultName).
				Validate(validateNotBlank("project name")),
		),
	).WithShowHelp(true).Run()
	if err != nil {
		return "", err
	}
	if name == "" {
		name = defaultName
	}
	return name, nil
}

func promptProjects() ([]Project, error) {
	var projects []Project

	for i := 0; ; i++ {
		proj, err := promptProject(i)
		if err != nil {
			return nil, err
		}
		projects = append(projects, Project{
			Name:    proj.name,
			Pattern: proj.pattern,
			Tooling: proj.tooling,
		})

		more, err := promptAddAnother()
		if err != nil {
			return nil, err
		}
		if !more {
			break
		}
	}

	return projects, nil
}

type promptedProject struct {
	name    string
	pattern string
	tooling []string
}

func promptProject(index int) (promptedProject, error) {
	name := ""
	pattern := ""

	var tooling []string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("A label for this project (e.g., backend, frontend, api).").
				Value(&name).
				Placeholder("backend").
				Validate(validateNotBlank("project name")),
			huh.NewInput().
				Title("Bazel pattern").
				Description("The Bazel target pattern to analyze (e.g., //server/..., //web/...).").
				Value(&pattern).
				Placeholder("//server/...").
				Validate(validateNotBlank("Bazel pattern")),
		).Title(fmt.Sprintf("Project %d", index+1)),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("Languages for project %d", index+1)).
				Description("Gavel will install aspects and tool binaries for each selected language.\nUse <space> to toggle, <enter> to confirm.").
				Options(
					huh.NewOption("Go (golangci-lint)", "go"),
					huh.NewOption("Java (PMD, CPD, SpotBugs, Error Prone)", "java"),
					huh.NewOption("Python (pycompile, Ruff, Bandit)", "python"),
					huh.NewOption("TypeScript (ESLint)", "typescript"),
					huh.NewOption("Rust (Clippy)", "rust"),
				).
				Value(&tooling).
				Validate(func(selected []string) error {
					if len(selected) == 0 {
						return fmt.Errorf("select at least one language")
					}
					return nil
				}),
		),
	).WithShowHelp(true).Run()
	if err != nil {
		return promptedProject{}, err
	}

	return promptedProject{name: name, pattern: pattern, tooling: tooling}, nil
}

func promptAddAnother() (bool, error) {
	addMore := false
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add another project?").
				Value(&addMore),
		),
	).WithShowHelp(true).Run()
	if err != nil {
		return false, err
	}
	return addMore, nil
}

func validateNotBlank(fieldName string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		return nil
	}
}
