package testkit

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var (
	once      sync.Once
	sharedDB  *database.DB
	sharedDSN string
	initErr   error
)

const (
	pgReadyLogOccurrences = 2
	containerStartTimeout = 30 * time.Second
)

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
		t.Skip("database: " + initErr.Error())
	}
	ctx := context.Background()
	conn, err := sharedDB.Conn(ctx)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock(8723451)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock(8723451)")
		_ = conn.Close()
	})
	_, err = sharedDB.ExecContext(ctx, truncateSQL)
	require.NoError(t, err)
	require.NoError(t, database.Seed(sharedDB))
	return sharedDB
}

func TestDSN(t *testing.T) string {
	t.Helper()
	once.Do(initialize)
	if initErr != nil {
		t.Skip("database: " + initErr.Error())
	}
	return sharedDSN
}

func initialize() {
	sharedDB, sharedDSN, initErr = startPostgresContainer(context.Background())
	if initErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: %v\n", initErr)
	}
}

func startPostgresContainer(ctx context.Context) (*database.DB, string, error) {
	container, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("gaveltest"),
		tcpg.WithUsername("gavel"),
		tcpg.WithPassword("gavel"),
		testcontainers.WithReuseByName("gavel-database-test-v5"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(pgReadyLogOccurrences).
				WithStartupTimeout(containerStartTimeout),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("container runtime unavailable: %w", err)
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
