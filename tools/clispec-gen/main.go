// clispec-gen reads a CLI specification YAML and generates Cobra flag
// registration boilerplate (Options struct + RegisterFlags function) for
// each command that declares flags. Generated files are written next to
// the handwritten command code as flags.gen.go.
//
// Usage: go run ./tools/clispec-gen <clispec.yaml>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const (
	expectedArgCount = 2
	exitUsageError   = 2
	filePermission   = 0o644
)

const (
	specTypeBool        = "bool"
	specTypeString      = "string"
	specTypeInt         = "int"
	specTypeDuration    = "duration"
	specTypeStringSlice = "string[]"
)

type spec struct {
	Commands map[string]commandSpec `yaml:"commands"`
}

type commandSpec struct {
	Package string              `yaml:"package"`
	Flags   map[string]flagSpec `yaml:"flags"`
}

type flagSpec struct {
	Type        string   `yaml:"type"`
	Field       string   `yaml:"field"`
	Short       string   `yaml:"short"`
	Default     any      `yaml:"default"`
	Env         string   `yaml:"env"`
	Description string   `yaml:"description"`
	Enum        []string `yaml:"enum"`
}

func main() { os.Exit(execute()) }

func execute() int {
	if len(os.Args) != expectedArgCount {
		fmt.Fprintln(os.Stderr, "usage: clispec-gen <clispec.yaml>")
		return exitUsageError
	}

	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func run(specFile string) error {
	specPath, err := filepath.Abs(specFile)
	if err != nil {
		return fmt.Errorf("resolve spec path: %w", err)
	}
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(specPath)))

	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}

	var s spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	for name, cmd := range s.Commands {
		if len(cmd.Flags) == 0 || cmd.Package == "" {
			continue
		}
		if err := generateCommand(repoRoot, name, cmd); err != nil {
			return err
		}
	}
	return nil
}

type flagData struct {
	FlagName    string
	FieldName   string
	GoType      string
	CobraMethod string
	Default     string
	Short       string
	Env         string
	Description string
}

func generateCommand(repoRoot, cmdName string, cmd commandSpec) error {
	pkgName := filepath.Base(cmd.Package)
	flags := buildFlagData(cmd.Flags)

	needsTime := false
	needsEnv := false
	for _, f := range flags {
		if f.GoType == "time.Duration" {
			needsTime = true
		}
		if f.Env != "" {
			needsEnv = true
		}
	}

	tplData := struct {
		Package   string
		Command   string
		Flags     []flagData
		NeedsTime bool
		NeedsEnv  bool
	}{
		Package:   pkgName,
		Command:   cmdName,
		Flags:     flags,
		NeedsTime: needsTime,
		NeedsEnv:  needsEnv,
	}

	var buf strings.Builder
	if err := genTemplate.Execute(&buf, tplData); err != nil {
		return fmt.Errorf("execute template for %s: %w", cmdName, err)
	}

	outPath := filepath.Join(repoRoot, cmd.Package, "flags.gen.go")
	return os.WriteFile(outPath, []byte(buf.String()), filePermission)
}

func buildFlagData(flags map[string]flagSpec) []flagData {
	var result []flagData
	for name, flag := range flags {
		flagEntry := flagData{
			FlagName:    name,
			FieldName:   fieldName(name, flag.Field),
			GoType:      goType(flag.Type),
			CobraMethod: cobraMethod(flag.Type, flag.Short),
			Default:     goDefault(flag.Type, flag.Default),
			Short:       flag.Short,
			Env:         flag.Env,
			Description: singleLine(flag.Description),
		}
		result = append(result, flagEntry)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FlagName < result[j].FlagName
	})
	return result
}

func fieldName(flagName, override string) string {
	if override != "" {
		return override
	}
	return kebabToPascal(flagName)
}

func kebabToPascal(kebab string) string {
	parts := strings.Split(kebab, "-")
	for idx, part := range parts {
		if part == "" {
			continue
		}
		upper := strings.ToUpper(part)
		if isAcronym(upper) {
			parts[idx] = upper
		} else {
			parts[idx] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func isAcronym(s string) bool {
	switch s {
	case "PR", "JSON", "SARIF", "URL", "ID", "SHA":
		return true
	}
	return false
}

func goType(specType string) string {
	switch specType {
	case specTypeBool:
		return "bool"
	case specTypeString:
		return "string"
	case specTypeInt:
		return "int"
	case specTypeDuration:
		return "time.Duration"
	case specTypeStringSlice:
		return "[]string"
	default:
		return "string"
	}
}

func cobraMethod(specType, short string) string {
	suffix := "Var"
	if short != "" {
		suffix = "VarP"
	}
	switch specType {
	case specTypeBool:
		return "Bool" + suffix
	case specTypeString:
		return "String" + suffix
	case specTypeInt:
		return "Int" + suffix
	case specTypeDuration:
		return "Duration" + suffix
	case specTypeStringSlice:
		return "StringSlice" + suffix
	default:
		return "String" + suffix
	}
}

func goDefault(specType string, val any) string {
	if val == nil {
		switch specType {
		case specTypeBool:
			return "false"
		case specTypeString:
			return `""`
		case specTypeInt:
			return "0"
		case specTypeDuration:
			return "0"
		case specTypeStringSlice:
			return "nil"
		default:
			return `""`
		}
	}
	switch specType {
	case specTypeBool:
		return fmt.Sprintf("%v", val)
	case specTypeString:
		return fmt.Sprintf("%q", val)
	case specTypeInt:
		return fmt.Sprintf("%v", val)
	case specTypeDuration:
		return parseDuration(fmt.Sprintf("%v", val))
	case specTypeStringSlice:
		return formatStringSlice(val)
	default:
		return fmt.Sprintf("%q", val)
	}
}

func parseDuration(source string) string {
	suffixes := []struct {
		suffix string
		expr   string
	}{
		{"ms", " * time.Millisecond"},
		{"s", " * time.Second"},
		{"m", " * time.Minute"},
		{"h", " * time.Hour"},
	}
	for _, entry := range suffixes {
		if num, ok := strings.CutSuffix(source, entry.suffix); ok {
			return num + entry.expr
		}
	}
	return "0"
}

func formatStringSlice(val any) string {
	switch v := val.(type) {
	case []any:
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = fmt.Sprintf("%q", item)
		}
		return "[]string{" + strings.Join(items, ", ") + "}"
	default:
		return "nil"
	}
}

func singleLine(source string) string {
	result := strings.TrimSpace(source)
	result = strings.ReplaceAll(result, "\n", " ")
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	result = strings.ReplaceAll(result, `"`, `\"`)
	return result
}


var genTemplate = template.Must(template.New("flags").Parse(`// Code generated by clispec-gen from clispec/v1/clispec.yaml. DO NOT EDIT.
package {{ .Package }}

import (
{{- if or .NeedsEnv .NeedsTime }}
{{- if .NeedsEnv }}
	"os"
{{- end }}
{{- if .NeedsTime }}
	"time"
{{- end }}

	"github.com/spf13/cobra"
{{- else }}
	"github.com/spf13/cobra"
{{- end }}
)

// Options holds the parsed flag values for the {{ .Command }} command.
type Options struct {
{{- range .Flags }}
	{{ .FieldName }} {{ .GoType }}
{{- end }}
}

// RegisterFlags binds all {{ .Command }} flags to the given Options struct.
func RegisterFlags(cmd *cobra.Command, opts *Options) {
{{- range .Flags }}
{{- if .Env }}
	cmd.Flags().{{ .CobraMethod }}(&opts.{{ .FieldName }}, "{{ .FlagName }}"{{ if .Short }}, "{{ .Short }}"{{ end }}, envOrDefault("{{ .Env }}", {{ .Default }}), "{{ .Description }}")
{{- else if .Short }}
	cmd.Flags().{{ .CobraMethod }}(&opts.{{ .FieldName }}, "{{ .FlagName }}", "{{ .Short }}", {{ .Default }}, "{{ .Description }}")
{{- else }}
	cmd.Flags().{{ .CobraMethod }}(&opts.{{ .FieldName }}, "{{ .FlagName }}", {{ .Default }}, "{{ .Description }}")
{{- end }}
{{- end }}
}
{{- if .NeedsEnv }}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
{{- end }}
`))
