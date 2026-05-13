package classify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerExecuteFirstAnalysisAllNew(t *testing.T) {
	repo := newFakeCaseFileRepo()
	current := []finding.Finding{mustFinding(t, "fp-1"), mustFinding(t, "fp-2")}

	handler := classify.NewHandler(repo)
	cmd := mustCommand(t, uuid.NewString(), "main", current)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Len(t, result.Tracking.NewFindings, 2, "no previous fingerprints means all are new")
	assert.Empty(t, result.Tracking.ExistingFindings)
	assert.Equal(t, 0, result.Tracking.ResolvedCount)
}

func TestHandlerExecuteClassifiesNewExistingResolved(t *testing.T) {
	repo := newFakeCaseFileRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	repo.seedFingerprints(projectID, "main", []finding.FingerprintID{
		mustFingerprintID(t, "fp-existing"),
		mustFingerprintID(t, "fp-resolved"),
	})

	current := []finding.Finding{
		mustFinding(t, "fp-new"),
		mustFinding(t, "fp-existing"),
	}

	handler := classify.NewHandler(repo)
	cmd := mustCommand(t, projectID.String(), "main", current)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Len(t, result.Tracking.NewFindings, 1)
	assert.Equal(t, "fp-new", result.Tracking.NewFindings[0].FingerprintID)

	assert.Len(t, result.Tracking.ExistingFindings, 1)
	assert.Equal(t, "fp-existing", result.Tracking.ExistingFindings[0].FingerprintID)

	assert.Equal(t, 1, result.Tracking.ResolvedCount, "fp-resolved present in baseline but absent in current")
}

func TestHandlerExecuteEmptyCurrentFindings(t *testing.T) {
	repo := newFakeCaseFileRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	repo.seedFingerprints(projectID, "main", []finding.FingerprintID{mustFingerprintID(t, "fp-1")})

	handler := classify.NewHandler(repo)
	cmd := mustCommand(t, projectID.String(), "main", nil)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Empty(t, result.Tracking.NewFindings)
	assert.Empty(t, result.Tracking.ExistingFindings)
	assert.Equal(t, 1, result.Tracking.ResolvedCount, "all baseline findings are resolved when current is empty")
}

func TestHandlerExecuteRepositoryErrorPropagated(t *testing.T) {
	repo := newFakeCaseFileRepo()
	repo.fingerprintsErr = errors.New("db down")

	handler := classify.NewHandler(repo)
	cmd := mustCommand(t, uuid.NewString(), "main", nil)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteInvalidProjectIDRejected(t *testing.T) {
	repo := newFakeCaseFileRepo()
	handler := classify.NewHandler(repo)

	cmd, err := classify.NewCommand("not-a-uuid", "main", nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err, "non-UUID project id is rejected at the domain boundary")
	assert.ErrorIs(t, err, projectmodel.ErrInvalidProject)
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { classify.NewHandler(nil) })
}

func mustCommand(t *testing.T, projectID, branch string, findings []finding.Finding) classify.Command {
	t.Helper()
	cmd, err := classify.NewCommand(projectID, branch, findings)
	require.NoError(t, err)
	return cmd
}

func mustFingerprintID(t *testing.T, value string) finding.FingerprintID {
	t.Helper()
	fp, err := finding.NewFingerprintID(value)
	require.NoError(t, err)
	return fp
}
