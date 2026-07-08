package submit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/submit"
)

func TestNewCommandRejectsEmptyCommitSHA(t *testing.T) {
	_, err := submit.NewCommand("tenant-1", "proj-1", "", "main", nil, nil, nil, finalize.ArchDeltaInput{}, nil, false, false, time.Now())
	assert.ErrorIs(t, err, submit.ErrInvalidCommand)
}

func TestNewCommandRejectsEmptyBranch(t *testing.T) {
	_, err := submit.NewCommand("tenant-1", "proj-1", "sha123", "", nil, nil, nil, finalize.ArchDeltaInput{}, nil, false, false, time.Now())
	assert.ErrorIs(t, err, submit.ErrInvalidCommand)
}
