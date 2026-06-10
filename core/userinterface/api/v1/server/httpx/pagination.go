package httpx

import (
	"encoding/base64"
	"strconv"
)

func PageFromCursor(limit *int, cursor *string) (int, int) {
	lim := 50
	if limit != nil && *limit > 0 {
		lim = *limit
	}
	offset := 0
	if cursor != nil && *cursor != "" {
		if v := decodeOffset(*cursor); v >= 0 {
			offset = v
		}
	}
	return lim, offset
}

func NextCursor(consumed, total int) *string {
	if consumed >= total {
		return nil
	}
	c := encodeOffset(consumed)
	return &c
}

func encodeOffset(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeOffset(cursor string) int {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return -1
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return -1
	}
	return n
}
