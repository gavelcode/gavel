package ingestncc

import "fmt"

type Command struct {
	rawLCOV      []byte
	changedLines map[string][]int
}

func NewCommand(rawLCOV []byte, changedLines map[string][]int) (Command, error) {
	if len(rawLCOV) == 0 {
		return Command{}, fmt.Errorf("%w: rawLCOV must not be empty", ErrInvalidCommand)
	}
	if len(changedLines) == 0 {
		return Command{}, fmt.Errorf("%w: changedLines must not be empty", ErrInvalidCommand)
	}
	return Command{rawLCOV: rawLCOV, changedLines: changedLines}, nil
}

func (c Command) RawLCOV() []byte           { return c.rawLCOV }
func (c Command) ChangedLines() map[string][]int { return c.changedLines }
