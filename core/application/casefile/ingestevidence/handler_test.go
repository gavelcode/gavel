package ingestevidence_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

var testTenantExternal = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

var testTime = time.Date(2026, time.June, 10, 12, 0, 0, 0, time.UTC)

func TestExecuteAppendsEvidencesAndPersists(t *testing.T) {
	caseFiles := casefilememory.NewCaseFileRepository()
	projects := projectmemory.NewProjectRepository()
	cfID := seedCaseFile(t, caseFiles, projects)

	handler := ingestevidence.NewHandler(caseFiles)
	cmd, err := ingestevidence.NewCommand(cfID, []evidencedto.Evidence{newFindings(t)})
	require.NoError(t, err)

	res, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, res.EvidenceIDs, 1)
	assert.NotEmpty(t, res.Events)
}

func TestExecuteCaseFileNotFound(t *testing.T) {
	caseFiles := casefilememory.NewCaseFileRepository()

	handler := ingestevidence.NewHandler(caseFiles)
	cmd, _ := ingestevidence.NewCommand(uuid.NewString(), []evidencedto.Evidence{newFindings(t)})

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load case file")
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := ingestevidence.NewCommand("", []evidencedto.Evidence{newFindings(t)})
	require.ErrorIs(t, err, ingestevidence.ErrInvalidCommand)

	_, err = ingestevidence.NewCommand(uuid.NewString(), nil)
	require.ErrorIs(t, err, ingestevidence.ErrInvalidCommand)
}

func seedCaseFile(t *testing.T, caseFiles *casefilememory.CaseFileRepository, projects *projectmemory.ProjectRepository) string {
	t.Helper()
	project, err := projectmodel.NewProject(testTenantExternal, "p", "P", "//...")
	require.NoError(t, err)
	project.ClearEvents()
	require.NoError(t, projects.Save(context.Background(), project))

	h := createcasefile.NewHandler(caseFiles, projects)
	cmd, err := createcasefile.NewCommand(testTenantExternal.String(), project.ID().String(), "abc123", "main", testTime)
	require.NoError(t, err)
	res, err := h.Execute(context.Background(), cmd)
	require.NoError(t, err)
	return res.CaseFileID
}

func newFindings(t *testing.T) evidencedto.Evidence {
	t.Helper()
	return evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "golangci-lint",
		CollectedAt: testTime,
		Findings: []evidencedto.Finding{
			{
				Tool:          "golangci-lint",
				RuleID:        "errcheck",
				Severity:      "error",
				FilePath:      "main.go",
				Line:          1,
				Message:       "ignored error",
				FingerprintID: uuid.NewString(),
			},
		},
	}
}
