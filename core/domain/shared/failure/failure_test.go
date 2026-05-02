package failure_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		kind failure.Kind
	}{
		{name: "validation", msg: "invalid email", kind: failure.Validation},
		{name: "notFound", msg: "project not found", kind: failure.NotFound},
		{name: "conflict", msg: "already exists", kind: failure.Conflict},
		{name: "internal", msg: "unexpected", kind: failure.Internal},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			err := failure.New(tcase.msg, tcase.kind)

			require.Error(t, err)
			assert.Equal(t, tcase.msg, err.Error())
			assert.Equal(t, tcase.kind, failure.Of(err))
		})
	}
}

func TestOfReturnsInternalForNilError(t *testing.T) {
	assert.Equal(t, failure.Internal, failure.Of(nil))
}

func TestOfReturnsInternalForPlainError(t *testing.T) {
	err := errors.New("plain error")

	assert.Equal(t, failure.Internal, failure.Of(err))
}

func TestOfUnwrapsWrappedFailure(t *testing.T) {
	inner := failure.New("not found", failure.NotFound)
	wrapped := fmt.Errorf("load project: %w", inner)

	assert.Equal(t, failure.NotFound, failure.Of(wrapped))
}
