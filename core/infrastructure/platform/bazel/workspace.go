package bazel

import (
	"fmt"
	"os"
	"path/filepath"
)

func WorkspaceDir() (string, error) {
	var dir string
	if workspace := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); workspace != "" {
		dir = workspace
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}

	if !isBazelWorkspace(dir) {
		return "", fmt.Errorf("directory %s is not a Bazel workspace (no MODULE.bazel or WORKSPACE file found)", dir)
	}
	return dir, nil
}

func isBazelWorkspace(dir string) bool {
	markers := []string{"MODULE.bazel", "WORKSPACE", "WORKSPACE.bazel"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}
