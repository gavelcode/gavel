package user

import (
	"fmt"
	"strings"
	"unicode"
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	normalised := strings.ToLower(strings.TrimSpace(raw))
	if normalised == "" {
		return Email{}, fmt.Errorf("%w: must not be empty", ErrInvalidEmail)
	}
	if containsWhitespace(normalised) {
		return Email{}, fmt.Errorf("%w: must not contain whitespace", ErrInvalidEmail)
	}
	local, domain, ok := splitOnceAtSign(normalised)
	if !ok {
		return Email{}, fmt.Errorf("%w: must contain exactly one '@'", ErrInvalidEmail)
	}
	if local == "" {
		return Email{}, fmt.Errorf("%w: missing local part", ErrInvalidEmail)
	}
	if domain == "" {
		return Email{}, fmt.Errorf("%w: missing domain", ErrInvalidEmail)
	}
	if !strings.Contains(domain, ".") {
		return Email{}, fmt.Errorf("%w: domain must contain a dot", ErrInvalidEmail)
	}
	return Email{value: normalised}, nil
}

func (e Email) String() string { return e.value }

func (e Email) Equal(other Email) bool { return e.value == other.value }

func containsWhitespace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func splitOnceAtSign(s string) (local, domain string, ok bool) {
	local, domain, ok = strings.Cut(s, "@")
	if !ok {
		return "", "", false
	}
	if strings.ContainsRune(domain, '@') {
		return "", "", false
	}
	return local, domain, true
}
