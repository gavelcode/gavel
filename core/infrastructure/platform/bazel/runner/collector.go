package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type SARIFReport struct {
	Data   []byte
	Source string
}

func bazelBinDir(ctx context.Context, cmd CommandRunner, workspace string) (string, error) {
	stdout, stderr, err := cmd.Run(ctx, workspace, "bazel", "info", "bazel-bin")
	if err != nil {
		return "", fmt.Errorf("bazel info bazel-bin: %w\n%s", err, string(stderr))
	}
	return strings.TrimSpace(string(stdout)), nil
}

func ExtractToolName(data []byte, filename string) string {
	var doc struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &doc); err == nil && len(doc.Runs) > 0 {
		name := doc.Runs[0].Tool.Driver.Name
		if name != "" {
			return name
		}
	}
	return filename
}
