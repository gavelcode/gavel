package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	maxOpenConns    = 25
	maxIdleConns    = 10
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
	pingTimeout     = 5 * time.Second

	migrationsDir  = "migrations"
	fingerprintLen = 12
)

// MigrationsFingerprint returns a short, deterministic hash of the embedded
// migration set. It changes whenever any migration file is added or edited, so
// a test harness can key an ephemeral database's identity to the schema it
// expects: a schema change yields a new fingerprint (and thus a fresh database)
// instead of reusing one whose accumulated state predates the new migrations.
func MigrationsFingerprint() string {
	entries, err := fs.ReadDir(migrationsFS, migrationsDir)
	if err != nil {
		return "unknown"
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	hasher := sha256.New()
	for _, fileName := range fileNames {
		content, readErr := migrationsFS.ReadFile(migrationsDir + "/" + fileName)
		if readErr != nil {
			return "unknown"
		}
		_, _ = hasher.Write([]byte(fileName))
		_, _ = hasher.Write(content)
	}
	return hex.EncodeToString(hasher.Sum(nil))[:fingerprintLen]
}

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

func Migrate(ctx context.Context, database *DB) error {
	provider, err := migrationProvider(database.DB)
	if err != nil {
		return err
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// migrationProvider builds a goose provider that holds a Postgres session-level
// advisory lock for the duration of the migration, so concurrent server
// replicas booting against a fresh database serialize instead of racing on
// CREATE TABLE. It logs through slog rather than goose's default stdout writer.
func migrationProvider(sqlDB *sql.DB) (*goose.Provider, error) {
	migrations, err := fs.Sub(migrationsFS, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("locate migrations: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("create migration lock: %w", err)
	}
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrations,
		goose.WithSessionLocker(locker),
		goose.WithLogger(gooseLogger{}),
	)
	if err != nil {
		return nil, fmt.Errorf("create migration provider: %w", err)
	}
	return provider, nil
}

type gooseLogger struct{}

func (gooseLogger) Printf(format string, v ...any) {
	slog.Info(strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (gooseLogger) Fatalf(format string, v ...any) {
	slog.Error(strings.TrimSpace(fmt.Sprintf(format, v...)))
}
