package sarif

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

const (
	unknownTool  = "unknown"
	errorLevel   = "error"
	warningLevel = "warning"
)

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
				if reason := degradedReason(inv); reason != "" {
					failures = append(failures, evidencedto.ToolFailure{
						Tool:     toolName,
						Reason:   reason,
						Degraded: true,
					})
				}
				continue
			}
			if isConfigurationOnly(inv) {
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

// degradedReason collects the warning-level execution notifications a tool
// attaches to a *successful* run — its way of saying "I ran but could only
// analyze part of this" (e.g. Error Prone on a target whose annotation
// processors it could not replay). Returns "" when the run was clean.
func degradedReason(inv invocation) string {
	var texts []string
	for _, note := range inv.ToolExecutionNotifications {
		if !strings.EqualFold(strings.TrimSpace(note.Level), warningLevel) {
			continue
		}
		if text := strings.TrimSpace(note.Message.Text); text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "; ")
}

func isConfigurationOnly(inv invocation) bool {
	if len(inv.ToolExecutionNotifications) > 0 || len(inv.ToolConfigurationNotifications) == 0 {
		return false
	}
	return !hasErrorNotification(inv.ToolConfigurationNotifications)
}

func hasErrorNotification(notes []notification) bool {
	for _, note := range notes {
		if strings.EqualFold(strings.TrimSpace(note.Level), errorLevel) {
			return true
		}
	}
	return false
}

func invocationReason(inv invocation) string {
	reasons := make([]string, 0, len(inv.ToolExecutionNotifications)+len(inv.ToolConfigurationNotifications))
	reasons = appendNotificationTexts(reasons, inv.ToolExecutionNotifications)
	reasons = appendNotificationTexts(reasons, inv.ToolConfigurationNotifications)
	if len(reasons) == 0 {
		return "analyzer reported an unsuccessful run"
	}
	return strings.Join(reasons, "; ")
}

func appendNotificationTexts(reasons []string, notes []notification) []string {
	for _, note := range notes {
		if text := strings.TrimSpace(note.Message.Text); text != "" {
			reasons = append(reasons, text)
		}
	}
	return reasons
}
