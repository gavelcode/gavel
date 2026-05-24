//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bazelrunfiles "github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
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
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	memcasefile "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	casefilev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
	iamv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
	opsv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
	pleadingv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
	projectv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/project"

	apiv1 "github.com/usegavel/gavel/apps/server/internal/api/v1"
)

const (
	e2eAdminEmail    = "admin@e2e.test"
	e2eAdminPassword = "e2e-testpass123!"
	e2eCookieName    = "gavel_session"
	commandTimeout   = 300 * time.Second
)

var e2eTestNow = time.Date(2026, time.June, 16, 12, 0, 0, 0, time.UTC)

func projectRoot(t *testing.T) string {
	t.Helper()

	if dir := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); dir != "" {
		return dir
	}

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "MODULE.bazel")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no MODULE.bazel found walking up from cwd)")
		}
		dir = parent
	}
}

func gavelBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("GAVEL_BINARY"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
		t.Skipf("GAVEL_BINARY set to %s but file not found", bin)
	}

	if bin, ok := bazelrunfiles.FindBinary("apps/cli/cmd/gavel", "gavel"); ok {
		return bin
	}

	root := projectRoot(t)
	cmd := exec.Command("bazel", "cquery", "//apps/cli/cmd/gavel", "--output=files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err == nil {
		rel := strings.TrimSpace(string(out))
		bin := filepath.Join(root, rel)
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}

	t.Skip("gavel binary not found, run: bazel build //apps/cli/cmd/gavel")
	return ""
}

func examplesGoRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(projectRoot(t), "examples", "go-repo")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("examples/go-repo not found at %s", dir)
	}
	return dir
}

type gavelResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runGavel(t *testing.T, args ...string) gavelResult {
	t.Helper()
	return runGavelInDir(t, examplesGoRepo(t), args...)
}

func runGavelInDir(t *testing.T, dir string, args ...string) gavelResult {
	t.Helper()

	bin := gavelBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "BUILD_WORKSPACE_DIRECTORY=") {
			env = append(env, e)
		}
	}
	cmd.Env = append(env, "GAVEL_LOG_LEVEL=error")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run gavel binary: %v", err)
		}
	}

	return gavelResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

func runGavelJSON(t *testing.T, args ...string) map[string]any {
	t.Helper()
	result := runGavel(t, args...)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &parsed),
		"expected valid JSON in stdout, got:\n%s\nstderr:\n%s", result.Stdout, result.Stderr)
	return parsed
}

func cleanWorkspace(t *testing.T) {
	t.Helper()

	root := projectRoot(t)

	checkoutRoot := exec.Command("git", "checkout", "--", ".gavel/", "BUILD.bazel")
	checkoutRoot.Dir = root
	if out, err := checkoutRoot.CombinedOutput(); err != nil {
		t.Logf("warning: git checkout root failed: %v\n%s", err, out)
	}

	workspace := filepath.Join("examples", "go-repo")

	checkout := exec.Command("git", "checkout", "--", workspace)
	checkout.Dir = root
	if out, err := checkout.CombinedOutput(); err != nil {
		t.Logf("warning: git checkout failed: %v\n%s", err, out)
	}

	baselineDir := filepath.Join(workspace, ".gavel", "baseline")
	clean := exec.Command("git", "clean", "-fd", baselineDir)
	clean.Dir = root
	if out, err := clean.CombinedOutput(); err != nil {
		t.Logf("warning: git clean failed: %v\n%s", err, out)
	}
}

type serverFixture struct {
	URL   string
	Token string
	ts    *httptest.Server
}

func startServer(t *testing.T) *serverFixture {
	t.Helper()
	ctx := context.Background()

	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	tokens := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secrets := memiam.NewFakeSecretGenerator()

	slug, err := tenant.NewSlug("default")
	require.NoError(t, err)
	tn, err := tenant.NewTenant(slug, "Default", e2eTestNow)
	require.NoError(t, err)
	tn.ClearEvents()
	require.NoError(t, tenants.Save(ctx, tn))

	email, err := user.NewEmail(e2eAdminEmail)
	require.NoError(t, err)
	hash, err := hasher.Hash(e2eAdminPassword)
	require.NoError(t, err)
	admin, err := user.NewUser(tn.ID(), email, "Admin", user.RoleAdmin, hash, false, e2eTestNow)
	require.NoError(t, err)
	admin.ClearEvents()
	require.NoError(t, users.Save(ctx, admin))

	loginH := iamlogin.NewHandler(tenants, users, sessions, hasher, secrets)
	changePwH := iamchangepw.NewHandler(users, sessions, hasher)
	issueTokenH := iamissuetoken.NewHandler(users, tokens, secrets)
	revokeTokenH := iamrevoketoken.NewHandler(tokens)
	listTokensH := iamlistmytokens.NewHandler(tokens)
	createUserH := iamcreateuser.NewHandler(tenants, users, hasher)
	resolveH := iamresolveprincipal.NewHandler(users, sessions, tokens)

	clock := func() time.Time { return e2eTestNow }

	projRepo := memproject.NewProjectRepository()
	projFinder := &e2eProjectFinder{repo: projRepo}
	projListH := projectlist.NewHandler(projFinder)
	projGetH := projectgetbykey.NewHandler(projFinder)
	projCreateH := projectcreate.NewHandler(projRepo)

	cfRepo := memcasefile.NewCaseFileRepository()
	cfFinder := &e2eCaseFileFinder{repo: cfRepo}
	cfCreateH := createcasefile.NewHandler(cfRepo, projRepo)
	cfIngestH := ingestevidence.NewHandler(cfRepo)
	cfGetH := casefileget.NewHandler(cfFinder)
	cfListH := casefilelist.NewHandler(cfFinder)
	findListH := findinglist.NewHandler(cfFinder)
	judgeH := corejudge.NewHandler(cfRepo, projRepo)
	classifyH := classify.NewHandler(cfRepo)
	cfFinalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)

	plRepo := &e2ePleadingRepo{byID: make(map[string]model.Pleading)}
	plFileH := pleadingfile.NewHandler(plRepo)

	cookie := auth.SessionCookie{Name: e2eCookieName, Secure: false, TTL: 24 * time.Hour}
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

	token := issueE2EIngestToken(t, ts)

	return &serverFixture{
		URL:   ts.URL,
		Token: token,
		ts:    ts,
	}
}

func issueE2EIngestToken(t *testing.T, ts *httptest.Server) string {
	t.Helper()

	loginReq := fmt.Sprintf(`{"email":"%s","password":"%s"}`, e2eAdminEmail, e2eAdminPassword)
	resp, err := ts.Client().Post(ts.URL+"/api/v1/sessions", "application/json", bytes.NewBufferString(loginReq))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "login failed")

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == e2eCookieName {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie, "session cookie not set after login")

	tokenReq, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/api/v1/me/tokens",
		bytes.NewBufferString(`{"name":"e2e-ci","scopes":["ingest","project_sync"]}`),
	)
	require.NoError(t, err)
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.AddCookie(sessionCookie)

	tokenResp, err := ts.Client().Do(tokenReq)
	require.NoError(t, err)
	defer tokenResp.Body.Close()
	require.Equal(t, http.StatusCreated, tokenResp.StatusCode, "token creation failed")

	var tokenResult struct{ Token string }
	require.NoError(t, json.NewDecoder(tokenResp.Body).Decode(&tokenResult))
	return tokenResult.Token
}

type e2eProjectFinder struct {
	repo *memproject.ProjectRepository
}

func (f *e2eProjectFinder) GetByKey(ctx context.Context, key string) (*projectview.ProjectDetail, error) {
	p, err := f.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return &projectview.ProjectDetail{
		ID:            p.ID().String(),
		Key:           p.Key(),
		Name:          p.Name(),
		DefaultBranch: "main",
		TargetPattern: p.TargetPattern(),
		CreatedAt:     e2eTestNow,
	}, nil
}

func (f *e2eProjectFinder) List(_ context.Context, _, _ int) ([]projectlist.ProjectSummary, int, error) {
	return nil, 0, nil
}

type e2eCaseFileFinder struct {
	repo *memcasefile.CaseFileRepository
}

func (f *e2eCaseFileFinder) GetByID(ctx context.Context, id string) (*casefileget.CaseFileDetail, error) {
	cfID, err := casefile.ParseCaseFileID(id)
	if err != nil {
		return nil, err
	}
	cf, err := f.repo.FindByID(ctx, cfID)
	if err != nil {
		return nil, err
	}
	detail := &casefileget.CaseFileDetail{
		ID:        cf.ID().String(),
		ProjectID: cf.ProjectID().String(),
		CommitSHA: cf.CommitSHA(),
		Branch:    cf.Branch(),
		StartedAt: cf.StartedAt(),
		CreatedAt: cf.StartedAt(),
	}
	if v, ok := cf.Verdict(); ok {
		detail.VerdictOutcome = v.Outcome().String()
		for _, r := range v.Rulings() {
			detail.Rulings = append(detail.Rulings, casefileget.RulingView{
				Subtype: r.Subtype().String(),
				Passed:  r.Passed(),
				Detail:  r.Detail(),
			})
		}
	}
	for _, ev := range cf.Evidences() {
		detail.Evidences = append(detail.Evidences, casefileget.EvidenceSummary{
			ID:          ev.ID().String(),
			Subtype:     ev.Subtype().String(),
			Source:      ev.Source(),
			CollectedAt: ev.CollectedAt(),
		})
		if fc, ok := ev.Content().(finding.Content); ok {
			detail.TotalFindings += len(fc.Findings())
		}
		if cc, ok := ev.Content().(coverage.Content); ok && cc.TotalLines() > 0 {
			pct := float64(cc.CoveredLines()) / float64(cc.TotalLines()) * 100
			detail.CoveragePercent = &pct
		}
	}
	return detail, nil
}

func (f *e2eCaseFileFinder) ListByProject(ctx context.Context, projectID, _ string, limit, _ int) ([]casefilelist.CaseFileSummary, int, error) {
	pid, err := projectmodel.ParseProjectID(projectID)
	if err != nil {
		return nil, 0, nil
	}
	cfs, err := f.repo.FindByProject(ctx, pid)
	if err != nil {
		return nil, 0, err
	}
	var summaries []casefilelist.CaseFileSummary
	for _, cf := range cfs {
		s := casefilelist.CaseFileSummary{
			ID:        cf.ID().String(),
			ProjectID: cf.ProjectID().String(),
			CommitSHA: cf.CommitSHA(),
			Branch:    cf.Branch(),
			StartedAt: cf.StartedAt(),
			CreatedAt: cf.StartedAt(),
		}
		if v, ok := cf.Verdict(); ok {
			s.VerdictOutcome = v.Outcome().String()
		}
		for _, ev := range cf.Evidences() {
			if fc, ok := ev.Content().(finding.Content); ok {
				s.TotalFindings += len(fc.Findings())
			}
			if cc, ok := ev.Content().(coverage.Content); ok && cc.TotalLines() > 0 {
				pct := float64(cc.CoveredLines()) / float64(cc.TotalLines()) * 100
				s.CoveragePercent = &pct
			}
		}
		summaries = append(summaries, s)
	}
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return summaries, len(summaries), nil
}

func (f *e2eCaseFileFinder) List(_ context.Context, _ findinglist.Filters, _, _ int) ([]findinglist.FindingView, int, error) {
	return nil, 0, nil
}

type e2ePleadingRepo struct {
	byID map[string]model.Pleading
}

func (r *e2ePleadingRepo) Save(_ context.Context, p model.Pleading) error {
	r.byID[p.ID().String()] = p
	return nil
}

func (r *e2ePleadingRepo) FindByID(_ context.Context, id model.PleadingID) (model.Pleading, error) {
	p, ok := r.byID[id.String()]
	if !ok {
		return model.Pleading{}, fmt.Errorf("%w: %s", failure.New("pleading not found", failure.NotFound), id.String())
	}
	return p, nil
}
