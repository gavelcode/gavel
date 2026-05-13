package ingestcoverage

import (
	"fmt"
	"strings"
)

type Command struct {
	data   []byte
	format string
	source string
}

func NewCommand(data []byte, format, source string) (Command, error) {
	if len(data) == 0 {
		return Command{}, fmt.Errorf("%w: data must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(format) == "" {
		return Command{}, fmt.Errorf("%w: format must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(source) == "" {
		return Command{}, fmt.Errorf("%w: source must not be empty", ErrInvalidCommand)
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	return Command{
		data:   copied,
		format: format,
		source: source,
	}, nil
}

func (c Command) Data() []byte {
	copied := make([]byte, len(c.data))
	copy(copied, c.data)
	return copied
}
func (c Command) Format() string { return c.format }
func (c Command) Source() string { return c.source }
