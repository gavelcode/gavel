package initgavel

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultCoverageMin = 60

type ConfigInstaller interface {
	Install(workspace string, tooling []string) (map[string]bool, error)
}

type ToolCatalogProvider interface {
	Catalog(tooling []string) (aspects []string, binaries []string)
}

type Project struct {
	Name    string
	Pattern string
	Tooling []string
}

type Server struct {
	URL   string
	Token string
}

type initResult struct {
	configPath string
	created    bool
	name       string
	projects   []Project
	server     Server
	aspects    []string
	binaries   []string
	modified   map[string]bool
}

func execute(configPath, workspacePath, name string, force bool, projects []Project, server Server, sourceFile string, installer ConfigInstaller, catalog ToolCatalogProvider) (initResult, error) {
	absPath := resolveAbsConfigPath(workspacePath, configPath)

	if !force {
		if _, err := os.Stat(absPath); err == nil {
			return initResult{
				configPath: configPath,
				name:       name,
				projects:   projects,
				server:     server,
			}, nil
		}
	}

	if sourceFile != "" {
		if err := copyConfig(sourceFile, absPath); err != nil {
			return initResult{}, fmt.Errorf("copy config: %w", err)
		}
	} else {
		if err := writeConfig(absPath, name, projects, server); err != nil {
			return initResult{}, fmt.Errorf("save config: %w", err)
		}
	}

	tooling := extractTooling(projects)

	if err := writeArchitectureConfig(workspacePath, tooling); err != nil {
		return initResult{}, fmt.Errorf("save architecture config: %w", err)
	}
	modified, err := installer.Install(workspacePath, tooling)
	if err != nil {
		return initResult{}, fmt.Errorf("install config: %w", err)
	}

	aspects, binaries := catalog.Catalog(tooling)

	return initResult{
		configPath: configPath,
		created:    true,
		name:       name,
		projects:   projects,
		server:     server,
		aspects:    aspects,
		binaries:   binaries,
		modified:   modified,
	}, nil
}

func resolveAbsConfigPath(workspace, configPath string) string {
	if filepath.IsAbs(configPath) {
		return configPath
	}
	return filepath.Join(workspace, configPath)
}

func copyConfig(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), dirPermission); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source config: %w", err)
	}
	return os.WriteFile(dst, data, filePermission)
}

func writeConfig(path string, name string, projects []Project, server Server) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPermission); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	dto := buildConfigDTO(name, projects, server)
	data, err := yaml.Marshal(dto)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, filePermission)
}

func defaultQualityGate() *qualityGateDTO {
	return &qualityGateDTO{
		Findings:   &findingsRuleDTO{MaxError: 0},
		Coverage:   &coverageRuleDTO{Min: defaultCoverageMin},
		Violations: &violationsRuleDTO{Max: 0},
	}
}

func buildConfigDTO(name string, projects []Project, server Server) configDTO {
	qGate := defaultQualityGate()
	dtoProjects := make([]projectDTO, 0, len(projects))
	for _, p := range projects {
		dtoProjects = append(dtoProjects, projectDTO{
			Name:    p.Name,
			Pattern: p.Pattern,
			Tooling: p.Tooling,
			Gate:    qGate,
		})
	}

	return configDTO{
		Name:     name,
		Projects: dtoProjects,
		Server:   serverDTO(server),
	}
}

func readFromConfig(path string) (string, []Project, Server, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, Server{}, fmt.Errorf("read %s: %w", path, err)
	}
	var dto configDTO
	if err := yaml.Unmarshal(data, &dto); err != nil {
		return "", nil, Server{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if dto.Name == "" {
		return "", nil, Server{}, fmt.Errorf("%s: name is required", path)
	}
	if len(dto.Projects) == 0 {
		return "", nil, Server{}, fmt.Errorf("%s: at least one project is required", path)
	}

	projects := make([]Project, 0, len(dto.Projects))
	for _, p := range dto.Projects {
		projects = append(projects, Project{
			Name:    p.Name,
			Pattern: p.Pattern,
			Tooling: p.Tooling,
		})
	}

	server := Server{URL: dto.Server.URL, Token: dto.Server.Token}
	return dto.Name, projects, server, nil
}

func extractTooling(projects []Project) []string {
	seen := make(map[string]bool)
	var result []string
	for _, p := range projects {
		for _, lang := range p.Tooling {
			if !seen[lang] {
				seen[lang] = true
				result = append(result, lang)
			}
		}
	}
	return result
}

type configDTO struct {
	Name     string       `yaml:"name"`
	Projects []projectDTO `yaml:"projects"`
	Server   serverDTO    `yaml:"server,omitempty"`
}

type projectDTO struct {
	Name    string          `yaml:"name"`
	Pattern string          `yaml:"pattern"`
	Tooling []string        `yaml:"tooling"`
	Gate    *qualityGateDTO `yaml:"quality_gate,omitempty"`
}

type qualityGateDTO struct {
	Findings   *findingsRuleDTO   `yaml:"findings,omitempty"`
	Coverage   *coverageRuleDTO   `yaml:"coverage,omitempty"`
	Violations *violationsRuleDTO `yaml:"architecture_violations,omitempty"`
}

type findingsRuleDTO struct {
	MaxError int `yaml:"max_error"`
}

type coverageRuleDTO struct {
	Min int `yaml:"min"`
}

type violationsRuleDTO struct {
	Max int `yaml:"max"`
}

type serverDTO struct {
	URL   string `yaml:"url,omitempty"`
	Token string `yaml:"token,omitempty"`
}
