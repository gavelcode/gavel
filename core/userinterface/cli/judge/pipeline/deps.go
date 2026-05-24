package pipeline

import (
	"log/slog"

	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

type Deps struct {
	Log          *slog.Logger
	Submit       *submit.Handler
	Findings     *ingestfind.Handler
	Coverage     *ingestcov.Handler
	ServerClient *apiclient.Client
}
