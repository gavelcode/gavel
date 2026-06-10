package httpx

import (
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func ParseUUIDOrZero(s string) types.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return types.UUID{}
	}
	return types.UUID(id)
}

func Deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
