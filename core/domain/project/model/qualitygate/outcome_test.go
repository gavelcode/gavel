package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestNewRuleOutcome(t *testing.T) {
	tests := []struct {
		name       string
		passed     bool
		detail     string
		wantPassed bool
		wantDetail string
	}{
		{
			name:       "passWithDetail",
			passed:     true,
			detail:     "all checks passed",
			wantPassed: true,
			wantDetail: "all checks passed",
		},
		{
			name:       "failWithDetail",
			passed:     false,
			detail:     "3 errors found",
			wantPassed: false,
			wantDetail: "3 errors found",
		},
		{
			name:       "passWithEmptyDetail",
			passed:     true,
			detail:     "",
			wantPassed: true,
			wantDetail: "",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			outcome := qualitygate.NewOutcome(tcase.passed, tcase.detail)
			assert.Equal(t, tcase.wantPassed, outcome.Passed())
			assert.Equal(t, tcase.wantDetail, outcome.Detail())
		})
	}
}
