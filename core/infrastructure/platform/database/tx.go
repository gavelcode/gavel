package database

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
)

type Tx struct {
	*sql.Tx
	driverName string
}

func (tx *Tx) rebind(query string) string {
	if tx.driverName != "postgres" {
		return query
	}
	var buf strings.Builder
	buf.Grow(len(query) + rebindBufferGrowth)
	idx := 1
	for pos := 0; pos < len(query); pos++ {
		if query[pos] == '?' {
			buf.WriteByte('$')
			buf.WriteString(strconv.Itoa(idx))
			idx++
		} else {
			buf.WriteByte(query[pos])
		}
	}
	return buf.String()
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tx.Tx.ExecContext(ctx, tx.rebind(query), args...)
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return tx.Tx.QueryContext(ctx, tx.rebind(query), args...)
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return tx.Tx.QueryRowContext(ctx, tx.rebind(query), args...)
}
