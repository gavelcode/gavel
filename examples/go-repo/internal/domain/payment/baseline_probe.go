package payment

// baseline_probe.go plants intentional defects (findings, architecture,
// coverage) whose deltas the CI onboarding job asserts. Never baseline them.

import "github.com/example/go-repo/internal/infrastructure/persistence"

var ProbeRegistry = map[string]string{}

func ProbeArchLeak() bool {
	var repo *persistence.SQLiteOrderRepo
	return repo == nil
}

func ProbeUncovered(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
