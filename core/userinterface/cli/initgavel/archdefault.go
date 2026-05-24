package initgavel

import (
	"os"
	"path/filepath"
)

const (
	architectureConfigFile = ".gavel/architecture.yml"
	dirPermission          = 0o755
	filePermission         = 0o644
)

var layersByLanguage = map[string]map[string][]string{
	"go": {
		"domain":         {"internal/domain/..."},
		"application":    {"internal/application/..."},
		"infrastructure": {"internal/infrastructure/..."},
		"userinterface":  {"internal/userinterface/..."},
	},
	"java": {
		"domain":         {"src/main/java/**/domain/..."},
		"application":    {"src/main/java/**/application/..."},
		"infrastructure": {"src/main/java/**/infrastructure/..."},
		"userinterface":  {"src/main/java/**/userinterface/..."},
	},
	"python": {
		"domain":         {"src/domain/..."},
		"application":    {"src/application/..."},
		"infrastructure": {"src/infrastructure/..."},
		"userinterface":  {"src/userinterface/..."},
	},
	"typescript": {
		"domain":         {"src/domain/..."},
		"application":    {"src/application/..."},
		"infrastructure": {"src/infrastructure/..."},
		"userinterface":  {"src/userinterface/..."},
	},
	"rust": {
		"domain":         {"src/domain/..."},
		"application":    {"src/application/..."},
		"infrastructure": {"src/infrastructure/..."},
		"userinterface":  {"src/userinterface/..."},
	},
}

func writeArchitectureConfig(workspace string, tooling []string) error {
	path := filepath.Join(workspace, architectureConfigFile)

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	layers := pickLayers(tooling)
	if len(layers) == 0 {
		return nil
	}

	content := renderArchitectureYAML(layers)
	return os.WriteFile(path, []byte(content), filePermission)
}

func pickLayers(tooling []string) map[string][]string {
	for _, lang := range tooling {
		if l, ok := layersByLanguage[lang]; ok {
			return l
		}
	}
	return nil
}

func renderArchitectureYAML(layers map[string][]string) string {
	var builder []byte
	builder = append(builder, "layers:\n"...)
	for _, name := range []string{"domain", "application", "infrastructure", "userinterface"} {
		patterns, ok := layers[name]
		if !ok {
			continue
		}
		builder = append(builder, "  "+name+": ["...)
		for i, p := range patterns {
			if i > 0 {
				builder = append(builder, ", "...)
			}
			builder = append(builder, "\""+p+"\""...)
		}
		builder = append(builder, "]\n"...)
	}

	builder = append(builder, "\nrules:\n"...)
	builder = append(builder, "  - name: domain-imports-nothing\n"...)
	builder = append(builder, "    source: domain\n"...)
	builder = append(builder, "    deny: [application, infrastructure, userinterface]\n\n"...)
	builder = append(builder, "  - name: application-imports-domain-only\n"...)
	builder = append(builder, "    source: application\n"...)
	builder = append(builder, "    deny: [infrastructure, userinterface]\n\n"...)
	builder = append(builder, "  - name: infrastructure-no-application\n"...)
	builder = append(builder, "    source: infrastructure\n"...)
	builder = append(builder, "    deny: [application, userinterface]\n\n"...)
	builder = append(builder, "  - name: userinterface-application-only\n"...)
	builder = append(builder, "    source: userinterface\n"...)
	builder = append(builder, "    deny: [domain, infrastructure]\n\n"...)
	builder = append(builder, "detect_cycles: true\n"...)

	return string(builder)
}
