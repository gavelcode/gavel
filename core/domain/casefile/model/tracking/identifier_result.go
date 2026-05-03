package tracking

type IdentifierResult struct {
	newIdentifiers map[string]bool
	existingCount  int
	resolvedCount  int
}

func ClassifyIdentifiers(current, previous []string) IdentifierResult {
	prevSet := make(map[string]bool, len(previous))
	for _, id := range previous {
		prevSet[id] = true
	}

	newIDs := make(map[string]bool)
	existingCount := 0
	for _, id := range current {
		if prevSet[id] {
			existingCount++
			delete(prevSet, id)
		} else {
			newIDs[id] = true
		}
	}

	return IdentifierResult{
		newIdentifiers: newIDs,
		existingCount:  existingCount,
		resolvedCount:  len(prevSet),
	}
}

func (c IdentifierResult) NewIdentifiers() map[string]bool {
	cp := make(map[string]bool, len(c.newIdentifiers))
	for k, v := range c.newIdentifiers {
		cp[k] = v
	}
	return cp
}

func (c IdentifierResult) NewCount() int      { return len(c.newIdentifiers) }
func (c IdentifierResult) ExistingCount() int { return c.existingCount }
func (c IdentifierResult) ResolvedCount() int { return c.resolvedCount }
