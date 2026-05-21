package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")

const (
	timeFormat         = "2006-01-02 15:04:05"
	rebindBufferGrowth = 16
)

type DB struct {
	*sql.DB
	DriverName string
}

func NewDB(sqlDB *sql.DB, driver string) *DB {
	return &DB{DB: sqlDB, DriverName: driver}
}

func (db *DB) Rebind(query string) string {
	if db.DriverName != "postgres" {
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

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.DB.ExecContext(ctx, db.Rebind(query), args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, db.Rebind(query), args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.DB.QueryRowContext(ctx, db.Rebind(query), args...)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, driverName: db.DriverName}, nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func NullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func ParseTime(dateStr string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		parsed, err = time.Parse(timeFormat, dateStr)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", dateStr, err)
	}
	return parsed.UTC(), nil
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

var likeReplacer = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func EscapeLike(s string) string {
	return likeReplacer.Replace(s)
}
