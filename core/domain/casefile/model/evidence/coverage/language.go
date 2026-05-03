package coverage

import (
	"fmt"
	"strings"
)

type Language struct {
	name string
}

func NewLanguage(name string) (Language, error) {
	if strings.TrimSpace(name) == "" {
		return Language{}, fmt.Errorf("%w: must not be empty", ErrInvalidLanguage)
	}
	return Language{name: strings.ToLower(name)}, nil
}

func (l Language) String() string {
	return l.name
}

func (l Language) Equal(other Language) bool {
	return l.name == other.name
}
