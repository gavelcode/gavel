package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var (
	testTime   = time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
)

func validProjectID(t *testing.T) projectmodel.ProjectID {
	t.Helper()
	id := projectmodel.NewProjectID(uuid.New())
	return id
}

func TestFilePleading(t *testing.T) {
	projectID := validProjectID(t)

	tests := []struct {
		name         string
		number       int
		title        string
		petitioner   string
		sourceBranch string
		targetBranch string
		commitSHA    string
		expectErr    bool
	}{
		{
			name:         "valid pleading",
			number:       42,
			title:        "Add login endpoint",
			petitioner:   "alice",
			sourceBranch: "feature/login",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    false,
		},
		{
			name:         "zero number rejected",
			number:       0,
			title:        "Title",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "negative number rejected",
			number:       -1,
			title:        "Title",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "empty title rejected",
			number:       42,
			title:        "",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "blank title rejected",
			number:       42,
			title:        "   ",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "empty source branch rejected",
			number:       42,
			title:        "Title",
			petitioner:   "alice",
			sourceBranch: "",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "empty target branch rejected",
			number:       42,
			title:        "Title",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "",
			commitSHA:    "abc123",
			expectErr:    true,
		},
		{
			name:         "empty commitSHA rejected",
			number:       42,
			title:        "Title",
			petitioner:   "alice",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "",
			expectErr:    true,
		},
		{
			name:         "empty petitioner allowed",
			number:       42,
			title:        "Title",
			petitioner:   "",
			sourceBranch: "feature",
			targetBranch: "main",
			commitSHA:    "abc123",
			expectErr:    false,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			plea, err := model.FilePleading(testTenant,
				projectID, tcase.number, tcase.title, tcase.petitioner,
				tcase.sourceBranch, tcase.targetBranch, tcase.commitSHA,
			)

			if tcase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidPleading)
				return
			}

			require.NoError(t, err)
			assert.True(t, testTenant.Equal(plea.TenantID()))
			assert.True(t, plea.ProjectID().Equal(projectID))
			assert.Equal(t, tcase.number, plea.Number())
			assert.Equal(t, tcase.title, plea.Title())
			assert.Equal(t, tcase.petitioner, plea.Petitioner())
			assert.Equal(t, tcase.sourceBranch, plea.SourceBranch())
			assert.Equal(t, tcase.targetBranch, plea.TargetBranch())
			assert.Equal(t, tcase.commitSHA, plea.CommitSHA())
			assert.True(t, plea.Status().Equal(model.StatusOpen))
		})
	}
}

func TestFilePleadingGeneratesUniqueIDs(t *testing.T) {
	projectID := validProjectID(t)

	pleadA, err := model.FilePleading(testTenant, projectID, 1, "t1", "alice", "find1", "main", "sha1")
	require.NoError(t, err)

	pleadB, err := model.FilePleading(testTenant, projectID, 2, "t2", "alice", "find2", "main", "sha2")
	require.NoError(t, err)

	assert.False(t, pleadA.ID().Equal(pleadB.ID()))
}

func TestPleadingMarkMerged(t *testing.T) {
	projectID := validProjectID(t)
	plea, err := model.FilePleading(testTenant, projectID, 7, "t", "alice", "src", "dst", "sha")
	require.NoError(t, err)

	require.NoError(t, plea.MarkMerged(testTime))

	assert.True(t, plea.Status().Equal(model.StatusMerged))

	events := plea.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.Merged)
	require.True(t, ok)
	assert.True(t, plea.ID().Equal(evt.PleadingID()))
	assert.Equal(t, testTime, evt.OccurredAt())
	assert.Equal(t, model.EventNameMerged, evt.EventName())
}

func TestPleadingMarkClosed(t *testing.T) {
	projectID := validProjectID(t)
	plea, err := model.FilePleading(testTenant, projectID, 7, "t", "alice", "src", "dst", "sha")
	require.NoError(t, err)

	require.NoError(t, plea.MarkClosed(testTime))

	assert.True(t, plea.Status().Equal(model.StatusClosed))

	events := plea.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.Closed)
	require.True(t, ok)
	assert.True(t, plea.ID().Equal(evt.PleadingID()))
	assert.Equal(t, testTime, evt.OccurredAt())
	assert.Equal(t, model.EventNameClosed, evt.EventName())
}

func TestPleadingRejectsIllegalTransitions(t *testing.T) {
	projectID := validProjectID(t)

	tests := []struct {
		name   string
		setup  func(*model.Pleading) error
		mutate func(*model.Pleading) error
	}{
		{
			name:   "merge after merge",
			setup:  func(plea *model.Pleading) error { return plea.MarkMerged(testTime) },
			mutate: func(plea *model.Pleading) error { return plea.MarkMerged(testTime) },
		},
		{
			name:   "close after merge",
			setup:  func(plea *model.Pleading) error { return plea.MarkMerged(testTime) },
			mutate: func(plea *model.Pleading) error { return plea.MarkClosed(testTime) },
		},
		{
			name:   "merge after close",
			setup:  func(plea *model.Pleading) error { return plea.MarkClosed(testTime) },
			mutate: func(plea *model.Pleading) error { return plea.MarkMerged(testTime) },
		},
		{
			name:   "close after close",
			setup:  func(plea *model.Pleading) error { return plea.MarkClosed(testTime) },
			mutate: func(plea *model.Pleading) error { return plea.MarkClosed(testTime) },
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			plea, err := model.FilePleading(testTenant, projectID, 1, "t", "alice", "src", "dst", "sha")
			require.NoError(t, err)

			require.NoError(t, tcase.setup(&plea))
			plea.ClearEvents()

			err = tcase.mutate(&plea)
			require.Error(t, err)
			assert.ErrorIs(t, err, model.ErrInvalidTransition)
			assert.Empty(t, plea.Events(), "no event recorded on illegal transition")
		})
	}
}

func TestReconstitutePleading(t *testing.T) {
	projectID := validProjectID(t)
	validID := model.NewPleadingID(uuid.New())

	tests := []struct {
		name         string
		id           model.PleadingID
		projectID    projectmodel.ProjectID
		number       int
		title        string
		petitioner   string
		sourceBranch string
		targetBranch string
		commitSHA    string
		status       model.Status
		expectErr    bool
	}{
		{
			name:         "valid reconstitution open",
			id:           validID,
			projectID:    projectID,
			number:       1,
			title:        "t",
			petitioner:   "alice",
			sourceBranch: "src",
			targetBranch: "dst",
			commitSHA:    "sha",
			status:       model.StatusOpen,
			expectErr:    false,
		},
		{
			name:         "valid reconstitution merged",
			id:           validID,
			projectID:    projectID,
			number:       1,
			title:        "t",
			petitioner:   "alice",
			sourceBranch: "src",
			targetBranch: "dst",
			commitSHA:    "sha",
			status:       model.StatusMerged,
			expectErr:    false,
		},
		{
			name:         "invariants still enforced on reconstitute",
			id:           validID,
			projectID:    projectID,
			number:       0,
			title:        "t",
			petitioner:   "alice",
			sourceBranch: "src",
			targetBranch: "dst",
			commitSHA:    "sha",
			status:       model.StatusOpen,
			expectErr:    true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			plea, err := model.ReconstitutePleading(
				tcase.id, testTenant, tcase.projectID, tcase.number, tcase.title, tcase.petitioner,
				tcase.sourceBranch, tcase.targetBranch, tcase.commitSHA, tcase.status,
			)

			if tcase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidPleading)
				return
			}

			require.NoError(t, err)
			assert.True(t, tcase.id.Equal(plea.ID()))
			assert.True(t, testTenant.Equal(plea.TenantID()))
			assert.True(t, tcase.projectID.Equal(plea.ProjectID()))
			assert.Equal(t, tcase.number, plea.Number())
			assert.Equal(t, tcase.title, plea.Title())
			assert.Equal(t, tcase.petitioner, plea.Petitioner())
			assert.Equal(t, tcase.sourceBranch, plea.SourceBranch())
			assert.Equal(t, tcase.targetBranch, plea.TargetBranch())
			assert.Equal(t, tcase.commitSHA, plea.CommitSHA())
			assert.True(t, tcase.status.Equal(plea.Status()))
			assert.Empty(t, plea.Events(), "reconstitution must not record events")
		})
	}
}
