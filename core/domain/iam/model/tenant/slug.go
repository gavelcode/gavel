package tenant

import (
	"fmt"
	"strings"
)

const tenantSlugMaxLen = 63

type Slug struct {
	value string
}

func NewSlug(raw string) (Slug, error) {
	normalised := strings.ToLower(strings.TrimSpace(raw))
	if normalised == "" {
		return Slug{}, fmt.Errorf("%w: slug must not be empty", ErrInvalidTenant)
	}
	if len(normalised) > tenantSlugMaxLen {
		return Slug{}, fmt.Errorf("%w: slug must be at most %d characters", ErrInvalidTenant, tenantSlugMaxLen)
	}
	if strings.HasPrefix(normalised, "-") || strings.HasSuffix(normalised, "-") {
		return Slug{}, fmt.Errorf("%w: slug must not start or end with a hyphen", ErrInvalidTenant)
	}
	for _, char := range normalised {
		if !isSlugRune(char) {
			return Slug{}, fmt.Errorf("%w: slug may only contain lowercase letters, digits, and hyphens", ErrInvalidTenant)
		}
	}
	return Slug{value: normalised}, nil
}

func (s Slug) String() string { return s.value }

func (s Slug) Equal(other Slug) bool { return s.value == other.value }

func isSlugRune(char rune) bool {
	switch {
	case char >= 'a' && char <= 'z':
		return true
	case char >= '0' && char <= '9':
		return true
	case char == '-':
		return true
	}
	return false
}
