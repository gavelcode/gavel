package v1integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"

	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	gscreate "github.com/usegavel/gavel/core/application/gavelspace/create"
	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	gsregisterproject "github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	gsremoveproject "github.com/usegavel/gavel/core/application/gavelspace/removeproject"
	iamchangepw "github.com/usegavel/gavel/core/application/iam/changepassword"
	iamcreateuser "github.com/usegavel/gavel/core/application/iam/createuser"
	iamissuetoken "github.com/usegavel/gavel/core/application/iam/issuetoken"
	iamlistmytokens "github.com/usegavel/gavel/core/application/iam/listmytokens"
	iamlogin "github.com/usegavel/gavel/core/application/iam/login"
	iamlogout "github.com/usegavel/gavel/core/application/iam/logout"
	iamresolveprincipal "github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	iamrevoketoken "github.com/usegavel/gavel/core/application/iam/revoketoken"
	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	pleadingresolve "github.com/usegavel/gavel/core/application/pleading/resolve"
	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	"github.com/usegavel/gavel/core/application/project/getbaseline"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	updatelanguages "github.com/usegavel/gavel/core/application/project/updatelanguages"
	updatequalitygate "github.com/usegavel/gavel/core/application/project/updatequalitygate"
	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memcasefile "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	casefilev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	gavelspacev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/gavelspace"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
	iamv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
	opsv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
	pleadingv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
	projectv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/project"
	searchv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/search"
	sourcev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/source"

	apiv1 "github.com/usegavel/gavel/apps/server/internal/api/v1"
)

const (
	defaultTenantSlug = "default"
	adminEmail        = "admin@example.com"
	adminPassword     = "hunter22!"
	viewerEmail       = "viewer@example.com"
	viewerPassword    = "viewerpw!"
	sessionCookieName = "gavel_session"
)

var testNow = time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC)

type testFixture struct {
	mux         http.Handler
	tenants     *memiam.TenantRepository
	users       *memiam.UserRepository
	sessions    *memiam.SessionRepository
	tokens      *memiam.APITokenRepository
	hasher      *memiam.FakeHasher
	secrets     *memiam.FakeSecretGenerator
	gavelspaces *gavelspaceStore
	projects    *projectStore
	projRepo    *memproject.ProjectRepository
	cfRepo      *memcasefile.CaseFileRepository
	casefiles   *casefileStore
	pleadings   *pleadingStore
	searches    *searchStore
	blobs       *blobStore
	admin       user.User
	viewer      user.User
	defTenantID string
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()
	ctx := context.Background()

	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	tokens := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secrets := memiam.NewFakeSecretGenerator()

	slug, err := tenant.NewSlug(defaultTenantSlug)
	require.NoError(t, err)
	defaultTenant, err := tenant.NewTenant(slug, "Default", testNow)
	require.NoError(t, err)
	defaultTenant.ClearEvents()
	require.NoError(t, tenants.Save(ctx, defaultTenant))

	admin := seedUser(t, ctx, users, hasher, defaultTenant, adminEmail, "Admin", user.RoleAdmin, adminPassword)
	viewer := seedUser(t, ctx, users, hasher, defaultTenant, viewerEmail, "Viewer", user.RoleViewer, viewerPassword)

	loginH := iamlogin.NewHandler(tenants, users, sessions, hasher, secrets)
	logoutH := iamlogout.NewHandler(sessions)
	changePwH := iamchangepw.NewHandler(users, sessions, hasher)
	issueTokenH := iamissuetoken.NewHandler(users, tokens, secrets)
	revokeTokenH := iamrevoketoken.NewHandler(tokens)
	listTokensH := iamlistmytokens.NewHandler(tokens)
	createUserH := iamcreateuser.NewHandler(tenants, users, hasher)
	resolveH := iamresolveprincipal.NewHandler(users, sessions, tokens)

	clock := func() time.Time { return testNow }
	gsStore := newGavelspaceStore(clock)
	gsCreateH := gscreate.NewHandler(gsStore)
	gsRegisterH := gsregisterproject.NewHandler(gsStore)
	gsRemoveH := gsremoveproject.NewHandler(gsStore)
	gsListH := gslist.NewHandler(gsStore)
	gsGetH := gsget.NewHandler(gsStore)

	projStore := newProjectStore(clock)
	projRepo := memproject.NewProjectRepository()
	projListH := projectlist.NewHandler(projStore)
	projGetH := projectgetbykey.NewHandler(projStore)
	projCreateH := projectcreate.NewHandler(projRepo)
	projUpdQGH := updatequalitygate.NewHandler(projRepo)
	projUpdLangsH := updatelanguages.NewHandler(projRepo)

	cfStore := newCaseFileStore()
	cfRepo := memcasefile.NewCaseFileRepository()
	cfListH := casefilelist.NewHandler(cfStore)
	cfGetH := casefileget.NewHandler(cfStore)
	findListH := findinglist.NewHandler(cfStore)
	cfCreateH := createcasefile.NewHandler(cfRepo, projRepo)
	cfIngestH := ingestevidence.NewHandler(cfRepo)
	judgeH := corejudge.NewHandler(cfRepo, projRepo)
	classifyH := classify.NewHandler(cfRepo)
	cfFinalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)


	plStore := newPleadingStore()
	plListH := pleadinglist.NewHandler(plStore)
	plGetH := pleadingget.NewHandler(plStore)
	plMemRepo := newPleadingMemRepo()
	plFileH := pleadingfile.NewHandler(plMemRepo)
	plResolveH := pleadingresolve.NewHandler(plMemRepo)

	srStore := newSearchStore()
	srH := searchquery.NewHandler(srStore)

	blobs := newBlobStore()

	cookie := auth.SessionCookie{Name: sessionCookieName, Secure: false, TTL: 24 * time.Hour}
	authMw := auth.NewMiddleware(resolveH, cookie, func() time.Time { return testNow })

	server := &apiv1.Server{
		CaseFileHandler: casefilev1.New(casefilev1.Deps{
			ListCaseFiles:       cfListH,
			GetCaseFile:         cfGetH,
			ListFindings:        findListH,
			GetBaseline:         getbaseline.NewHandler(projStore, projRepo),
			CreateCaseFile:      cfCreateH,
			IngestEvidence:      cfIngestH,
			FinalizeCaseFile:    cfFinalizeH,
			ResolveProjectByKey: projGetH,
			Now:                 clock,
		}),
		GavelspaceHandler: gavelspacev1.New(gavelspacev1.Deps{
			ListGavelspaces:           gsListH,
			CreateGavelspace:          gsCreateH,
			GetGavelspace:             gsGetH,
			RegisterGavelspaceProject: gsRegisterH,
			RemoveGavelspaceProject:   gsRemoveH,
		}),
		IAMHandler: iamv1.New(iamv1.Deps{
			Login:          loginH,
			Logout:         logoutH,
			ChangePassword: changePwH,
			IssueToken:     issueTokenH,
			RevokeToken:    revokeTokenH,
			ListMyTokens:   listTokensH,
			CreateUser:     createUserH,
			Cookie:         cookie,
			DefaultTenant:  defaultTenantSlug,
			Now:            clock,
		}),
		OpsHandler: opsv1.New(),
		PleadingHandler: pleadingv1.New(pleadingv1.Deps{
			ListPleadings:       plListH,
			GetPleading:         plGetH,
			FilePleading:        plFileH,
			ResolvePleading:     plResolveH,
			ResolveProjectByKey: projGetH,
		}),
		ProjectHandler: projectv1.New(projectv1.Deps{
			ListProjects:             projListH,
			CreateProject:            projCreateH,
			GetProject:               projGetH,
			UpdateProjectLanguages:   projUpdLangsH,
			UpdateProjectQualityGate: projUpdQGH,
		}),
		SearchHandler: searchv1.New(searchv1.Deps{Search: srH}),
		SourceHandler: sourcev1.New(sourcev1.Deps{
			Blobs:               blobs,
			ResolveProjectByKey: projGetH,
		}),
	}
	mux := apiv1.NewMux(server, authMw)

	return &testFixture{
		mux:         mux,
		tenants:     tenants,
		users:       users,
		sessions:    sessions,
		tokens:      tokens,
		hasher:      hasher,
		secrets:     secrets,
		gavelspaces: gsStore,
		projects:    projStore,
		projRepo:    projRepo,
		cfRepo:      cfRepo,
		casefiles:   cfStore,
		pleadings:   plStore,
		searches:    srStore,
		blobs:       blobs,
		admin:       admin,
		viewer:      viewer,
		defTenantID: defaultTenant.ID().String(),
	}
}

func seedUser(t *testing.T, ctx context.Context, users *memiam.UserRepository, hasher *memiam.FakeHasher, ownerTenant tenant.Tenant, emailStr, displayName string, role user.Role, password string) user.User {
	t.Helper()
	email, err := user.NewEmail(emailStr)
	require.NoError(t, err)
	hash, err := hasher.Hash(password)
	require.NoError(t, err)
	u, err := user.NewUser(ownerTenant.ID(), email, displayName, role, hash, false, testNow)
	require.NoError(t, err)
	u.ClearEvents()
	require.NoError(t, users.Save(ctx, u))
	return u
}

func (f *testFixture) loginCookie(t *testing.T, emailStr, password string) *http.Cookie {
	t.Helper()
	body := map[string]string{"email": emailStr, "password": password}
	res := f.do(t, http.MethodPost, "/sessions", body, nil)
	require.Equal(t, http.StatusOK, res.Code, "login should succeed: %s", res.Body.String())
	for _, c := range res.Result().Cookies() {
		if c.Name == sessionCookieName {
			return c
		}
	}
	t.Fatal("login did not set session cookie")
	return nil
}

func (f *testFixture) do(t *testing.T, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	return rec
}

func mustDecode(t *testing.T, body []byte, dst any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(body, dst))
}
