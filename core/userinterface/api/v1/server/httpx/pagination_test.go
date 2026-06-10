package httpx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func TestPageFromCursorDefaults(t *testing.T) {
	limit, offset := httpx.PageFromCursor(nil, nil)

	assert.Equal(t, 50, limit)
	assert.Equal(t, 0, offset)
}

func TestPageFromCursorCustomLimit(t *testing.T) {
	lim := 20
	limit, offset := httpx.PageFromCursor(&lim, nil)

	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)
}

func TestPageFromCursorZeroLimitFallsToDefault(t *testing.T) {
	lim := 0
	limit, _ := httpx.PageFromCursor(&lim, nil)

	assert.Equal(t, 50, limit)
}

func TestPageFromCursorNegativeLimitFallsToDefault(t *testing.T) {
	lim := -5
	limit, _ := httpx.PageFromCursor(&lim, nil)

	assert.Equal(t, 50, limit)
}

func TestPageFromCursorValidCursor(t *testing.T) {
	lim := 10
	cursor := httpx.NextCursor(0, 100)
	require.NotNil(t, cursor)

	limit, offset := httpx.PageFromCursor(&lim, cursor)

	assert.Equal(t, 10, limit)
	assert.Equal(t, 0, offset)
}

func TestPageFromCursorEmptyCursor(t *testing.T) {
	empty := ""
	_, offset := httpx.PageFromCursor(nil, &empty)

	assert.Equal(t, 0, offset)
}

func TestPageFromCursorInvalidCursor(t *testing.T) {
	bad := "not-valid-base64!!!"
	_, offset := httpx.PageFromCursor(nil, &bad)

	assert.Equal(t, 0, offset)
}

func TestNextCursorReturnsNilWhenExhausted(t *testing.T) {
	cursor := httpx.NextCursor(50, 50)

	assert.Nil(t, cursor)
}

func TestNextCursorReturnsNilWhenBeyondTotal(t *testing.T) {
	cursor := httpx.NextCursor(60, 50)

	assert.Nil(t, cursor)
}

func TestNextCursorReturnsValueWhenMore(t *testing.T) {
	cursor := httpx.NextCursor(20, 50)

	require.NotNil(t, cursor)
	assert.NotEmpty(t, *cursor)
}

func TestCursorRoundTrip(t *testing.T) {
	cursor := httpx.NextCursor(25, 100)
	require.NotNil(t, cursor)

	_, offset := httpx.PageFromCursor(nil, cursor)

	assert.Equal(t, 25, offset)
}
