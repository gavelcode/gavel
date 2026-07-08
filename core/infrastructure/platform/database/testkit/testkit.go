package testkit

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	tenantprovision "github.com/usegavel/gavel/core/application/iam/tenant/provision"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/infrastructure/iam/argon2"
	"github.com/usegavel/gavel/core/infrastructure/iam/bootstrap"
	pgiam "github.com/usegavel/gavel/core/infrastructure/iam/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

const SeedAdminPassword = "changeme"

var (
	once          sync.Once
	sharedDB      *database.DB
	sharedDSN     string
	seedAdminHash usermodel.PasswordHash
	initErr       error
	seedTime      = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
)

var errRuntimeUnavailable = errors.New("container runtime unavailable")

func skipOrFail(t *testing.T) {
	t.Helper()
	if errors.Is(initErr, errRuntimeUnavailable) {
		t.Skip("database: " + initErr.Error())
	}
	t.Fatal("database: " + initErr.Error())
}

const (
	pgReadyLogOccurrences = 2
	containerStartTimeout = 30 * time.Second

	testIsolationAdvisoryLock = 8723451
)

type cachedHasher struct{ hash usermodel.PasswordHash }

func (h cachedHasher) Hash(string) (usermodel.PasswordHash, error) { return h.hash, nil }
func (h cachedHasher) Verify(plain string, hash usermodel.PasswordHash) (bool, error) {
	return argon2.New(rand.Reader).Verify(plain, hash)
}

const truncateSQL = `
	TRUNCATE gavelspace_projects, gavelspaces,
	         pleadings,
	         rulings,
	         architecture_violations, tool_execution_failures, new_code_coverage_data,
	         coverage_by_language, coverage_data,
	         findings, evidences, casefile_file_coverage, casefiles,
	         project_baselines, project_quality_gate_rules, project_languages, projects,
	         source_blobs,
	         iam_api_tokens, iam_sessions, iam_users, iam_tenants
	CASCADE
`

func TestDB(t *testing.T) *database.DB {
	t.Helper()
	once.Do(initialize)
	if initErr != nil {
		skipOrFail(t)
	}
	ctx := context.Background()
	conn, err := sharedDB.Conn(ctx)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", testIsolationAdvisoryLock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", testIsolationAdvisoryLock)
		_ = conn.Close()
	})
	_, err = sharedDB.ExecContext(ctx, truncateSQL)
	require.NoError(t, err)

	handler := tenantprovision.NewHandler(pgiam.NewTenantProvisioner(sharedDB), cachedHasher{hash: seedAdminHash})
	cmd, err := tenantprovision.NewCommand(
		bootstrap.DefaultTenantSlug, bootstrap.DefaultTenantDisplayName, bootstrap.DefaultAdminEmail,
		bootstrap.DefaultAdminDisplayName, SeedAdminPassword, seedTime)
	require.NoError(t, err)
	_, err = handler.Execute(ctx, cmd)
	require.NoError(t, err)
	return sharedDB
}

func TestDSN(t *testing.T) string {
	t.Helper()
	once.Do(initialize)
	if initErr != nil {
		skipOrFail(t)
	}
	return sharedDSN
}

func initialize() {
	sharedDB, sharedDSN, initErr = startPostgresContainer(context.Background())
	if initErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: %v\n", initErr)
		return
	}
	seedAdminHash, initErr = argon2.New(rand.Reader).Hash(SeedAdminPassword)
}

func startPostgresContainer(ctx context.Context) (*database.DB, string, error) {
	container, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("gaveltest"),
		tcpg.WithUsername("gavel"),
		tcpg.WithPassword("gavel"),
		testcontainers.WithReuseByName("gavel-database-test-"+database.MigrationsFingerprint()),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(pgReadyLogOccurrences).
				WithStartupTimeout(containerStartTimeout),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", errRuntimeUnavailable, err)
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", fmt.Errorf("connection string: %w", err)
	}
	pgDB, err := database.Open(ctx, dsn)
	if err != nil {
		return nil, "", fmt.Errorf("open database: %w", err)
	}
	if err := database.Migrate(ctx, pgDB); err != nil {
		return nil, "", fmt.Errorf("migrate: %w", err)
	}
	return pgDB, dsn, nil
}
