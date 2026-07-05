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

	currentSet := make(map[string]bool, len(current))
	for _, id := range current {
		currentSet[id] = true
	}

	newIDs := make(map[string]bool)
	existingCount := 0
	for id := range currentSet {
		if prevSet[id] {
			existingCount++
		} else {
			newIDs[id] = true
		}
	}

	resolvedCount := 0
	for id := range prevSet {
		if !currentSet[id] {
			resolvedCount++
		}
	}

	return IdentifierResult{
		newIdentifiers: newIDs,
		existingCount:  existingCount,
		resolvedCount:  resolvedCount,
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
