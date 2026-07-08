package v1integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	iamchangepw "github.com/usegavel/gavel/core/application/iam/changepassword"
	iamcreateuser "github.com/usegavel/gavel/core/application/iam/createuser"
	iamissuetoken "github.com/usegavel/gavel/core/application/iam/issuetoken"
	iamlistmytokens "github.com/usegavel/gavel/core/application/iam/listmytokens"
	iamlogin "github.com/usegavel/gavel/core/application/iam/login"
	iamresolveprincipal "github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	iamrevoketoken "github.com/usegavel/gavel/core/application/iam/revoketoken"
	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	"github.com/usegavel/gavel/core/application/project/getbaseline"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	"github.com/usegavel/gavel/core/application/project/projectview"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memcasefile "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
	casefilev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
	iamv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
	opsv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
	pleadingv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
	projectv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/project"

	apiv1 "github.com/usegavel/gavel/apps/server/internal/api/v1"
)

type repoBackedCaseFileFinder struct {
	repo *memcasefile.CaseFileRepository
}

func (finder *repoBackedCaseFileFinder) GetByID(ctx context.Context, _, id string) (*casefileget.CaseFileDetail, error) {
	cfID, err := casefile.ParseCaseFileID(id)
	if err != nil {
		return nil, err
	}
	caseFile, err := finder.repo.FindByID(ctx, tenant.LocalTenantID, cfID)
	if err != nil {
		return nil, err
	}
	detail := &casefileget.CaseFileDetail{
		ID:        caseFile.ID().String(),
		ProjectID: caseFile.ProjectID().String(),
		CommitSHA: caseFile.CommitSHA(),
		Branch:    caseFile.Branch(),
		StartedAt: caseFile.StartedAt(),
		CreatedAt: caseFile.StartedAt(),
	}
	if v, ok := caseFile.Verdict(); ok {
		detail.VerdictOutcome = v.Outcome().String()
		for _, r := range v.Rulings() {
			detail.Rulings = append(detail.Rulings, casefileget.RulingView{
				Subtype: r.Subtype().String(),
				Passed:  r.Passed(),
				Detail:  r.Detail(),
			})
		}
	}
	for _, evItem := range caseFile.Evidences() {
		detail.Evidences = append(detail.Evidences, casefileget.EvidenceSummary{
			ID:          evItem.ID().String(),
			Subtype:     evItem.Subtype().String(),
			Source:      evItem.Source(),
			CollectedAt: evItem.CollectedAt(),
		})
		if fc, ok := evItem.Content().(finding.Content); ok {
			detail.TotalFindings += len(fc.Findings())
		}
		if cc, ok := evItem.Content().(coverage.Content); ok && cc.TotalLines() > 0 {
			pct := float64(cc.CoveredLines()) / float64(cc.TotalLines()) * 100
			detail.CoveragePercent = &pct
		}
	}
	return detail, nil
}

func (finder *repoBackedCaseFileFinder) ListByProject(_ context.Context, _, _, _ string, _, _ int) ([]casefilelist.CaseFileSummary, int, error) {
	return nil, 0, nil
}

func (finder *repoBackedCaseFileFinder) List(_ context.Context, _ string, _ findinglist.Filters, _, _ int) ([]findinglist.FindingView, int, error) {
	return nil, 0, nil
}

type repoBackedProjectFinder struct {
	repo *memproject.ProjectRepository
}

func (finder *repoBackedProjectFinder) GetByKey(ctx context.Context, _ tenant.TenantID, key string) (*projectview.ProjectDetail, error) {
	project, err := finder.repo.FindByKey(ctx, tenant.LocalTenantID, key)
	if err != nil {
		return nil, err
	}
	return &projectview.ProjectDetail{
		ID:            project.ID().String(),
		Key:           project.Key(),
		Name:          project.Name(),
		DefaultBranch: "main",
		TargetPattern: project.TargetPattern(),
		CreatedAt:     testNow,
	}, nil
}

func (finder *repoBackedProjectFinder) List(_ context.Context, _ tenant.TenantID, _, _ int) ([]projectlist.ProjectSummary, int, error) {
	return nil, 0, nil
}

func newServerModeServer(t *testing.T) (*httptest.Server, *memiam.SessionRepository) {
	t.Helper()
	ctx := context.Background()

	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	tokens := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secrets := memiam.NewFakeSecretGenerator()

	slug, _ := tenant.NewSlug("default")
	defaultTenant, _ := tenant.NewTenant(slug, "Default", testNow)
	defaultTenant.ClearEvents()
	require.NoError(t, tenants.Save(ctx, defaultTenant))

	email, _ := user.NewEmail(adminEmail)
	hash, _ := hasher.Hash(adminPassword)
	admin, _ := user.NewUser(defaultTenant.ID(), email, "Admin", user.RoleAdmin, hash, false, testNow)
	admin.ClearEvents()
	require.NoError(t, users.Save(ctx, admin))

	loginH := iamlogin.NewHandler(tenants, users, sessions, hasher, secrets)
	changePwH := iamchangepw.NewHandler(users, sessions, hasher)
	issueTokenH := iamissuetoken.NewHandler(users, tokens, secrets)
	revokeTokenH := iamrevoketoken.NewHandler(tokens)
	listTokensH := iamlistmytokens.NewHandler(tokens)
	createUserH := iamcreateuser.NewHandler(tenants, users, hasher)
	resolveH := iamresolveprincipal.NewHandler(users, sessions, tokens)

	clock := func() time.Time { return testNow }

	projRepo := memproject.NewProjectRepository()
	projFinder := &repoBackedProjectFinder{repo: projRepo}
	projListH := projectlist.NewHandler(projFinder)
	projGetH := projectgetbykey.NewHandler(projFinder)
	projCreateH := projectcreate.NewHandler(projRepo)

	cfRepo := memcasefile.NewCaseFileRepository()
	cfFinder := &repoBackedCaseFileFinder{repo: cfRepo}
	cfCreateH := createcasefile.NewHandler(cfRepo, projRepo)
	cfIngestH := ingestevidence.NewHandler(cfRepo)
	cfGetH := casefileget.NewHandler(cfFinder)
	cfListH := casefilelist.NewHandler(cfFinder)
	findListH := findinglist.NewHandler(cfFinder)
	judgeH := corejudge.NewHandler(cfRepo, projRepo)
	classifyH := classify.NewHandler(cfRepo)
	cfFinalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)

	plMemRepo := newPleadingMemRepo()
	plFileH := pleadingfile.NewHandler(plMemRepo)

	cookie := auth.SessionCookie{Name: sessionCookieName, Secure: false, TTL: 24 * time.Hour}
	authMw := auth.NewMiddleware(resolveH, cookie, clock)

	server := &apiv1.Server{
		CaseFileHandler: casefilev1.New(casefilev1.Deps{
			ListCaseFiles:       cfListH,
			GetCaseFile:         cfGetH,
			ListFindings:        findListH,
			GetBaseline:         getbaseline.NewHandler(projFinder, projRepo),
			CreateCaseFile:      cfCreateH,
			IngestEvidence:      cfIngestH,
			FinalizeCaseFile:    cfFinalizeH,
			ResolveProjectByKey: projGetH,
			Now:                 clock,
		}),
		IAMHandler: iamv1.New(iamv1.Deps{
			Login:          loginH,
			ChangePassword: changePwH,
			IssueToken:     issueTokenH,
			RevokeToken:    revokeTokenH,
			ListMyTokens:   listTokensH,
			CreateUser:     createUserH,
			Cookie:         cookie,
			DefaultTenant:  "default",
			Now:            clock,
		}),
		OpsHandler: opsv1.New(),
		PleadingHandler: pleadingv1.New(pleadingv1.Deps{
			FilePleading:        plFileH,
			ResolveProjectByKey: projGetH,
		}),
		ProjectHandler: projectv1.New(projectv1.Deps{
			ListProjects:  projListH,
			CreateProject: projCreateH,
			GetProject:    projGetH,
		}),
	}
	mux := apiv1.NewMux(server, authMw)

	root := chi.NewRouter()
	root.Mount("/api/v1", mux)
	ts := httptest.NewServer(root)
	t.Cleanup(ts.Close)
	return ts, sessions
}

func serverModeLogin(t *testing.T, testServer *httptest.Server) string {
	t.Helper()
	// Login to get session cookie, then issue token
	loginReq := `{"email":"` + adminEmail + `","password":"` + adminPassword + `"}`
	resp, err := testServer.Client().Post(testServer.URL+"/api/v1/sessions", "application/json", bytes.NewBufferString(loginReq))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, 200, resp.StatusCode)

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie)

	tokenReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/me/tokens", bytes.NewBufferString(`{"name":"ci","scopes":["ingest","project_sync"]}`))
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.AddCookie(sessionCookie)
	tokenResp, err := testServer.Client().Do(tokenReq)
	require.NoError(t, err)
	defer func() { _ = tokenResp.Body.Close() }()
	require.Equal(t, 201, tokenResp.StatusCode)

	var tokenResult struct{ Token string }
	require.NoError(t, json.NewDecoder(tokenResp.Body).Decode(&tokenResult))
	return tokenResult.Token
}

func TestServerMode_SubmitAndFinalize(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	detail, err := cli.CreateProject(ctx, "myapp", "My App", "//app/...")
	require.NoError(t, err)
	require.Equal(t, "myapp", detail.Key)

	caseFileID, err := cli.OpenCaseFile(ctx, detail.ID, "abc123", "main", false)
	require.NoError(t, err)
	require.NotEmpty(t, caseFileID)

	evidence := apiclient.EvidenceToWire(evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "golangci-lint",
		CollectedAt: testNow,
		Findings: []evidencedto.Finding{{
			Tool: "golangci-lint", RuleID: "errcheck", Severity: "warning",
			FilePath: "main.go", Line: 42, Message: "error return not checked", FingerprintID: "fp-001",
		}},
	})
	_, err = cli.IngestCaseFileEvidence(ctx, caseFileID, evidence)
	require.NoError(t, err)

	verdict := apiclient.VerdictResult{
		Outcome:     "pass",
		EvaluatedAt: testNow,
	}
	counters := apiclient.CountersResult{FindingsCount: 1}
	result, err := cli.FinalizeCaseFileWithVerdict(ctx, caseFileID, verdict, counters)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)
	assert.Equal(t, 1, result.Counters.FindingsCount)
}

func TestServerMode_AutoCreateAndFetch(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	_, err := cli.FetchProject(ctx, "nonexistent")
	require.Error(t, err)

	detail, err := cli.CreateProject(ctx, "synced", "Synced", "//synced/...")
	require.NoError(t, err)
	assert.Equal(t, "synced", detail.Key)

	fetched, err := cli.FetchProject(ctx, "synced")
	require.NoError(t, err)
	assert.Equal(t, detail.ID, fetched.ID)
}

func TestServerMode_FilePleading(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	_, err := cli.CreateProject(ctx, "prtest", "PR Test", "//pr/...")
	require.NoError(t, err)

	pleadingID, err := cli.FilePleading(ctx, "prtest", 42, "Fix the bug", "dev", "feature/fix", "main", "abc123")
	require.NoError(t, err)
	require.NotEmpty(t, pleadingID)
}

func TestServerMode_PrecomputedVerdictFail(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	detail, err := cli.CreateProject(ctx, "precomp", "Precomputed", "//precomp/...")
	require.NoError(t, err)

	caseFileID, err := cli.OpenCaseFile(ctx, detail.ID, "abc123", "main", false)
	require.NoError(t, err)

	evidence := apiclient.EvidenceToWire(evidencedto.Evidence{
		Subtype: "code_quality", Source: "golangci-lint", CollectedAt: testNow,
		Findings: []evidencedto.Finding{{
			Tool: "golangci-lint", RuleID: "errcheck", Severity: "error",
			FilePath: "main.go", Line: 42, Message: "unchecked error", FingerprintID: "fp-001",
		}},
	})
	_, err = cli.IngestCaseFileEvidence(ctx, caseFileID, evidence)
	require.NoError(t, err)

	verdict := apiclient.VerdictResult{
		Outcome:     "fail",
		EvaluatedAt: testNow,
		Rulings: []apiclient.RulingResult{
			{Subtype: "code_quality", Passed: false, Detail: "1 error (max 0)"},
		},
	}
	counters := apiclient.CountersResult{
		FindingsCount:   1,
		CoveragePercent: 0,
		NewCount:        1,
		ExistingCount:   0,
		ResolvedCount:   0,
		HasTracking:     false,
	}
	result, err := cli.FinalizeCaseFileWithVerdict(ctx, caseFileID, verdict, counters)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict.Outcome, "server must store the pre-computed verdict, not re-evaluate")
	require.Len(t, result.Verdict.Rulings, 1)
	assert.Equal(t, "code_quality", result.Verdict.Rulings[0].Subtype)
	assert.False(t, result.Verdict.Rulings[0].Passed)
	assert.Equal(t, "1 error (max 0)", result.Verdict.Rulings[0].Detail)
	assert.Equal(t, 1, result.Counters.FindingsCount)
}

func TestServerMode_RoundTrip_DashboardSeesPrecomputedVerdict(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	detail, err := cli.CreateProject(ctx, "roundtrip", "Round Trip", "//rt/...")
	require.NoError(t, err)

	caseFileID, err := cli.OpenCaseFile(ctx, detail.ID, "commit-rt", "main", false)
	require.NoError(t, err)

	evidence := apiclient.EvidenceToWire(evidencedto.Evidence{
		Subtype: "code_quality", Source: "golangci-lint", CollectedAt: testNow,
		Findings: []evidencedto.Finding{
			{Tool: "golangci-lint", RuleID: "errcheck", Severity: "error", FilePath: "main.go", Line: 10, Message: "unchecked error", FingerprintID: "fp-rt-1"},
			{Tool: "golangci-lint", RuleID: "unused", Severity: "warning", FilePath: "util.go", Line: 5, Message: "unused var", FingerprintID: "fp-rt-2"},
		},
	})
	_, err = cli.IngestCaseFileEvidence(ctx, caseFileID, evidence)
	require.NoError(t, err)

	covEv := apiclient.EvidenceToWire(evidencedto.Evidence{
		Subtype: "coverage", Source: "go-coverage", CollectedAt: testNow,
		Coverage: &evidencedto.Coverage{TotalLines: 200, CoveredLines: 170},
	})
	_, err = cli.IngestCaseFileEvidence(ctx, caseFileID, covEv)
	require.NoError(t, err)

	verdict := apiclient.VerdictResult{
		Outcome:     "fail",
		EvaluatedAt: testNow,
		Rulings: []apiclient.RulingResult{
			{Subtype: "code_quality", Passed: false, Detail: "2 errors (max 0)"},
			{Subtype: "coverage", Passed: true, Detail: "85.0% (min 80%)"},
		},
	}
	counters := apiclient.CountersResult{
		FindingsCount:   2,
		CoveragePercent: 85.0,
		NewCount:        1,
		ExistingCount:   1,
		ResolvedCount:   3,
		HasTracking:     true,
	}
	_, err = cli.FinalizeCaseFileWithVerdict(ctx, caseFileID, verdict, counters)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/casefiles/"+caseFileID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, 200, resp.StatusCode)

	var caseFileView struct {
		VerdictOutcome   string   `json:"verdict_outcome"`
		TotalFindings    int      `json:"total_findings"`
		NewFindings      int      `json:"new_findings"`
		ExistingFindings int      `json:"existing_findings"`
		ResolvedFindings int      `json:"resolved_findings"`
		CoveragePercent  *float64 `json:"coverage_percent"`
		Rulings          []struct {
			Subtype string `json:"subtype"`
			Passed  bool   `json:"passed"`
			Detail  string `json:"detail"`
		} `json:"rulings"`
		Evidences []struct {
			Subtype string `json:"subtype"`
		} `json:"evidences"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&caseFileView))

	assert.Equal(t, "fail", caseFileView.VerdictOutcome, "dashboard must show the CLI's verdict, not a server-side re-evaluation")
	assert.Equal(t, 2, caseFileView.TotalFindings)
	require.NotNil(t, caseFileView.CoveragePercent)
	assert.InDelta(t, 85.0, *caseFileView.CoveragePercent, 0.1)

	require.Len(t, caseFileView.Rulings, 2)
	assert.Equal(t, "code_quality", caseFileView.Rulings[0].Subtype)
	assert.False(t, caseFileView.Rulings[0].Passed)
	assert.Equal(t, "2 errors (max 0)", caseFileView.Rulings[0].Detail)
	assert.Equal(t, "coverage", caseFileView.Rulings[1].Subtype)
	assert.True(t, caseFileView.Rulings[1].Passed)

	require.Len(t, caseFileView.Evidences, 2, "both evidence pieces must be stored")
}

func TestServerMode_BaselineDelta(t *testing.T) {
	testServer, _ := newServerModeServer(t)
	token := serverModeLogin(t, testServer)
	cli, _ := apiclient.New(testServer.URL, token)
	ctx := context.Background()

	detail, err := cli.CreateProject(ctx, "deltatest", "Delta", "//delta/...")
	require.NoError(t, err)

	evidence := apiclient.EvidenceToWire(evidencedto.Evidence{
		Subtype: "code_quality", Source: "lint", CollectedAt: testNow,
		Findings: []evidencedto.Finding{{
			Tool: "lint", RuleID: "err", Severity: "warning",
			FilePath: "main.go", Line: 10, Message: "x", FingerprintID: "fp-100",
		}},
	})

	passVerdict := apiclient.VerdictResult{Outcome: "pass", EvaluatedAt: testNow}
	zeroCounts := apiclient.CountersResult{FindingsCount: 1}

	cfID1, err := cli.OpenCaseFile(ctx, detail.ID, "commit1", "main", false)
	require.NoError(t, err)
	_, err = cli.IngestCaseFileEvidence(ctx, cfID1, evidence)
	require.NoError(t, err)
	res1, err := cli.FinalizeCaseFileWithVerdict(ctx, cfID1, passVerdict, zeroCounts)
	require.NoError(t, err)
	assert.Equal(t, "pass", res1.Verdict.Outcome)

	cfID2, err := cli.OpenCaseFile(ctx, detail.ID, "commit2", "main", false)
	require.NoError(t, err)
	_, err = cli.IngestCaseFileEvidence(ctx, cfID2, evidence)
	require.NoError(t, err)
	res2, err := cli.FinalizeCaseFileWithVerdict(ctx, cfID2, passVerdict, zeroCounts)
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Counters.NewCount)
	assert.Equal(t, 1, res2.Counters.ExistingCount)
}
