package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
)

type Architecture struct {
	Violations []Violation
}

func fromDomainArchitecture(c architecture.Content) Architecture {
	violations := c.Violations()
	out := make([]Violation, 0, len(violations))
	for _, v := range violations {
		out = append(out, Violation{
			Rule:      v.Rule(),
			SourcePkg: v.SourcePkg(),
			TargetPkg: v.TargetPkg(),
			Message:   v.Message(),
		})
	}
	return Architecture{Violations: out}
}

func toDomainArchitecture(in Architecture) (architecture.Content, error) {
	violations := make([]architecture.Violation, 0, len(in.Violations))
	for i, v := range in.Violations {
		av, err := architecture.NewViolation(v.Rule, v.SourcePkg, v.TargetPkg, v.Message)
		if err != nil {
			return architecture.Content{}, fmt.Errorf("violations[%d]: %w", i, err)
		}
		violations = append(violations, av)
	}
	return architecture.NewContent(violations)
}
