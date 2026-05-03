package verdict_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
)

func TestNewRuling(t *testing.T) {
	tests := []struct {
		name    string
		subtype evidence.Subtype
		passed  bool
		detail  string
	}{
		{
			name:    "passing ruling",
			subtype: evidence.SubtypeCodeQuality,
			passed:  true,
			detail:  "",
		},
		{
			name:    "failing ruling with detail",
			subtype: evidence.SubtypeSAST,
			passed:  false,
			detail:  "3 errors (max 0)",
		},
		{
			name:    "failing coverage ruling",
			subtype: evidence.SubtypeCoverage,
			passed:  false,
			detail:  "50.0% coverage (min 80.0%)",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			ruling := verdict.NewRuling(tcase.subtype, tcase.passed, tcase.detail)

			assert.Equal(t, tcase.subtype, ruling.Subtype())
			assert.Equal(t, tcase.passed, ruling.Passed())
			assert.Equal(t, tcase.detail, ruling.Detail())
		})
	}
}
