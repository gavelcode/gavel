package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFlags(t *testing.T) {
	t.Helper()
	orig := flag.CommandLine
	t.Cleanup(func() { flag.CommandLine = orig })
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
}

func setArgs(t *testing.T, args []string) {
	t.Helper()
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = args
}

func TestExecute_MissingArgs(t *testing.T) {
	resetFlags(t)
	setArgs(t, []string{"clispec-gen"})

	code := execute()

	assert.Equal(t, 2, code)
}

func TestExecute_TooManyArgs(t *testing.T) {
	resetFlags(t)
	setArgs(t, []string{"clispec-gen", "a", "b"})

	code := execute()

	assert.Equal(t, 2, code)
}

func TestExecute_NonexistentFile(t *testing.T) {
	resetFlags(t)
	setArgs(t, []string{"clispec-gen", "/nonexistent/spec.yaml"})

	code := execute()

	assert.Equal(t, 1, code)
}

func TestExecute_InvalidYAML(t *testing.T) {
	resetFlags(t)
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	require.NoError(t, os.WriteFile(specPath, []byte("invalid: [yaml"), 0o644))
	setArgs(t, []string{"clispec-gen", specPath})

	code := execute()

	assert.Equal(t, 1, code)
}

func TestExecute_Success(t *testing.T) {
	resetFlags(t)
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	pkgDir := filepath.Join(dir, "core", "userinterface", "cli", "judge")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	specContent := `commands:
  judge:
    package: core/userinterface/cli/judge
    flags:
      verbose:
        type: bool
        description: Enable verbose output
`
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))
	setArgs(t, []string{"clispec-gen", specPath})

	code := execute()

	assert.Equal(t, 0, code)
	genFile := filepath.Join(pkgDir, "flags.gen.go")
	data, err := os.ReadFile(genFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "package judge")
	assert.Contains(t, string(data), "Verbose bool")
}

func TestRun_ReadError(t *testing.T) {
	err := run("/nonexistent/spec.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read spec")
}

func TestRun_ParseError(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	require.NoError(t, os.WriteFile(specPath, []byte("invalid: [yaml"), 0o644))

	err := run(specPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse spec")
}

func TestRun_SkipsCommandsWithoutFlags(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	specContent := `commands:
  validate:
    package: core/userinterface/cli/validate
`
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))

	err := run(specPath)

	require.NoError(t, err)
}

func TestRun_SkipsCommandsWithoutPackage(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	specContent := `commands:
  judge:
    flags:
      verbose:
        type: bool
`
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))

	err := run(specPath)

	require.NoError(t, err)
}

func TestRun_GenerateCommandWriteError(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "clispec", "v1", "clispec.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0o755))
	specContent := `commands:
  judge:
    package: nonexistent/deep/path
    flags:
      verbose:
        type: bool
        description: verbose
`
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))

	err := run(specPath)

	require.Error(t, err)
}

func TestGenerateCommand_ProducesValidGo(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "judge")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/judge",
		Flags: map[string]flagSpec{
			"verbose": {Type: "bool", Description: "verbose output"},
			"output":  {Type: "string", Short: "o", Default: "text", Description: "output format"},
		},
	}

	err := generateCommand(dir, "judge", cmd)

	require.NoError(t, err)
	data, readErr := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	require.NoError(t, readErr)
	content := string(data)
	assert.Contains(t, content, "package judge")
	assert.Contains(t, content, "Verbose bool")
	assert.Contains(t, content, "Output string")
	assert.Contains(t, content, `"verbose"`)
	assert.Contains(t, content, `"output", "o"`)
}

func TestGenerateCommand_WithDuration(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "watch")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/watch",
		Flags: map[string]flagSpec{
			"interval": {Type: "duration", Default: "500ms", Description: "poll interval"},
		},
	}

	err := generateCommand(dir, "watch", cmd)

	require.NoError(t, err)
	data, readErr := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	require.NoError(t, readErr)
	content := string(data)
	assert.Contains(t, content, `"time"`)
	assert.Contains(t, content, "time.Duration")
	assert.Contains(t, content, "500 * time.Millisecond")
}

func TestGenerateCommand_WithEnvVar(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "judge")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/judge",
		Flags: map[string]flagSpec{
			"server": {Type: "string", Env: "GAVEL_SERVER_URL", Description: "server URL"},
		},
	}

	err := generateCommand(dir, "judge", cmd)

	require.NoError(t, err)
	data, readErr := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	require.NoError(t, readErr)
	content := string(data)
	assert.Contains(t, content, `"os"`)
	assert.Contains(t, content, "envOrDefault")
	assert.Contains(t, content, "GAVEL_SERVER_URL")
}

func TestGenerateCommand_WithStringSlice(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "judge")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/judge",
		Flags: map[string]flagSpec{
			"projects": {Type: "string[]", Default: []any{"core", "cli"}, Description: "projects"},
		},
	}

	err := generateCommand(dir, "judge", cmd)

	require.NoError(t, err)
	data, readErr := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	require.NoError(t, readErr)
	content := string(data)
	assert.Contains(t, content, "[]string")
	assert.Contains(t, content, `[]string{"core", "cli"}`)
}

func TestGenerateCommand_WriteError(t *testing.T) {
	cmd := commandSpec{
		Package: "nonexistent/path",
		Flags:   map[string]flagSpec{"v": {Type: "bool"}},
	}

	err := generateCommand("/dev/null", "cmd", cmd)

	require.Error(t, err)
}

func TestBuildFlagData_SortsByFlagName(t *testing.T) {
	flags := map[string]flagSpec{
		"zoo":    {Type: "string"},
		"alpha":  {Type: "bool"},
		"middle": {Type: "int"},
	}

	result := buildFlagData(flags)

	require.Len(t, result, 3)
	assert.Equal(t, "alpha", result[0].FlagName)
	assert.Equal(t, "middle", result[1].FlagName)
	assert.Equal(t, "zoo", result[2].FlagName)
}

func TestBuildFlagData_MapsFieldsCorrectly(t *testing.T) {
	flags := map[string]flagSpec{
		"skip-tests": {
			Type:        "bool",
			Field:       "SkipTests",
			Short:       "s",
			Default:     true,
			Env:         "SKIP",
			Description: "Skip test files",
		},
	}

	result := buildFlagData(flags)

	require.Len(t, result, 1)
	flagEntry := result[0]
	assert.Equal(t, "skip-tests", flagEntry.FlagName)
	assert.Equal(t, "SkipTests", flagEntry.FieldName)
	assert.Equal(t, "bool", flagEntry.GoType)
	assert.Equal(t, "BoolVarP", flagEntry.CobraMethod)
	assert.Equal(t, "true", flagEntry.Default)
	assert.Equal(t, "s", flagEntry.Short)
	assert.Equal(t, "SKIP", flagEntry.Env)
	assert.Equal(t, "Skip test files", flagEntry.Description)
}

func TestFieldName_WithOverride(t *testing.T) {
	assert.Equal(t, "CustomName", fieldName("flag-name", "CustomName"))
}

func TestFieldName_WithoutOverride(t *testing.T) {
	assert.Equal(t, "FlagName", fieldName("flag-name", ""))
}

func TestKebabToPascal(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"verbose", "Verbose"},
		{"skip-tests", "SkipTests"},
		{"output-format", "OutputFormat"},
		{"pr-number", "PRNumber"},
		{"json-output", "JSONOutput"},
		{"sarif-file", "SARIFFile"},
		{"server-url", "ServerURL"},
		{"case-id", "CaseID"},
		{"commit-sha", "CommitSHA"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, kebabToPascal(tt.input))
		})
	}
}

func TestIsAcronym(t *testing.T) {
	assert.True(t, isAcronym("PR"))
	assert.True(t, isAcronym("JSON"))
	assert.True(t, isAcronym("SARIF"))
	assert.True(t, isAcronym("URL"))
	assert.True(t, isAcronym("ID"))
	assert.True(t, isAcronym("SHA"))
	assert.False(t, isAcronym("FOO"))
	assert.False(t, isAcronym("name"))
}

func TestGoType(t *testing.T) {
	assert.Equal(t, "bool", goType("bool"))
	assert.Equal(t, "string", goType("string"))
	assert.Equal(t, "int", goType("int"))
	assert.Equal(t, "time.Duration", goType("duration"))
	assert.Equal(t, "[]string", goType("string[]"))
	assert.Equal(t, "string", goType("unknown"))
}

func TestCobraMethod(t *testing.T) {
	assert.Equal(t, "BoolVar", cobraMethod("bool", ""))
	assert.Equal(t, "BoolVarP", cobraMethod("bool", "b"))
	assert.Equal(t, "StringVar", cobraMethod("string", ""))
	assert.Equal(t, "StringVarP", cobraMethod("string", "s"))
	assert.Equal(t, "IntVar", cobraMethod("int", ""))
	assert.Equal(t, "IntVarP", cobraMethod("int", "n"))
	assert.Equal(t, "DurationVar", cobraMethod("duration", ""))
	assert.Equal(t, "DurationVarP", cobraMethod("duration", "d"))
	assert.Equal(t, "StringSliceVar", cobraMethod("string[]", ""))
	assert.Equal(t, "StringSliceVarP", cobraMethod("string[]", "p"))
	assert.Equal(t, "StringVar", cobraMethod("unknown", ""))
	assert.Equal(t, "StringVarP", cobraMethod("unknown", "x"))
}

func TestGoDefault_NilValues(t *testing.T) {
	assert.Equal(t, "false", goDefault("bool", nil))
	assert.Equal(t, `""`, goDefault("string", nil))
	assert.Equal(t, "0", goDefault("int", nil))
	assert.Equal(t, "0", goDefault("duration", nil))
	assert.Equal(t, "nil", goDefault("string[]", nil))
	assert.Equal(t, `""`, goDefault("unknown", nil))
}

func TestGoDefault_WithValues(t *testing.T) {
	assert.Equal(t, "true", goDefault("bool", true))
	assert.Equal(t, "false", goDefault("bool", false))
	assert.Equal(t, `"text"`, goDefault("string", "text"))
	assert.Equal(t, "42", goDefault("int", 42))
	assert.Equal(t, "500 * time.Millisecond", goDefault("duration", "500ms"))
	assert.Equal(t, `[]string{"a", "b"}`, goDefault("string[]", []any{"a", "b"}))
	assert.Equal(t, `"other"`, goDefault("unknown", "other"))
}

func TestParseDuration(t *testing.T) {
	assert.Equal(t, "500 * time.Millisecond", parseDuration("500ms"))
	assert.Equal(t, "5 * time.Second", parseDuration("5s"))
	assert.Equal(t, "10 * time.Minute", parseDuration("10m"))
	assert.Equal(t, "2 * time.Hour", parseDuration("2h"))
	assert.Equal(t, "0", parseDuration("unknown"))
	assert.Equal(t, "0", parseDuration(""))
}

func TestFormatStringSlice(t *testing.T) {
	assert.Equal(t, `[]string{"a", "b", "c"}`, formatStringSlice([]any{"a", "b", "c"}))
	assert.Equal(t, `[]string{}`, formatStringSlice([]any{}))
	assert.Equal(t, "nil", formatStringSlice("not a slice"))
	assert.Equal(t, "nil", formatStringSlice(nil))
}

func TestSingleLine(t *testing.T) {
	assert.Equal(t, "hello world", singleLine("hello\nworld"))
	assert.Equal(t, "one two", singleLine("  one  two  "))
	assert.Equal(t, `say \"hi\"`, singleLine(`say "hi"`))
	assert.Equal(t, "a b c", singleLine("a  b  c"))
	assert.Equal(t, "", singleLine(""))
}

func TestSingleLine_MultipleSpaces(t *testing.T) {
	assert.Equal(t, "a b", singleLine("a    b"))
}

func TestGenerateCommand_FieldOverride(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "judge")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/judge",
		Flags: map[string]flagSpec{
			"skip-tests": {Type: "bool", Field: "SkipTests", Description: "skip"},
		},
	}

	err := generateCommand(dir, "judge", cmd)

	require.NoError(t, err)
	data, _ := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	assert.Contains(t, string(data), "SkipTests bool")
}

func TestGenerateCommand_NoTimeOrEnv(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cli", "simple")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	cmd := commandSpec{
		Package: "cli/simple",
		Flags: map[string]flagSpec{
			"name": {Type: "string", Description: "a name"},
		},
	}

	err := generateCommand(dir, "simple", cmd)

	require.NoError(t, err)
	data, _ := os.ReadFile(filepath.Join(pkgDir, "flags.gen.go"))
	content := string(data)
	assert.NotContains(t, content, `"time"`)
	assert.NotContains(t, content, `"os"`)
	assert.NotContains(t, content, "envOrDefault")
}

func TestKebabToPascal_EmptyPart(t *testing.T) {
	got := kebabToPascal("-leading")

	assert.True(t, strings.HasSuffix(got, "Leading"))
}
