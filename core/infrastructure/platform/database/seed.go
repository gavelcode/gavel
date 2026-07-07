package database

import (
	"context"
	"fmt"
)

const (
	defaultTenantID = "00000000-0000-0000-0000-000000000001"
	defaultAdminID  = "00000000-0000-0000-0000-000000000002"
	seedCreatedAt   = "2026-01-01T00:00:00Z"

	// seedAdvisoryLock serializes first-boot seeding across concurrent server
	// replicas, the same way the migration provider serializes CREATE TABLE.
	seedAdvisoryLock = 8723452
)

// Seed inserts the default tenant and first admin user on a first boot (no
// users yet), reporting whether it seeded. It holds a Postgres transaction-level
// advisory lock, so concurrent replicas booting a fresh database serialize: the
// winner seeds and the others no-op. adminHash is called only when seeding
// actually proceeds, so a one-time password is resolved exactly once by the
// winner and never on a re-run. seeded is true only after the insert commits, so
// a caller can surface a generated credential knowing it was persisted. It stays
// free of IAM and crypto concerns — the caller hashes.
func Seed(ctx context.Context, dbConn *DB, adminHash func() (string, error)) (seeded bool, err error) {
	transaction, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin seed tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if _, err := transaction.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", seedAdvisoryLock); err != nil {
		return false, fmt.Errorf("acquire seed lock: %w", err)
	}

	var userCount int
	if err := transaction.QueryRowContext(ctx, "SELECT count(*) FROM iam_users").Scan(&userCount); err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	if userCount > 0 {
		return false, nil
	}

	adminPasswordHash, err := adminHash()
	if err != nil {
		return false, fmt.Errorf("resolve admin password: %w", err)
	}

	if _, err := transaction.ExecContext(ctx, seedTenantSQL, defaultTenantID, seedCreatedAt); err != nil {
		return false, fmt.Errorf("seed tenant: %w", err)
	}
	if _, err := transaction.ExecContext(ctx, seedAdminSQL, defaultAdminID, defaultTenantID, adminPasswordHash, seedCreatedAt); err != nil {
		return false, fmt.Errorf("seed admin: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return false, fmt.Errorf("commit seed: %w", err)
	}
	return true, nil
}

const seedTenantSQL = `
	INSERT INTO iam_tenants (id, slug, display_name, status, created_at)
	VALUES ($1, 'default', 'Default', 'active', $2)
	ON CONFLICT (id) DO NOTHING`

const seedAdminSQL = `
	INSERT INTO iam_users (
		id, tenant_id, email, display_name, role,
		password_hash, must_change_password, is_active, created_at
	)
	VALUES ($1, $2, 'admin@gavel.local', 'Administrator', 'admin', $3, true, true, $4)
	ON CONFLICT (id) DO NOTHING`
