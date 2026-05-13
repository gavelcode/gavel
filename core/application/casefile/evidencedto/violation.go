package evidencedto

type Violation struct {
	Rule      string
	SourcePkg string
	TargetPkg string
	Message   string
}

func ExtractViolations(ev *Evidence) []Violation {
	if ev == nil || ev.Architecture == nil {
		return nil
	}
	return ev.Architecture.Violations
}

func ExtractArchIDs(violations []Violation) []string {
	ids := make([]string, 0, len(violations))
	for _, v := range violations {
		ids = append(ids, v.Rule+":"+v.SourcePkg+":"+v.TargetPkg)
	}
	return ids
}
