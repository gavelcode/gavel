package clispec_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllSpecCommandsExistInBinary(t *testing.T) {
	spec := loadSpec(t)
	for name := range spec.Commands {
		t.Run(name, func(t *testing.T) {
			stdout, _, exitCode := runGavel(t, name, "--help")
			assert.Equal(t, 0, exitCode, "gavel %s --help should exit 0", name)
			assert.NotEmpty(t, stdout, "gavel %s --help should produce output", name)
		})
	}
}

func TestAllBinaryCommandsDocumentedInSpec(t *testing.T) {
	spec := loadSpec(t)
	stdout, _, _ := runGavel(t, "--help")
	commands := parseHelpCommands(stdout)
	require.NotEmpty(t, commands, "binary should list at least one command")

	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			_, exists := spec.Commands[name]
			assert.True(t, exists, "command %q found in binary but not in clispec", name)
		})
	}
}

func TestAllSpecFlagsExistInBinary(t *testing.T) {
	spec := loadSpec(t)
	for cmdName, cmdSpec := range spec.Commands {
		t.Run(cmdName, func(t *testing.T) {
			stdout, _, _ := runGavel(t, cmdName, "--help")
			binaryFlags := parseHelpFlags(stdout)
			binaryFlagNames := make(map[string]bool, len(binaryFlags))
			for _, f := range binaryFlags {
				binaryFlagNames[f.Name] = true
			}

			for flagName := range cmdSpec.Flags {
				t.Run(flagName, func(t *testing.T) {
					assert.True(t, binaryFlagNames[flagName],
						"flag --%s declared in spec for command %q but not found in binary help output",
						flagName, cmdName)
				})
			}
		})
	}
}

func TestAllBinaryFlagsDocumentedInSpec(t *testing.T) {
	spec := loadSpec(t)
	globalFlags := make(map[string]bool, len(spec.Globals.Flags))
	for name := range spec.Globals.Flags {
		globalFlags[name] = true
	}

	for cmdName, cmdSpec := range spec.Commands {
		t.Run(cmdName, func(t *testing.T) {
			stdout, _, _ := runGavel(t, cmdName, "--help")
			binaryFlags := parseHelpFlags(stdout)

			for _, flag := range binaryFlags {
				if globalFlags[flag.Name] {
					continue
				}
				t.Run(flag.Name, func(t *testing.T) {
					_, exists := cmdSpec.Flags[flag.Name]
					assert.True(t, exists,
						"flag --%s found in binary for command %q but not in clispec",
						flag.Name, cmdName)
				})
			}
		})
	}
}

func TestFlagTypesMatchSpec(t *testing.T) {
	spec := loadSpec(t)
	for cmdName, cmdSpec := range spec.Commands {
		t.Run(cmdName, func(t *testing.T) {
			stdout, _, _ := runGavel(t, cmdName, "--help")
			binaryFlags := parseHelpFlags(stdout)
			binaryFlagMap := make(map[string]parsedFlag, len(binaryFlags))
			for _, f := range binaryFlags {
				binaryFlagMap[f.Name] = f
			}

			for flagName, flagDef := range cmdSpec.Flags {
				binaryFlag, exists := binaryFlagMap[flagName]
				if !exists {
					continue
				}
				t.Run(flagName, func(t *testing.T) {
					expectedType := specTypeToCobraType(flagDef.Type)
					assert.Equal(t, expectedType, binaryFlag.CobraType,
						"type mismatch for --%s in %q: spec says %q, binary shows %q",
						flagName, cmdName, flagDef.Type, binaryFlag.CobraType)
				})
			}
		})
	}
}

func TestFlagShortAliasesMatchSpec(t *testing.T) {
	spec := loadSpec(t)
	for cmdName, cmdSpec := range spec.Commands {
		t.Run(cmdName, func(t *testing.T) {
			stdout, _, _ := runGavel(t, cmdName, "--help")
			binaryFlags := parseHelpFlags(stdout)
			binaryFlagMap := make(map[string]parsedFlag, len(binaryFlags))
			for _, f := range binaryFlags {
				binaryFlagMap[f.Name] = f
			}

			for flagName, flagDef := range cmdSpec.Flags {
				if flagDef.Short == "" {
					continue
				}
				binaryFlag, exists := binaryFlagMap[flagName]
				if !exists {
					continue
				}
				t.Run(fmt.Sprintf("%s/-%s", flagName, flagDef.Short), func(t *testing.T) {
					assert.Equal(t, flagDef.Short, binaryFlag.Short,
						"short alias mismatch for --%s in %q: spec says -%s, binary shows -%s",
						flagName, cmdName, flagDef.Short, binaryFlag.Short)
				})
			}
		})
	}
}

func TestGlobalFlagsExistInBinary(t *testing.T) {
	spec := loadSpec(t)
	stdout, _, _ := runGavel(t, "--help")
	binaryFlags := parseHelpFlags(stdout)
	binaryFlagNames := make(map[string]bool, len(binaryFlags))
	for _, f := range binaryFlags {
		binaryFlagNames[f.Name] = true
	}

	for flagName := range spec.Globals.Flags {
		t.Run(flagName, func(t *testing.T) {
			assert.True(t, binaryFlagNames[flagName],
				"global flag --%s declared in spec but not found in binary", flagName)
		})
	}
}

func TestUnknownFlagRejected(t *testing.T) {
	spec := loadSpec(t)
	for name := range spec.Commands {
		t.Run(name, func(t *testing.T) {
			_, _, exitCode := runGavel(t, name, "--nonexistent-flag-xyz")
			assert.NotEqual(t, 0, exitCode,
				"gavel %s should reject unknown flags", name)
		})
	}
}
