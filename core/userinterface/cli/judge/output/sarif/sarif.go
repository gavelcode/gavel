package sarif

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirPermission  = 0o755
	filePermission = 0o644
)

func Write(path string, docs [][]byte) error {
	data, err := merge(docs)
	if err != nil {
		return fmt.Errorf("merge SARIF: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPermission); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, data, filePermission)
}
