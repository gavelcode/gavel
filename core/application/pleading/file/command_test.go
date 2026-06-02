package file_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/pleading/file"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name         string
		projectID    string
		number       int
		title        string
		petitioner   string
		sourceBranch string
		targetBranch string
		commitSHA    string
		expectErr    bool
	}{
		{
			name:         "valid",
			projectID:    "proj-1",
			number:       42,
			title:        "Add login",
			petitioner:   "alice",
			sourceBranch: "feature/login",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    false,
		},
		{
			name:         "valid without petitioner",
			projectID:    "proj-1",
			number:       42,
			title:        "Add login",
			petitioner:   "",
			sourceBranch: "feature/login",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    false,
		},
		{name: "empty projectID rejected", projectID: "", number: 1, title: "t", sourceBranch: "s", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "zero number rejected", projectID: "p", number: 0, title: "t", sourceBranch: "s", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "negative number rejected", projectID: "p", number: -1, title: "t", sourceBranch: "s", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "empty title rejected", projectID: "p", number: 1, title: "", sourceBranch: "s", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "whitespace title rejected", projectID: "p", number: 1, title: "   ", sourceBranch: "s", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "empty sourceBranch rejected", projectID: "p", number: 1, title: "t", sourceBranch: "", targetBranch: "d", commitSHA: "x", expectErr: true},
		{name: "empty targetBranch rejected", projectID: "p", number: 1, title: "t", sourceBranch: "s", targetBranch: "", commitSHA: "x", expectErr: true},
		{name: "empty commitSHA rejected", projectID: "p", number: 1, title: "t", sourceBranch: "s", targetBranch: "d", commitSHA: "", expectErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := file.NewCommand(testCase.projectID, testCase.number, testCase.title, testCase.petitioner, testCase.sourceBranch, testCase.targetBranch, testCase.commitSHA)
			if testCase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, file.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
			assert.Equal(t, testCase.number, cmd.Number())
			assert.Equal(t, testCase.title, cmd.Title())
			assert.Equal(t, testCase.petitioner, cmd.Petitioner())
			assert.Equal(t, testCase.sourceBranch, cmd.SourceBranch())
			assert.Equal(t, testCase.targetBranch, cmd.TargetBranch())
			assert.Equal(t, testCase.commitSHA, cmd.CommitSHA())
		})
	}
}
