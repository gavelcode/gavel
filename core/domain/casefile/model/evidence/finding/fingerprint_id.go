package finding

import (
	"fmt"
	"strings"
)

type FingerprintID struct {
	value string
}

func NewFingerprintID(value string) (FingerprintID, error) {
	if strings.TrimSpace(value) == "" {
		return FingerprintID{}, fmt.Errorf("%w: must not be empty", ErrInvalidFingerprintID)
	}
	return FingerprintID{value: value}, nil
}

func (f FingerprintID) Value() string {
	return f.value
}

func (f FingerprintID) Equal(other FingerprintID) bool {
	return f.value == other.value
}
