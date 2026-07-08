package tenant

import (
	"fmt"

	"github.com/google/uuid"
)

type TenantID struct {
	value uuid.UUID
}

var LocalTenantID = TenantID{value: uuid.MustParse("11111111-1111-1111-1111-111111111111")}

func NewTenantID(value uuid.UUID) TenantID {
	return TenantID{value: value}
}

func ParseTenantID(s string) (TenantID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return TenantID{}, fmt.Errorf("%w: %v", ErrInvalidTenant, err)
	}
	if parsedUUID == uuid.Nil {
		return TenantID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidTenant)
	}
	return TenantID{value: parsedUUID}, nil
}

func (id TenantID) IsZero() bool { return id.value == uuid.Nil }

func (id TenantID) UUID() uuid.UUID { return id.value }

func (id TenantID) String() string { return id.value.String() }

func (id TenantID) Equal(other TenantID) bool { return id.value == other.value }
