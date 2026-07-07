package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"

	apiv1 "github.com/usegavel/gavel/apps/server/internal/api/v1"
	"github.com/usegavel/gavel/apps/server/internal/platform/config"
	"github.com/usegavel/gavel/apps/server/internal/platform/firstadmin"
	"github.com/usegavel/gavel/apps/server/internal/platform/frontend"
	"github.com/usegavel/gavel/apps/server/internal/platform/spa"
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
	tenantprovision "github.com/usegavel/gavel/core/application/iam/tenant/provision"
	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	pleadingresolve "github.com/usegavel/gavel/core/application/pleading/resolve"
	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	"github.com/usegavel/gavel/core/application/project/getbaseline"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	"github.com/usegavel/gavel/core/application/project/updatelanguages"
	"github.com/usegavel/gavel/core/application/project/updatequalitygate"
	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	gavelspacepostgres "github.com/usegavel/gavel/core/infrastructure/gavelspace/postgres"
	"github.com/usegavel/gavel/core/infrastructure/iam/argon2"
	"github.com/usegavel/gavel/core/infrastructure/iam/crypto"
	pgiam "github.com/usegavel/gavel/core/infrastructure/iam/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/sourceblob"
	pleadingpostgres "github.com/usegavel/gavel/core/infrastructure/pleading/postgres"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
	"github.com/usegavel/gavel/core/infrastructure/supporting/search"
	casefilev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	gavelspacev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/gavelspace"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
	iamv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
	opsv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
	pleadingv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
	projectv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/project"
	searchv1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/search"
	sourcev1 "github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/source"
)

const (
	defaultTenantSlug        = "default"
	defaultTenantDisplayName = "Default"
	defaultAdminEmail        = "admin@gavel.local"

	readHeaderTimeout = 10 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
	cleanupInterval   = 15 * time.Minute
)

func main() {
	root := &cobra.Command{
		Use:   "gavel-server",
		Short: "Gavel Server — API and dashboard backend",
	}
	root.AddCommand(serveCmd())
	root.AddCommand(migrateCmd())
	root.AddCommand(tenantCmd())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			dbConn, err := openAndMigrateDB(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer func() { _ = dbConn.Close() }()

			if err := seedFirstAdmin(cmd.Context(), dbConn, cfg, logger); err != nil {
				return err
			}

			server, authMw, sessions := buildAPIServer(dbConn, cfg, logger)
			router := mountRootRouter(apiv1.NewMux(server, authMw), logger)
			return serveHTTP(cfg, router, sessions, logger)
		},
	}
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			dbConn, err := openAndMigrateDB(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer func() { _ = dbConn.Close() }()
			return nil
		},
	}
}

func openAndMigrateDB(ctx context.Context, cfg *config.Config) (*database.DB, error) {
	dbConn, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := database.Migrate(ctx, dbConn); err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	return dbConn, nil
}

// seedFirstAdmin provisions the default tenant and its admin on a fresh
// database. Only serve does this — migrate stays schema-only, so a migration
// job never logs a credential. It short-circuits when the default tenant is
// already there, so a re-boot neither generates a password nor pays the Argon2
// cost. The password comes from GAVEL_ADMIN_PASSWORD or is generated; provision
// commits the tenant and admin atomically, and the generated password is logged
// only after that commit. Concurrent replicas racing a fresh database serialize
// on the tenant's unique slug: the loser gets ErrSlugTaken and no-ops.
func seedFirstAdmin(ctx context.Context, dbConn *database.DB, cfg *config.Config, logger *slog.Logger) error {
	slug, err := tenant.NewSlug(defaultTenantSlug)
	if err != nil {
		return err
	}
	if _, err := pgiam.NewTenantRepo(dbConn).BySlug(ctx, slug); err == nil {
		return nil
	} else if !errors.Is(err, tenant.ErrTenantNotFound) {
		return fmt.Errorf("check default tenant: %w", err)
	}

	password, generated, err := firstadmin.ResolvePassword(cfg.AdminPassword, rand.Reader)
	if err != nil {
		return err
	}

	handler := tenantprovision.NewHandler(pgiam.NewTenantProvisioner(dbConn), argon2.New(rand.Reader))
	cmd, err := tenantprovision.NewCommand(
		defaultTenantSlug, defaultTenantDisplayName, defaultAdminEmail, defaultAdminDisplayName, password, time.Now().UTC())
	if err != nil {
		return err
	}
	if _, err := handler.Execute(ctx, cmd); err != nil {
		if errors.Is(err, tenant.ErrSlugTaken) {
			return nil
		}
		return err
	}

	if generated {
		logger.Warn("GAVEL_ADMIN_PASSWORD not set; generated a one-time initial admin password, change it after first login",
			"admin_password", password)
	}
	return nil
}

func buildAPIServer(dbConn *database.DB, cfg *config.Config, logger *slog.Logger) (*apiv1.Server, *auth.Middleware, *pgiam.SessionRepo) {
	tenantRepo := pgiam.NewTenantRepo(dbConn)
	userRepo := pgiam.NewUserRepo(dbConn)
	sessionRepo := pgiam.NewSessionRepo(dbConn)
	tokenRepo := pgiam.NewAPITokenRepo(dbConn)
	hasher := argon2.New(rand.Reader)
	secrets := crypto.NewSecretGenerator(rand.Reader)

	loginH := iamlogin.NewHandler(tenantRepo, userRepo, sessionRepo, hasher, secrets)
	logoutH := iamlogout.NewHandler(sessionRepo)
	changePwH := iamchangepw.NewHandler(userRepo, sessionRepo, hasher)
	createUserH := iamcreateuser.NewHandler(tenantRepo, userRepo, hasher)
	issueTokenH := iamissuetoken.NewHandler(userRepo, tokenRepo, secrets)
	revokeTokenH := iamrevoketoken.NewHandler(tokenRepo)
	listTokensH := iamlistmytokens.NewHandler(tokenRepo)
	resolveH := iamresolveprincipal.NewHandler(userRepo, sessionRepo, tokenRepo)

	cookie := auth.SessionCookie{Name: cfg.SessionCookie, Secure: cfg.SecureCookies, TTL: cfg.SessionTTL}
	authMw := auth.NewMiddleware(resolveH, cookie, time.Now)

	coreProjectRepo := projectpostgres.NewRepository(dbConn)
	caseFileRepo := casefilepostgres.NewRepository(dbConn)
	pleadingRepo := pleadingpostgres.NewRepository(dbConn)
	gavelspaceRepo := gavelspacepostgres.NewRepository(dbConn)
	sourceBlobRepo := sourceblob.NewStorage(dbConn)
	fileCoverageStore := casefilepostgres.NewFileCoverageStore(dbConn)

	sqlCaseFileQuery := casefilepostgres.NewCaseFileFinder(dbConn)
	sqlFindingQuery := casefilepostgres.NewFindingFinder(dbConn)
	sqlProjectQuery := projectpostgres.NewProjectFinder(dbConn)
	sqlSearchQuery := search.NewFinder(dbConn)
	sqlPleadingQuery := pleadingpostgres.NewPleadingFinder(dbConn)
	sqlGavelspaceQuery := gavelspacepostgres.NewGavelspaceFinder(dbConn)

	classifyH := classify.NewHandler(caseFileRepo)
	judgeH := corejudge.NewHandler(caseFileRepo, coreProjectRepo)
	createCFH := createcasefile.NewHandler(caseFileRepo, coreProjectRepo)
	ingestEvH := ingestevidence.NewHandler(caseFileRepo)
	finalizeH := finalize.NewHandler(caseFileRepo, coreProjectRepo, classifyH, judgeH, caseFileRepo, finalize.WithLogger(logger))
	createProjectH := projectcreate.NewHandler(coreProjectRepo)
	updateQGH := updatequalitygate.NewHandler(coreProjectRepo)
	updateLangsH := updatelanguages.NewHandler(coreProjectRepo)
	gsCreateH := gscreate.NewHandler(gavelspaceRepo)
	gsRegisterH := gsregisterproject.NewHandler(gavelspaceRepo)
	gsRemoveH := gsremoveproject.NewHandler(gavelspaceRepo)
	pleadingFileH := pleadingfile.NewHandler(pleadingRepo)
	pleadingResolveH := pleadingresolve.NewHandler(pleadingRepo)

	caseFileListH := casefilelist.NewHandler(sqlCaseFileQuery)
	caseFileGetH := casefileget.NewHandler(sqlCaseFileQuery)
	findingListH := findinglist.NewHandler(sqlFindingQuery)
	searchH := searchquery.NewHandler(sqlSearchQuery)
	projectListH := projectlist.NewHandler(sqlProjectQuery)
	projectGetH := projectgetbykey.NewHandler(sqlProjectQuery)
	plListH := pleadinglist.NewHandler(sqlPleadingQuery)
	plGetH := pleadingget.NewHandler(sqlPleadingQuery)
	gsListH := gslist.NewHandler(sqlGavelspaceQuery)
	gsGetH := gsget.NewHandler(sqlGavelspaceQuery)

	server := &apiv1.Server{
		CaseFileHandler: casefilev1.New(casefilev1.Deps{
			ListCaseFiles:       caseFileListH,
			GetCaseFile:         caseFileGetH,
			ListFindings:        findingListH,
			GetBaseline:         getbaseline.NewHandler(sqlProjectQuery, coreProjectRepo),
			CreateCaseFile:      createCFH,
			IngestEvidence:      ingestEvH,
			FinalizeCaseFile:    finalizeH,
			ResolveProjectByKey: projectGetH,
			FileCoverage:        fileCoverageStore,
			Now:                 time.Now,
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
			Now:            time.Now,
		}),
		OpsHandler: opsv1.New(),
		PleadingHandler: pleadingv1.New(pleadingv1.Deps{
			ListPleadings:       plListH,
			GetPleading:         plGetH,
			FilePleading:        pleadingFileH,
			ResolvePleading:     pleadingResolveH,
			ResolveProjectByKey: projectGetH,
		}),
		ProjectHandler: projectv1.New(projectv1.Deps{
			ListProjects:             projectListH,
			CreateProject:            createProjectH,
			GetProject:               projectGetH,
			UpdateProjectLanguages:   updateLangsH,
			UpdateProjectQualityGate: updateQGH,
		}),
		SearchHandler: searchv1.New(searchv1.Deps{Search: searchH}),
		SourceHandler: sourcev1.New(sourcev1.Deps{
			Blobs:               sourceBlobRepo,
			ResolveProjectByKey: projectGetH,
			FileCoverage:        fileCoverageStore,
			Findings:            sqlFindingQuery,
		}),
	}
	return server, authMw, sessionRepo
}

func mountRootRouter(v1Mux http.Handler, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})
	router.Mount("/api/v1", v1Mux)
	router.HandleFunc("/api/*", func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "not found", http.StatusNotFound)
	})
	frontendFS, err := frontend.Load()
	if err != nil {
		logger.Warn("frontend assets not found, serving API only", "err", err)
		return router
	}
	router.Mount("/", spa.Handler(frontendFS))
	return router
}

func serveHTTP(cfg *config.Config, router http.Handler, sessions *pgiam.SessionRepo, logger *slog.Logger) error {
	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go runSessionCleanup(cleanupCtx, sessions, logger)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("gavel listening", "addr", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		logger.Error("server error", "err", err)
		return err
	case <-stop:
		logger.Info("shutting down")
	}

	cleanupCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return httpSrv.Shutdown(shutdownCtx)
}

func runSessionCleanup(ctx context.Context, sessions *pgiam.SessionRepo, logger *slog.Logger) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n, err := sessions.DeleteExpired(ctx, time.Now())
			if err != nil {
				logger.Error("session cleanup failed", "err", err)
			} else if n > 0 {
				logger.Info("expired sessions cleaned", "count", n)
			}
		case <-ctx.Done():
			return
		}
	}
}
