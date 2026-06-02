package file

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	pleadingmodel "github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type stubRepo struct{}

func (s *stubRepo) Save(_ context.Context, _ pleadingmodel.Pleading) error { return nil }
func (s *stubRepo) FindByID(_ context.Context, _ pleadingmodel.PleadingID) (pleadingmodel.Pleading, error) {
	return pleadingmodel.Pleading{}, nil
}

func TestExecuteFilePleadingDomainValidationError(t *testing.T) {
	handler := &Handler{pleadings: &stubRepo{}}

	cmd := Command{
		projectID:    projectmodel.NewProjectID(uuid.New()).String(),
		number:       -1,
		title:        "title",
		sourceBranch: "src",
		targetBranch: "dst",
		commitSHA:    "sha",
	}

	_, err := handler.Execute(context.Background(), cmd)
	assert.Error(t, err)
}
