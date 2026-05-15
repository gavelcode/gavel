package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/project/get"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "validQuery", id: "proj-1", wantErr: false},
		{name: "emptyID", id: "", wantErr: true},
		{name: "whitespaceID", id: "  ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := get.NewQuery(testCase.id)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, get.ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.id, query.ID())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { get.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	now := time.Now()
	expected := &get.ProjectDetail{
		ID:            "proj-1",
		Key:           "my-project",
		Name:          "My Project",
		DefaultBranch: "main",
		LatestVerdict: "pass",
		TotalFindings: 10,
		CreatedAt:     now,
		TargetPattern: "//...",
		Languages:     []string{"java", "go"},
		QualityGateRules: []get.QualityGateRuleView{
			{Subtype: "bug", StrategyType: "absolute"},
		},
		SeverityCounts: map[string]int{"error": 5, "warning": 3},
	}
	fake := &fakeProjectGetter{result: expected}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("proj-1")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Key, result.Key)
	assert.Equal(t, expected.Name, result.Name)
	assert.Equal(t, expected.Languages, result.Languages)
	assert.Equal(t, expected.QualityGateRules, result.QualityGateRules)
	assert.Equal(t, expected.SeverityCounts, result.SeverityCounts)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakeProjectGetter{err: queryErr}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("proj-1")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := get.NewQuery("")
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
