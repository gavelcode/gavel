package project_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/project"
)

func TestProjectDetailFromView_NilLanguagesSerializesAsEmptyArray(t *testing.T) {
	view := &projectview.ProjectDetail{
		ID:            "00000000-0000-0000-0000-000000000001",
		Key:           "core",
		Name:          "Core",
		DefaultBranch: "main",
		Languages:     nil,
	}

	detail := project.ProjectDetailFromView(view)

	data, err := json.Marshal(detail)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	langs, ok := raw["languages"]
	require.True(t, ok, "languages field must be present")
	assert.NotNil(t, langs, "languages must not serialize as null")

	arr, ok := langs.([]any)
	require.True(t, ok, "languages must be an array")
	assert.Empty(t, arr, "languages must be empty when input is nil")
}
