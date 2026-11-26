package sarif

import (
	"encoding/json"
	"fmt"
	"strings"
)

// gavelPropertyKey is the SARIF property-bag key under which gavel records
// provenance on each run it emits.
const gavelPropertyKey = "gavel"

// ValidateBridgeSubmission checks a bridge's SARIF against the bridge contract
// and returns one human-readable problem per violation (none when it
// conforms). The contract is owned by gavel-tools (docs/bridge.md) and is
// deliberately stricter than Parser, which tolerates missing locations for
// catalog tools; this strictness applies to bridges only.
func ValidateBridgeSubmission(data []byte) []string {
	var doc bridgeDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return []string{fmt.Sprintf("not valid SARIF JSON: %v", err)}
	}
	if len(doc.Runs) == 0 {
		return []string{"runs: at least one run is required"}
	}
	var problems []string
	for runIndex, run := range doc.Runs {
		report := func(format string, args ...any) {
			problems = append(problems, fmt.Sprintf("runs[%d]: %s", runIndex, fmt.Sprintf(format, args...)))
		}
		if strings.TrimSpace(run.Tool.Driver.Name) == "" {
			report("tool.driver.name must be non-empty")
		}
		for index, res := range run.Results {
			problems = append(problems, validateBridgeResult(runIndex, index, res)...)
		}
		for index, inv := range run.Invocations {
			if !inv.ExecutionSuccessful && !hasExplanation(inv.ToolExecutionNotifications) {
				report("invocations[%d]: executionSuccessful=false requires a toolExecutionNotifications entry with message.text", index)
			}
		}
	}
	return problems
}

func validateBridgeResult(runIndex, index int, res bridgeResultDoc) []string {
	var problems []string
	report := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf("runs[%d]: results[%d]: %s", runIndex, index, fmt.Sprintf(format, args...)))
	}
	if strings.TrimSpace(res.RuleID) == "" {
		report("ruleId must be non-empty")
	}
	if strings.TrimSpace(res.Message.Text) == "" {
		report("message.text must be non-empty")
	}
	switch res.Level {
	case "", "error", "warning", "note", "none":
	default:
		report("level %q must be error, warning, note or none", res.Level)
	}
	if len(res.Locations) == 0 {
		report("locations[0] is required")
		return problems
	}
	physical := res.Locations[0].PhysicalLocation
	switch uri := physical.ArtifactLocation.URI; {
	case uri == "":
		report("locations[0].physicalLocation.artifactLocation.uri must be non-empty")
	case !workspaceRelative(uri):
		report("uri %q must be workspace-relative (no scheme, leading '/', '..' or execroot)", uri)
	}
	if physical.Region.StartLine < 1 {
		report("locations[0].physicalLocation.region.startLine must be >= 1")
	}
	return problems
}

func hasExplanation(notes []bridgeNotification) bool {
	for _, note := range notes {
		if strings.TrimSpace(note.Message.Text) != "" {
			return true
		}
	}
	return false
}

func workspaceRelative(uri string) bool {
	if strings.Contains(uri, "://") || strings.HasPrefix(uri, "/") || strings.Contains(uri, "execroot/") {
		return false
	}
	for _, part := range strings.Split(uri, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

// TagRuns records properties under runs[].properties.gavel without otherwise
// altering the document, so emitted SARIF carries gavel's provenance marker.
func TagRuns(data []byte, properties map[string]any) ([]byte, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeSARIF, err)
	}
	var runs []map[string]json.RawMessage
	if err := json.Unmarshal(doc["runs"], &runs); err != nil {
		return nil, fmt.Errorf("%w: runs: %w", ErrDecodeSARIF, err)
	}
	gavelBag, err := json.Marshal(properties)
	if err != nil {
		return nil, fmt.Errorf("encode gavel properties: %w", err)
	}
	for _, run := range runs {
		bag := map[string]json.RawMessage{}
		if existing, ok := run["properties"]; ok {
			if err := json.Unmarshal(existing, &bag); err != nil {
				return nil, fmt.Errorf("%w: run properties: %w", ErrDecodeSARIF, err)
			}
		}
		bag[gavelPropertyKey] = gavelBag
		encoded, err := json.Marshal(bag)
		if err != nil {
			return nil, fmt.Errorf("encode run properties: %w", err)
		}
		run["properties"] = encoded
	}
	encodedRuns, err := json.Marshal(runs)
	if err != nil {
		return nil, fmt.Errorf("encode runs: %w", err)
	}
	doc["runs"] = encodedRuns
	return json.Marshal(doc)
}

type bridgeDocument struct {
	Runs []struct {
		Tool struct {
			Driver struct {
				Name string `json:"name"`
			} `json:"driver"`
		} `json:"tool"`
		Results     []bridgeResultDoc `json:"results"`
		Invocations []struct {
			ExecutionSuccessful        bool                 `json:"executionSuccessful"`
			ToolExecutionNotifications []bridgeNotification `json:"toolExecutionNotifications"`
		} `json:"invocations"`
	} `json:"runs"`
}

type bridgeNotification struct {
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

type bridgeResultDoc struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []struct {
		PhysicalLocation struct {
			ArtifactLocation struct {
				URI string `json:"uri"`
			} `json:"artifactLocation"`
			Region struct {
				StartLine int `json:"startLine"`
			} `json:"region"`
		} `json:"physicalLocation"`
	} `json:"locations"`
}
