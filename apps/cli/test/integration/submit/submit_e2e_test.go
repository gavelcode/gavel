package submit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func TestSubmitFlow_HappyPathProducesVerdict(t *testing.T) {
	ctx := context.Background()
	db := testkit.TestDB(t)

	projectRepo := projectpostgres.NewRepository(db)
	caseFileRepo := casefilepostgres.NewRepository(db)

	createH := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestH := ingestevidence.NewHandler(caseFileRepo)
	judgeH := judge.NewHandler(caseFileRepo, projectRepo)
	classifyH := classify.NewHandler(caseFileRepo)
	finalizeH := finalize.NewHandler(caseFileRepo, projectRepo, classifyH, judgeH, nil)

	project := seedProject(t, ctx, projectRepo, "dogfood")

	createCmd, err := createcasefile.NewCommand(tenant.LocalTenantID.String(), project.ID().String(), "abc123", "main", time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	createRes, err := createH.Execute(ctx, createCmd)
	require.NoError(t, err)
	require.NotEmpty(t, createRes.CaseFileID, "createcasefile must return a case_file_id")

	ingestCmd, err := ingestevidence.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID, []evidencedto.Evidence{cleanFinding(t)})
	require.NoError(t, err)
	ingestRes, err := ingestH.Execute(ctx, ingestCmd)
	require.NoError(t, err)
	require.Len(t, ingestRes.EvidenceIDs, 1, "ingestevidence must return one id per evidence")

	finalizeCmd, err := finalize.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID)
	require.NoError(t, err)
	finalizeRes, err := finalizeH.Execute(ctx, finalizeCmd)
	require.NoError(t, err, "finalize must succeed when classify + judge have everything they need")

	assert.Equal(t, createRes.CaseFileID, finalizeRes.CaseFileID)
	assert.Equal(t, "pass", finalizeRes.Verdict.Outcome, "a clean finding set with no quality-gate violations must pass")
	assert.Equal(t, 1, finalizeRes.Counters.FindingsCount, "the counters must reflect the single finding ingested")
	assert.True(t, finalizeRes.Counters.HasTracking, "classify always runs unless fresh_evaluation=true")
}

func TestSubmitFlow_FreshEvaluationSkipsClassify(t *testing.T) {
	ctx := context.Background()
	db := testkit.TestDB(t)
	projectRepo := projectpostgres.NewRepository(db)
	caseFileRepo := casefilepostgres.NewRepository(db)
	createH := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestH := ingestevidence.NewHandler(caseFileRepo)
	judgeH := judge.NewHandler(caseFileRepo, projectRepo)
	classifyH := classify.NewHandler(caseFileRepo)
	finalizeH := finalize.NewHandler(caseFileRepo, projectRepo, classifyH, judgeH, nil)

	project := seedProject(t, ctx, projectRepo, "fresh-dogfood")

	createCmd, err := createcasefile.NewCommand(
		tenant.LocalTenantID.String(),
		project.ID().String(),
		"abc123", "main",
		time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		createcasefile.WithFreshEvaluation(),
	)
	require.NoError(t, err)
	createRes, err := createH.Execute(ctx, createCmd)
	require.NoError(t, err)

	ingestCmd, _ := ingestevidence.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID, []evidencedto.Evidence{cleanFinding(t)})
	_, err = ingestH.Execute(ctx, ingestCmd)
	require.NoError(t, err)

	finalizeCmd, _ := finalize.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID)
	finalizeRes, err := finalizeH.Execute(ctx, finalizeCmd)
	require.NoError(t, err)

	assert.False(t, finalizeRes.Counters.HasTracking, "fresh_evaluation=true must skip classify")
	assert.Equal(t, "pass", finalizeRes.Verdict.Outcome)
}

func TestSubmitFlow_DoubleFinalizeIsRejected(t *testing.T) {
	ctx := context.Background()
	db := testkit.TestDB(t)
	projectRepo := projectpostgres.NewRepository(db)
	caseFileRepo := casefilepostgres.NewRepository(db)
	createH := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestH := ingestevidence.NewHandler(caseFileRepo)
	judgeH := judge.NewHandler(caseFileRepo, projectRepo)
	classifyH := classify.NewHandler(caseFileRepo)
	finalizeH := finalize.NewHandler(caseFileRepo, projectRepo, classifyH, judgeH, nil)

	project := seedProject(t, ctx, projectRepo, "double-finalize")
	createCmd, _ := createcasefile.NewCommand(tenant.LocalTenantID.String(), project.ID().String(), "abc123", "main", time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	createRes, err := createH.Execute(ctx, createCmd)
	require.NoError(t, err)

	ingestCmd, _ := ingestevidence.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID, []evidencedto.Evidence{cleanFinding(t)})
	_, err = ingestH.Execute(ctx, ingestCmd)
	require.NoError(t, err)

	finalizeCmd, _ := finalize.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID)
	_, err = finalizeH.Execute(ctx, finalizeCmd)
	require.NoError(t, err)

	_, err = finalizeH.Execute(ctx, finalizeCmd)
	require.Error(t, err, "the aggregate must refuse a second judge call")
}

func seedProject(t *testing.T, ctx context.Context, repo *projectpostgres.Repository, key string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(tenant.LocalTenantID, key, "Dogfood "+key, "//...")
	require.NoError(t, err)
	p.ClearEvents()
	require.NoError(t, repo.Save(ctx, p))
	return p
}

func cleanFinding(t *testing.T) evidencedto.Evidence {
	t.Helper()
	return evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "golangci-lint",
		CollectedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		Findings: []evidencedto.Finding{{
			Tool:          "golangci-lint",
			RuleID:        "errcheck",
			Severity:      "note",
			FilePath:      "main.go",
			Line:          1,
			Message:       "ignored error",
			FingerprintID: "11111111-1111-1111-1111-111111111111",
		}},
	}
}
