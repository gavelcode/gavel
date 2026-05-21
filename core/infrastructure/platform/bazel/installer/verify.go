package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StructureVerifier interface {
	VerifyStructure(workspacePath string) ([]string, error)
}

func (i *Installer) VerifyStructure(workspacePath string) ([]string, error) {
	var issues []string

	for _, file := range []string{GavelBazelrc, GavelModule} {
		path := filepath.Join(workspacePath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("%s not found", file))
		}
	}

	if err := verifyIncludeLine(workspacePath, ".bazelrc", bazelrcInclude); err != nil {
		issues = append(issues, err.Error())
	}

	if err := verifyIncludeLine(workspacePath, "MODULE.bazel", moduleInclude); err != nil {
		issues = append(issues, err.Error())
	}

	return issues, nil
}

func verifyIncludeLine(root, filename, line string) error {
	path := filepath.Join(root, filename)
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found", filename)
		}
		return fmt.Errorf("read %s: %w", filename, err)
	}

	if !strings.Contains(string(existing), line) {
		return fmt.Errorf("%s missing: %s", filename, line)
	}
	return nil
}
