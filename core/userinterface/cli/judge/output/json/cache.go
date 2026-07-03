package json

import (
	encjson "encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

const (
	dirPermission  = 0o755
	filePermission = 0o644
)

func WriteCache(workspace string, results []pipeline.Result) error {
	for _, result := range results {
		if err := writeProjectCache(workspace, result); err != nil {
			return fmt.Errorf("cache %s: %w", result.Name, err)
		}
	}
	return nil
}

func writeProjectCache(workspace string, result pipeline.Result) error {
	dir := filepath.Join(workspace, gavelDir, resultsDir, result.Name)
	if err := os.MkdirAll(dir, dirPermission); err != nil {
		return fmt.Errorf("create results dir: %w", err)
	}

	dto := toProjectDTO(result)
	data, err := encjson.MarshalIndent(dto, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal verdict: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, verdictFile), data, filePermission)
}
