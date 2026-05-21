package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

func TestNewDB(t *testing.T) {
	db := database.NewDB(nil, "postgres")
	assert.Equal(t, "postgres", db.DriverName)
}

func TestRebindPostgres(t *testing.T) {
	db := database.NewDB(nil, "postgres")
	result := db.Rebind("SELECT * FROM t WHERE a = ? AND b = ?")
	assert.Equal(t, "SELECT * FROM t WHERE a = $1 AND b = $2", result)
}

func TestRebindNonPostgres(t *testing.T) {
	db := database.NewDB(nil, "sqlite")
	result := db.Rebind("SELECT * FROM t WHERE a = ? AND b = ?")
	assert.Equal(t, "SELECT * FROM t WHERE a = ? AND b = ?", result)
}

func TestRebindNoPlaceholders(t *testing.T) {
	db := database.NewDB(nil, "postgres")
	result := db.Rebind("SELECT * FROM t")
	assert.Equal(t, "SELECT * FROM t", result)
}

func TestBoolToIntTrue(t *testing.T) {
	assert.Equal(t, 1, database.BoolToInt(true))
}

func TestBoolToIntFalse(t *testing.T) {
	assert.Equal(t, 0, database.BoolToInt(false))
}

func TestNullableStringEmpty(t *testing.T) {
	assert.Nil(t, database.NullableString(""))
}

func TestNullableStringNonEmpty(t *testing.T) {
	assert.Equal(t, "hello", database.NullableString("hello"))
}

func TestNullStringEmpty(t *testing.T) {
	ns := database.NullString("")
	assert.False(t, ns.Valid)
}

func TestNullStringNonEmpty(t *testing.T) {
	ns := database.NullString("hello")
	assert.True(t, ns.Valid)
	assert.Equal(t, "hello", ns.String)
}

func TestParseTimeRFC3339(t *testing.T) {
	parsed, err := database.ParseTime("2024-06-15T10:30:00Z")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), parsed)
}

func TestParseTimeCustomFormat(t *testing.T) {
	parsed, err := database.ParseTime("2024-06-15 10:30:00")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), parsed)
}

func TestParseTimeInvalid(t *testing.T) {
	_, err := database.ParseTime("not-a-date")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse time")
}

func TestNowReturnsRFC3339(t *testing.T) {
	result := database.Now()
	_, err := time.Parse(time.RFC3339, result)
	require.NoError(t, err)
}

func TestEscapeLikeSpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "shouldEscapePercent", input: "50%", expected: `50\%`},
		{name: "shouldEscapeUnderscore", input: "a_b", expected: `a\_b`},
		{name: "shouldEscapeBackslash", input: `a\b`, expected: `a\\b`},
		{name: "shouldPassThroughPlainText", input: "hello", expected: "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, database.EscapeLike(tt.input))
		})
	}
}
