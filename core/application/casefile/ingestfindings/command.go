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
	// provenance defaults to catalog.
	provenance evidence.Provenance
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

// WithProvenance returns a copy whose findings are attributed to who stands behind the tool.
func (c Command) WithProvenance(provenance evidence.Provenance) Command {
	c.provenance = provenance
	return c
}

// WithProvenanceName is WithProvenance for callers that only hold the name ("catalog" or "bridge"; empty is catalog).
func (c Command) WithProvenanceName(name string) (Command, error) {
	provenance, err := evidence.NewProvenance(name)
	if err != nil {
		return Command{}, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
	}
	return c.WithProvenance(provenance), nil
}

func (c Command) Provenance() evidence.Provenance { return c.provenance }

func (c Command) Data() []byte {
	copied := make([]byte, len(c.data))
	copy(copied, c.data)
	return copied
}
func (c Command) Format() string            { return c.format }
func (c Command) Source() string            { return c.source }
func (c Command) Subtype() evidence.Subtype { return c.subtype }
