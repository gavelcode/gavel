package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
)

func NewProblem(status int, title string) gen.Problem {
	return gen.Problem{
		Type:   "about:blank",
		Title:  title,
		Status: int32(status),
	}
}

func BadRequest(detail string) gen.BadRequestJSONResponse {
	return gen.BadRequestJSONResponse(NewProblem(http.StatusBadRequest, detail))
}

func Unauthorized(detail string) gen.UnauthorizedJSONResponse {
	return gen.UnauthorizedJSONResponse(NewProblem(http.StatusUnauthorized, detail))
}

func InvalidCredentials(detail string) gen.InvalidCredentialsJSONResponse {
	return gen.InvalidCredentialsJSONResponse(NewProblem(http.StatusUnauthorized, detail))
}

func CurrentPasswordIncorrect(detail string) gen.CurrentPasswordIncorrectJSONResponse {
	return gen.CurrentPasswordIncorrectJSONResponse(NewProblem(http.StatusUnauthorized, detail))
}

func NotFound(detail string) gen.NotFoundJSONResponse {
	return gen.NotFoundJSONResponse(NewProblem(http.StatusNotFound, detail))
}

func WriteProblem(writer http.ResponseWriter, status int, detail string) {
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(status)
	prob := NewProblem(status, http.StatusText(status))
	if detail != "" {
		prob.Detail = &detail
	}
	_ = json.NewEncoder(writer).Encode(prob)
}
