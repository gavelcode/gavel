package ingestfindings

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Command struct {
	data    []byte
	format  string
	source  string
	subtype evidence.Subtype
}

func NewCommand(data []byte, format, source, subtype string) (Command, error) {
	if len(data) == 0 {
		return Command{}, fmt.Errorf("%w: data must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(format) == "" {
		return Command{}, fmt.Errorf("%w: format must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(source) == "" {
		return Command{}, fmt.Errorf("%w: source must not be empty", ErrInvalidCommand)
	}
	parsedSubtype, err := evidence.NewSubtype(subtype)
	if err != nil {
		return Command{}, fmt.Errorf("%w: %s", ErrInvalidCommand, err.Error())
	}
	if !evidence.IsSubtypeFindingBased(parsedSubtype) {
		return Command{}, fmt.Errorf("%w: subtype %q is not finding-based", ErrInvalidCommand, subtype)
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	return Command{
		data:    copied,
		format:  format,
		source:  source,
		subtype: parsedSubtype,
	}, nil
}

func (c Command) Data() []byte {
	copied := make([]byte, len(c.data))
	copy(copied, c.data)
	return copied
}
func (c Command) Format() string            { return c.format }
func (c Command) Source() string            { return c.source }
func (c Command) Subtype() evidence.Subtype { return c.subtype }
