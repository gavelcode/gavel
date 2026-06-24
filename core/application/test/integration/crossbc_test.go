package appintegration_test

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
	"github.com/usegavel/gavel/core/application/casefile/submit"
	gscreate "github.com/usegavel/gavel/core/application/gavelspace/create"
	gsregister "github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	"github.com/usegavel/gavel/core/application/project/updatequalitygate"
	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	memgavelspace "github.com/usegavel/gavel/core/infrastructure/gavelspace/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

type crossBCFixture struct {
	projectRepo    *memproject.ProjectRepository
	caseFileRepo   *fakeCaseFileRepo
	gavelspaceRepo *memgavelspace.GavelspaceRepository

	createProject   *projectcreate.Handler
	updateGate      *updatequalitygate.Handler
	submitCaseFile  *submit.Handler
	createGS        *gscreate.Handler
	registerProject *gsregister.Handler
}

func newCrossBCFixture(t *testing.T) crossBCFixture {
	t.Helper()

	projectRepo := memproject.NewProjectRepository()
	caseFileRepo := newFakeCaseFileRepo()
	gavelspaceRepo := memgavelspace.NewGavelspaceRepository()

	createProject := projectcreate.NewHandler(projectRepo)
	updateGate := updatequalitygate.NewHandler(projectRepo)

	classifyH := classify.NewHandler(caseFileRepo)
	judgeH := judge.NewHandler(caseFileRepo, projectRepo)
	finalizeH := finalize.NewHandler(caseFileRepo, projectRepo, classifyH, judgeH, nil)
	createCF := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestEv := ingestevidence.NewHandler(caseFileRepo)
	submitH := submit.NewHandler(createCF, ingestEv, finalizeH)

	createGS := gscreate.NewHandler(gavelspaceRepo)
	registerH := gsregister.NewHandler(gavelspaceRepo)

	return crossBCFixture{
		projectRepo:    projectRepo,
		caseFileRepo:   caseFileRepo,
		gavelspaceRepo: gavelspaceRepo,

		createProject:   createProject,
		updateGate:      updateGate,
		submitCaseFile:  submitH,
		createGS:        createGS,
		registerProject: registerH,
	}
}

func TestCrossBC_ProjectGateControlsVerdict(t *testing.T) {
	ctx := context.Background()
	crossFixture := newCrossBCFixture(t)

	createCmd, err := projectcreate.NewCommand("gate-test", "Gate Test", "//gate/...")
	require.NoError(t, err)
	createRes, err := crossFixture.createProject.Execute(ctx, createCmd)
	require.NoError(t, err)
	projectID := createRes.ProjectID

	gateInput := updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype: "code_quality",
				Strategy: updatequalitygate.StrategyInput{
					Type: updatequalitygate.StrategyTypeZeroTolerance,
				},
			},
		},
	}
	gateCmd, err := updatequalitygate.NewCommand(projectID, gateInput)
	require.NoError(t, err)
	gateRes, err := crossFixture.updateGate.Execute(ctx, gateCmd)
	require.NoError(t, err)
	assert.True(t, gateRes.Changed)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-gate-1"),
	}
	submitCmd, err := submit.NewCommand(
		projectID, "abc123", "main",
		evidences, []string{"fp-gate-1"}, nil,
		finalize.ArchDeltaInput{}, nil, false, false,
		time.Now().UTC(),
	)
	require.NoError(t, err)

	result, err := crossFixture.submitCaseFile.Execute(ctx, submitCmd)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict.Outcome)
	require.NotEmpty(t, result.Verdict.Rulings)
	assert.False(t, result.Verdict.Rulings[0].Passed)
	assert.Contains(t, result.Verdict.Rulings[0].Detail, "1")

	emptyEvidences := []evidencedto.Evidence{
		{
			Subtype:     "code_quality",
			Source:      "test",
			CollectedAt: time.Now().UTC(),
			Findings:    nil,
		},
	}
	passCmd, err := submit.NewCommand(
		projectID, "def456", "main",
		emptyEvidences, nil, nil,
		finalize.ArchDeltaInput{}, nil, false, false,
		time.Now().UTC(),
	)
	require.NoError(t, err)

	passResult, err := crossFixture.submitCaseFile.Execute(ctx, passCmd)
	require.NoError(t, err)

	assert.Equal(t, "pass", passResult.Verdict.Outcome)
}

func TestCrossBC_GavelspaceProjectRegistration(t *testing.T) {
	ctx := context.Background()
	fixture := newCrossBCFixture(t)

	createCmd, err := projectcreate.NewCommand("gs-proj", "GS Project", "//gs/...")
	require.NoError(t, err)
	createRes, err := fixture.createProject.Execute(ctx, createCmd)
	require.NoError(t, err)
	projectID := createRes.ProjectID

	gsCmd, err := gscreate.NewCommand("test-monorepo")
	require.NoError(t, err)
	gsRes, err := fixture.createGS.Execute(ctx, gsCmd)
	require.NoError(t, err)
	assert.Equal(t, "test-monorepo", gsRes.Name)

	regCmd, err := gsregister.NewCommand("test-monorepo", projectID, "//gs/...")
	require.NoError(t, err)
	_, err = fixture.registerProject.Execute(ctx, regCmd)
	require.NoError(t, err)

	gsID, err := model.NewGavelspaceID("test-monorepo")
	require.NoError(t, err)
	gs, err := fixture.gavelspaceRepo.FindByName(ctx, gsID)
	require.NoError(t, err)

	projects := gs.Projects()
	require.Len(t, projects, 1)
	assert.Equal(t, projectID, projects[0].ID().String())
	assert.Equal(t, "//gs/...", projects[0].TargetPattern())

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-gs-1"),
		buildCoverageEvidence(200, 180),
	}
	submitCmd, err := submit.NewCommand(
		projectID, "aaa111", "main",
		evidences, []string{"fp-gs-1"}, nil,
		finalize.ArchDeltaInput{}, nil, false, false,
		time.Now().UTC(),
	)
	require.NoError(t, err)

	result, err := fixture.submitCaseFile.Execute(ctx, submitCmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.CaseFileID)
	assert.Equal(t, "pass", result.Verdict.Outcome)
	assert.Equal(t, 1, result.Counters.FindingsCount)
	assert.InDelta(t, 90.0, result.Counters.CoveragePercent, 0.001)
}
