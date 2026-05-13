package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

func TestOfClassifiesDomainErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{name: "notFoundSentinel", err: failure.New("not found", failure.NotFound), want: apperr.NotFound},
		{name: "validationSentinel", err: failure.New("bad input", failure.Validation), want: apperr.Validation},
		{name: "conflictSentinel", err: failure.New("conflict", failure.Conflict), want: apperr.Conflict},
		{name: "wrappedNotFound", err: fmt.Errorf("wrap: %w", failure.New("inner", failure.NotFound)), want: apperr.NotFound},
		{name: "plainError", err: errors.New("plain"), want: apperr.Internal},
		{name: "nilError", err: nil, want: apperr.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, apperr.Of(tt.err))
		})
	}
}

func TestKindConstants(t *testing.T) {
	assert.Equal(t, apperr.Kind(0), apperr.Internal)
	assert.Equal(t, apperr.Kind(1), apperr.Validation)
	assert.Equal(t, apperr.Kind(2), apperr.NotFound)
	assert.Equal(t, apperr.Kind(3), apperr.Conflict)
}
