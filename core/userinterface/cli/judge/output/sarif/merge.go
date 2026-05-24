package sarif

import (
	"encoding/json"
	"fmt"
)

const (
	sarifSchema  = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json"
	sarifVersion = "2.1.0"
)

type mergedDocument struct {
	Schema  string            `json:"$schema"`
	Version string            `json:"version"`
	Runs    []json.RawMessage `json:"runs"`
}

type runsEnvelope struct {
	Runs []json.RawMessage `json:"runs"`
}

func merge(docs [][]byte) ([]byte, error) {
	merged := mergedDocument{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs:    make([]json.RawMessage, 0),
	}

	for _, doc := range docs {
		var env runsEnvelope
		if err := json.Unmarshal(doc, &env); err != nil {
			return nil, fmt.Errorf("parse SARIF document: %w", err)
		}
		merged.Runs = append(merged.Runs, env.Runs...)
	}

	return json.MarshalIndent(merged, "", "  ")
}
