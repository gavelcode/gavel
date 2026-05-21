package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plainString", input: "hello", want: "hello"},
		{name: "percentEscaped", input: "test%path", want: `test\%path`},
		{name: "underscoreEscaped", input: "test_path", want: `test\_path`},
		{name: "backslashEscaped", input: `test\path`, want: `test\\path`},
		{name: "allMetacharacters", input: `a%b_c\d`, want: `a\%b\_c\\d`},
		{name: "emptyString", input: "", want: ""},
		{name: "onlyPercent", input: "%", want: `\%`},
		{name: "onlyUnderscore", input: "_", want: `\_`},
		{name: "consecutiveMetachars", input: "%%__", want: `\%\%\_\_`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, database.EscapeLike(tc.input))
		})
	}
}
