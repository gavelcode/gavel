package sarif

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

const unknownTool = "unknown"

// ParseToolExecutions extracts the analyzer runs that did not complete from a
// SARIF document: each invocation with executionSuccessful=false becomes a tool
// failure carrying its tool name and the concrete reason from the
// toolExecutionNotifications. A clean document yields no failures.
func (*Parser) ParseToolExecutions(data []byte) ([]evidencedto.ToolFailure, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeSARIF, err)
	}

	var failures []evidencedto.ToolFailure
	for _, currentRun := range doc.Runs {
		toolName := currentRun.Tool.Driver.Name
		if strings.TrimSpace(toolName) == "" {
			toolName = unknownTool
		}
		for _, inv := range currentRun.Invocations {
			if inv.ExecutionSuccessful {
				continue
			}
			failures = append(failures, evidencedto.ToolFailure{
				Tool:   toolName,
				Reason: invocationReason(inv),
			})
		}
	}
	return failures, nil
}

func invocationReason(inv invocation) string {
	reasons := make([]string, 0, len(inv.ToolExecutionNotifications))
	for _, note := range inv.ToolExecutionNotifications {
		if text := strings.TrimSpace(note.Message.Text); text != "" {
			reasons = append(reasons, text)
		}
	}
	if len(reasons) == 0 {
		return "analyzer reported an unsuccessful run"
	}
	return strings.Join(reasons, "; ")
}
