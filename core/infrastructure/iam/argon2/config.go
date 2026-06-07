package argon2

const (
	defaultTime    uint32 = 3
	defaultMemory  uint32 = 64 * 1024
	defaultThreads uint8  = 4
	defaultKeyLen  uint32 = 32
	defaultSaltLen        = 16
)

type Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen int
}

func DefaultConfig() Config {
	return Config{
		Time:    defaultTime,
		Memory:  defaultMemory,
		Threads: defaultThreads,
		KeyLen:  defaultKeyLen,
		SaltLen: defaultSaltLen,
	}
}
