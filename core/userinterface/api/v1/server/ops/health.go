package ops

import (
	"context"

	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
)

type Handler struct{}

func New() *Handler { return &Handler{} }

func (*Handler) GetHealth(_ context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: gen.Ok}, nil
}
