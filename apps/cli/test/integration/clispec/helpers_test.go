package clispec_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type cliSpec struct {
	Commands map[string]commandSpec `yaml:"commands"`
	Globals  globalSpec             `yaml:"globals"`
}

type globalSpec struct {
	Flags map[string]flagSpec `yaml:"flags"`
}

type commandSpec struct {
	Summary string                `yaml:"summary"`
	Flags   map[string]flagSpec   `yaml:"flags"`
	Output  *outputSpec           `yaml:"output"`
}

type flagSpec struct {
	Type      string   `yaml:"type"`
	Short     string   `yaml:"short"`
	Default   any      `yaml:"default"`
	Env       string   `yaml:"env"`
	Sensitive bool     `yaml:"sensitive"`
	Enum      []string `yaml:"enum"`
}

type outputSpec struct {
	Default string                  `yaml:"default"`
	Formats map[string]formatSpec   `yaml:"formats"`
}

type formatSpec struct {
	Flag        string          `yaml:"flag"`
	Description string          `yaml:"description"`
	Schema      any             `yaml:"schema"`
}

func loadSpec(t *testing.T) cliSpec {
	t.Helper()
	path := specPath(t)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "read clispec YAML")

	var spec cliSpec
	require.NoError(t, yaml.Unmarshal(data, &spec), "parse clispec YAML")
	require.NotEmpty(t, spec.Commands, "spec must define at least one command")
	return spec
}

func specPath(t *testing.T) string {
	t.Helper()
	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		workspace := os.Getenv("TEST_WORKSPACE")
		if workspace == "" {
			workspace = "_main"
		}
		p := filepath.Join(srcDir, workspace, "clispec/v1/clispec.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	root := findRepoRoot(t)
	return filepath.Join(root, "clispec/v1/clispec.yaml")
}

func gavelBinary(t *testing.T) string {
	t.Helper()
	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		workspace := os.Getenv("TEST_WORKSPACE")
		if workspace == "" {
			workspace = "_main"
		}
		p := filepath.Join(srcDir, workspace, "apps/cli/cmd/gavel/gavel_/gavel")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p := os.Getenv("GAVEL_BINARY"); p != "" {
		return p
	}
	if p, err := exec.LookPath("gavel"); err == nil {
		return p
	}
	t.Skip("gavel binary not found; run with bazel test or set GAVEL_BINARY")
	return ""
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "MODULE.bazel")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no MODULE.bazel found)")
		}
		dir = parent
	}
}

func runGavel(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	binary := gavelBinary(t)
	cmd := exec.Command(binary, args...)
	cmd.Env = []string{"HOME=" + t.TempDir()}

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run gavel: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

var (
	commandLineRe = regexp.MustCompile(`^\s{2}(\w[\w-]*)\s{2,}`)
	shortFlagRe   = regexp.MustCompile(`^\s+-(\w),\s+--`)
)

var cobraBuiltinCommands = map[string]bool{
	"completion": true,
	"help":       true,
}

var cobraBuiltinFlags = map[string]bool{
	"help": true,
}

func parseHelpCommands(helpOutput string) []string {
	var commands []string
	inSection := false
	for _, line := range strings.Split(helpOutput, "\n") {
		if strings.HasPrefix(line, "Available Commands:") {
			inSection = true
			continue
		}
		if inSection {
			if line == "" || !strings.HasPrefix(line, "  ") {
				break
			}
			if m := commandLineRe.FindStringSubmatch(line); m != nil {
				name := m[1]
				if !cobraBuiltinCommands[name] {
					commands = append(commands, name)
				}
			}
		}
	}
	return commands
}

type parsedFlag struct {
	Name      string
	CobraType string
	Short     string
}

const (
	cobraTypeBool   = "bool"
	cobraTypeString = "string"
)

var knownCobraTypes = map[string]bool{
	"string": true, "int": true, "duration": true,
	"strings": true, "float64": true,
}

func parseHelpFlags(helpOutput string) []parsedFlag {
	var flags []parsedFlag
	inSection := false
	for _, line := range strings.Split(helpOutput, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Flags:" || trimmed == "Global Flags:" {
			inSection = true
			continue
		}
		if inSection {
			if line == "" || (!strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t")) {
				inSection = false
				continue
			}
			flag, ok := parseFlagLine(line)
			if !ok {
				continue
			}
			if cobraBuiltinFlags[flag.Name] {
				continue
			}
			flags = append(flags, flag)
		}
	}
	return flags
}

func parseFlagLine(line string) (parsedFlag, bool) {
	dashIdx := strings.Index(line, "--")
	if dashIdx < 0 {
		return parsedFlag{}, false
	}

	var short string
	if sm := shortFlagRe.FindStringSubmatch(line); sm != nil {
		short = sm[1]
	}

	afterDash := line[dashIdx+2:]
	flagCol := extractFlagColumn(afterDash)

	parts := strings.Fields(flagCol)
	if len(parts) == 0 {
		return parsedFlag{}, false
	}

	name := parts[0]
	cobraType := cobraTypeBool
	if len(parts) == 2 && knownCobraTypes[parts[1]] {
		cobraType = parts[1]
	} else if len(parts) >= 2 {
		cobraType = cobraTypeString
	}

	return parsedFlag{Name: name, CobraType: cobraType, Short: short}, true
}

func extractFlagColumn(source string) string {
	for i := 0; i < len(source)-1; i++ {
		if source[i] == ' ' && source[i+1] == ' ' {
			return strings.TrimSpace(source[:i])
		}
	}
	return strings.TrimSpace(source)
}

func specTypeToCobraType(specType string) string {
	switch specType {
	case cobraTypeBool:
		return cobraTypeBool
	case cobraTypeString:
		return cobraTypeString
	case "int":
		return "int"
	case "duration":
		return "duration"
	case "string[]":
		return "strings"
	default:
		return specType
	}
}
