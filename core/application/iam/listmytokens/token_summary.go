package listmytokens

import "time"

type TokenSummary struct {
	ID         string
	Name       string
	Prefix     string
	Scopes     []string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	IsRevoked  bool
	IsExpired  bool
}
