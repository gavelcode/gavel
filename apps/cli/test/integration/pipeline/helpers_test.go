package pipeline_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/infrastructure/casefile/lcov"
	memcasefile "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

const (
	testCommitSHA = "abc123def456"
	testBranch    = "main"
	testWorkspace = "/tmp/test-workspace"
)

type localFixture struct {
	deps     pipeline.Deps
	cfRepo   *memcasefile.CaseFileRepository
	projRepo *memproject.ProjectRepository
}

func newLocalFixture(t *testing.T) localFixture {
	t.Helper()

	cfRepo := memcasefile.NewCaseFileRepository()
	projRepo := memproject.NewProjectRepository()

	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	createCF := createcasefile.NewHandler(cfRepo, projRepo)
	ingestEv := ingestevidence.NewHandler(cfRepo)
	finalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)
	submitH := submit.NewHandler(createCF, ingestEv, finalizeH)

	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()

	findingsH := ingestfind.NewHandler(map[string]ingestfind.Parser{
		"sarif": sarifParser,
	})
	coverageH := ingestcov.NewHandler(map[string]ingestcov.Parser{
		"lcov": lcovParser,
	})

	return localFixture{
		deps: pipeline.Deps{
			Log:          slog.Default(),
			Submit:       submitH,
			Findings:     findingsH,
			Coverage:     coverageH,
			ServerClient: nil,
		},
		cfRepo:   cfRepo,
		projRepo: projRepo,
	}
}

func buildCollectedEvidence(t *testing.T, fingerprints []string, coveragePct float64) collectevidence.Result {
	t.Helper()

	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	findingDTOs := make([]evidencedto.Finding, 0, len(fingerprints))
	for idx, fingerprint := range fingerprints {
		findingDTOs = append(findingDTOs, evidencedto.Finding{
			Tool:          "test-linter",
			RuleID:        fmt.Sprintf("rule-%d", idx),
			Severity:      "error",
			FilePath:      fmt.Sprintf("file%d.go", idx),
			Line:          idx + 1,
			Message:       fmt.Sprintf("finding %d", idx),
			FingerprintID: fingerprint,
		})
	}

	var evidences []evidencedto.Evidence

	evidences = append(evidences, evidencedto.Evidence{
		Subtype:     evidence.SubtypeCodeQuality.String(),
		Source:      "test-linter",
		CollectedAt: now,
		Findings:    findingDTOs,
	})

	totalLines := 1000
	coveredLines := int(coveragePct * float64(totalLines) / 100)
	evidences = append(evidences, evidencedto.Evidence{
		Subtype:     evidence.SubtypeCoverage.String(),
		Source:      "test-coverage",
		CollectedAt: now,
		Coverage: &evidencedto.Coverage{
			TotalLines:   totalLines,
			CoveredLines: coveredLines,
		},
	})

	return collectevidence.Result{
		Evidences:     evidences,
		FindingsCount: len(fingerprints),
		CovPercent:    coveragePct,
		Fingerprints:  fingerprints,
		Findings:      findingDTOs,
	}
}

type projectOption func(t *testing.T, p *projectmodel.Project)

func withGate(gate qualitygate.Gate) projectOption {
	return func(_ *testing.T, p *projectmodel.Project) {
		p.UpdateQualityGate(gate, time.Now().UTC())
	}
}

func withBaseline(branch string, fingerprints []string) projectOption {
	return func(_ *testing.T, p *projectmodel.Project) {
		p.UpdateBaseline(branch, fingerprints, nil, nil, nil)
	}
}

func mustSeedProject(t *testing.T, repo *memproject.ProjectRepository, key string, opts ...projectOption) projectmodel.Project {
	t.Helper()

	project, err := projectmodel.NewProject(tenant.LocalTenantID, key, "Test "+key, "//...")
	require.NoError(t, err)

	for _, opt := range opts {
		opt(t, &project)
	}

	project.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}

func mustBuildZeroToleranceGate(t *testing.T) qualitygate.Gate {
	t.Helper()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance())
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	return gate
}
