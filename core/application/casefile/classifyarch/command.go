package classifyarch

type Command struct {
	currentIDs  []string
	previousIDs []string
}

func NewCommand(currentIDs, previousIDs []string) Command {
	return Command{
		currentIDs:  copyStrings(currentIDs),
		previousIDs: copyStrings(previousIDs),
	}
}

func (c Command) CurrentIDs() []string  { return copyStrings(c.currentIDs) }
func (c Command) PreviousIDs() []string { return copyStrings(c.previousIDs) }

func copyStrings(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}
