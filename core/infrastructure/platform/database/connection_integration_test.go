package database_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

func TestOpenAndMigrateIdempotent(t *testing.T) {
	dsn := testkit.TestDSN(t)

	ctx := context.Background()
	db, err := database.Open(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, database.Migrate(ctx, db))
}

func TestMigrateRecordsSchemaVersion(t *testing.T) {
	db := testkit.TestDB(t)

	version, err := goose.GetDBVersion(db.DB)
	require.NoError(t, err, "Migrate must track applied migrations")
	assert.GreaterOrEqual(t, version, int64(1), "the bootstrap migration must be recorded")
}

func TestBeginTxAndCommit(t *testing.T) {
	db := testkit.TestDB(t)
	ctx := context.Background()

	dbTx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = dbTx.ExecContext(ctx, "SELECT 1")
	require.NoError(t, err)

	require.NoError(t, dbTx.Commit())
}

func TestBeginTxAndRollback(t *testing.T) {
	db := testkit.TestDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())
}

func TestMigrateSkipsAlreadyMigratedDB(t *testing.T) {
	db := testkit.TestDB(t)
	require.NoError(t, database.Migrate(context.Background(), db))
}

func TestMigrateAppliesSchemaToFreshDB(t *testing.T) {
	dsn := testkit.TestDSN(t)
	ctx := context.Background()

	mainDB, err := database.Open(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mainDB.Close() })

	_, err = mainDB.DB.ExecContext(ctx, "CREATE DATABASE test_fresh_migrate")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = mainDB.DB.ExecContext(context.Background(), "DROP DATABASE test_fresh_migrate")
	})

	freshDSN := strings.Replace(dsn, "/gaveltest?", "/test_fresh_migrate?", 1)
	freshDB, err := database.Open(ctx, freshDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = freshDB.Close() })

	require.NoError(t, database.Migrate(ctx, freshDB))

	var exists bool
	err = freshDB.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'iam_tenants')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestDBQueryRowContext(t *testing.T) {
	db := testkit.TestDB(t)
	var result int
	err := db.QueryRowContext(context.Background(), "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestMigrateReturnsErrorOnClosedDB(t *testing.T) {
	dsn := testkit.TestDSN(t)
	ctx := context.Background()
	db, err := database.Open(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	err = database.Migrate(ctx, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "apply migrations")
}

func TestOpenReturnsErrorOnUnreachableHost(t *testing.T) {
	ctx := context.Background()
	_, err := database.Open(ctx, "postgres://localhost:1/nonexistent?sslmode=disable&connect_timeout=1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping postgres")
}

func TestBeginTxReturnsErrorOnCancelledContext(t *testing.T) {
	db := testkit.TestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := db.BeginTx(ctx, nil)
	assert.Error(t, err)
}

func TestMigrateReturnsErrorOnSchemaConflict(t *testing.T) {
	dsn := testkit.TestDSN(t)
	ctx := context.Background()

	mainDB, err := database.Open(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mainDB.Close() })

	_, err = mainDB.DB.ExecContext(ctx, "CREATE DATABASE test_schema_conflict")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = mainDB.DB.ExecContext(context.Background(), "DROP DATABASE test_schema_conflict")
	})

	freshDSN := strings.Replace(dsn, "/gaveltest?", "/test_schema_conflict?", 1)
	freshDB, err := database.Open(ctx, freshDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = freshDB.Close() })

	_, err = freshDB.DB.ExecContext(ctx, "CREATE TABLE projects (id int)")
	require.NoError(t, err)

	err = database.Migrate(ctx, freshDB)
	require.Error(t, err)
}

func TestTxRebindNonPostgres(t *testing.T) {
	pgDB := testkit.TestDB(t)
	nonPgDB := database.NewDB(pgDB.DB, "mysql")
	ctx := context.Background()

	tx, err := nonPgDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	var result int
	err = tx.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestDBQueryContext(t *testing.T) {
	db := testkit.TestDB(t)
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT 1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = rows.Close() })
	require.True(t, rows.Next())
}

func TestTxQueryContext(t *testing.T) {
	db := testkit.TestDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	rows, err := tx.QueryContext(ctx, "SELECT 1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = rows.Close() })
	require.True(t, rows.Next())
}

func TestTxRebindPostgres(t *testing.T) {
	db := testkit.TestDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	var result int
	err = tx.QueryRowContext(ctx, "SELECT ?::int", 42).Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 42, result)
}
