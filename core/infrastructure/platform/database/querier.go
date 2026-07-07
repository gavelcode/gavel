package database

import (
	"context"
	"database/sql"
)

// Querier is the query surface shared by *DB and *Tx, so a repository can run
// against the pooled connection or inside a transaction interchangeably — the
// caller decides the boundary by choosing which one it hands the repository.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var (
	_ Querier = (*DB)(nil)
	_ Querier = (*Tx)(nil)
)
