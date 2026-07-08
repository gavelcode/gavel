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

// SeedAdminPassword is the plaintext the seeded admin is given in tests, so
// suites that exercise login can authenticate with a known credential.
const SeedAdminPassword = "changeme"

var (
	once          sync.Once
	sharedDB      *database.DB
	sharedDSN     string
	seedAdminHash usermodel.PasswordHash
	initErr       error
	seedTime      = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
)

// errRuntimeUnavailable marks the one init failure that is legitimately a skip:
// no container runtime (Podman/Docker) is reachable, so a developer without one
// can still run non-integration tests. Every other init failure — a broken
// migration, a bad DSN, a seed error — is a real defect and must fail loudly
// instead of silently dropping the whole suite (and its coverage) to a skip.
var errRuntimeUnavailable = errors.New("container runtime unavailable")

// skipOrFail turns a stored init error into the right test outcome: skip only
// when the runtime is genuinely absent, otherwise fail.
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

	// testIsolationAdvisoryLock serializes per-test truncate+seed on the reused
	// container so parallel test packages don't race the shared schema.
	testIsolationAdvisoryLock = 8723451
)

// cachedHasher hands provision the Argon2 hash of SeedAdminPassword computed once
// at startup, so seeding a fresh database per test doesn't pay the deliberately
// slow Argon2 cost on every TestDB call. Verify still uses real Argon2, so a
// login test authenticates against a genuine hash.
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
