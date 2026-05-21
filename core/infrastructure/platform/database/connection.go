package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed bootstrap.sql
var bootstrapFS embed.FS

//go:embed seed.sql
var seedFS embed.FS

const (
	maxOpenConns    = 25
	maxIdleConns    = 10
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
	pingTimeout     = 5 * time.Second
)

func Open(ctx context.Context, dsn string) (*DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return NewDB(sqlDB, "postgres"), nil
}

func Migrate(database *DB) error {
	fresh, err := isFreshDB(database.DB)
	if err != nil {
		return fmt.Errorf("check fresh db: %w", err)
	}

	if !fresh {
		return nil
	}

	if err := applySchema(database.DB); err != nil {
		return err
	}
	return Seed(database)
}

func Seed(db *DB) error {
	ddl, err := seedFS.ReadFile("seed.sql")
	if err != nil {
		return fmt.Errorf("read seed.sql: %w", err)
	}
	if _, err := db.Exec(string(ddl)); err != nil {
		return fmt.Errorf("apply seed: %w", err)
	}
	return nil
}

func isFreshDB(db *sql.DB) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'iam_tenants'
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func applySchema(database *sql.DB) error {
	ddl, err := bootstrapFS.ReadFile("bootstrap.sql")
	if err != nil {
		return fmt.Errorf("read bootstrap.sql: %w", err)
	}

	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if _, err := transaction.Exec(string(ddl)); err != nil {
		return fmt.Errorf("exec bootstrap: %w", err)
	}

	return transaction.Commit()
}
