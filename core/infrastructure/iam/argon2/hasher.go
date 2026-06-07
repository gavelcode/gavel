package argon2

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	xargon2 "golang.org/x/crypto/argon2"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

var (
	errInvalidHash      = errors.New("argon2: invalid hash format")
	errUnknownAlgorithm = errors.New("argon2: unknown algorithm")
)

type Hasher struct {
	cfg Config
	rng io.Reader
}

var _ service.PasswordHasher = (*Hasher)(nil)

func New(rng io.Reader) *Hasher                   { return &Hasher{cfg: DefaultConfig(), rng: rng} }
func NewWithConfig(rng io.Reader, c Config) *Hasher { return &Hasher{cfg: c, rng: rng} }

func (h *Hasher) Hash(plain string) (user.PasswordHash, error) {
	if h.cfg.SaltLen <= 0 || h.cfg.KeyLen == 0 {
		return user.PasswordHash{}, fmt.Errorf("argon2: invalid config")
	}
	salt := make([]byte, h.cfg.SaltLen)
	if _, err := io.ReadFull(h.rng, salt); err != nil {
		return user.PasswordHash{}, fmt.Errorf("argon2: read salt: %w", err)
	}
	key := xargon2.IDKey([]byte(plain), salt, h.cfg.Time, h.cfg.Memory, h.cfg.Threads, h.cfg.KeyLen)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		xargon2.Version, h.cfg.Memory, h.cfg.Time, h.cfg.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return user.NewPasswordHash(encoded)
}

const argon2HashParts = 6

func (h *Hasher) Verify(plain string, hash user.PasswordHash) (bool, error) {
	parts := strings.Split(hash.String(), "$")
	if len(parts) != argon2HashParts {
		return false, errInvalidHash
	}
	if parts[1] != "argon2id" {
		return false, errUnknownAlgorithm
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, errInvalidHash
	}
	var memory, timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false, errInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errInvalidHash
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errInvalidHash
	}
	gotKey := xargon2.IDKey([]byte(plain), salt, timeCost, memory, threads, uint32(len(expectedKey)))
	return subtle.ConstantTimeCompare(expectedKey, gotKey) == 1, nil
}
