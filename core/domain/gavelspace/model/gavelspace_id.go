package model

import (
	"fmt"
	"strings"
)

type GavelspaceID struct {
	value string
}

func NewGavelspaceID(value string) (GavelspaceID, error) {
	if strings.TrimSpace(value) == "" {
		return GavelspaceID{}, fmt.Errorf("%w: gavelspace name must not be empty", ErrInvalidGavelspace)
	}
	return GavelspaceID{value: value}, nil
}

func (n GavelspaceID) String() string {
	return n.value
}

func (n GavelspaceID) Equal(other GavelspaceID) bool {
	return n.value == other.value
}
