package httpx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func TestParseUUIDOrZeroValid(t *testing.T) {
	result := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", result.String())
}

func TestParseUUIDOrZeroInvalid(t *testing.T) {
	result := httpx.ParseUUIDOrZero("not-a-uuid")

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", result.String())
}

func TestParseUUIDOrZeroEmpty(t *testing.T) {
	result := httpx.ParseUUIDOrZero("")

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", result.String())
}

func TestDerefNil(t *testing.T) {
	assert.Equal(t, "", httpx.Deref(nil))
}

func TestDerefValue(t *testing.T) {
	value := "hello"
	assert.Equal(t, "hello", httpx.Deref(&value))
}
