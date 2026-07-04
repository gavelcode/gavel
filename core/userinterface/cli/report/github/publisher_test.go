package github_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/report/checks"
	"github.com/usegavel/gavel/core/userinterface/cli/report/github"
)

func TestPublishCreatesCheckRun(t *testing.T) {
	var method, calledPath, authorization string
	requestBody := map[string]any{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		method, calledPath, authorization = request.Method, request.URL.Path, request.Header.Get("Authorization")
		rawBody, _ := io.ReadAll(request.Body)
		_ = json.Unmarshal(rawBody, &requestBody)
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"id": 42, "html_url": "https://github.com/octo/repo/runs/42"}`))
	}))
	defer server.Close()

	publisher, err := github.NewPublisher(github.Config{Token: "secret", Repo: "octo/repo", BaseURL: server.URL})
	require.NoError(t, err)

	checkRun := checks.CheckRun{
		Name: "gavel", HeadSHA: "abc123", Conclusion: checks.ConclusionFailure,
		Title: "Gavel", Summary: "the summary",
		Annotations: []checks.Annotation{{Path: "a.go", StartLine: 1, EndLine: 1, Level: checks.LevelFailure, Message: "boom"}},
	}
	result, err := publisher.Publish(context.Background(), checkRun)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/repos/octo/repo/check-runs", calledPath)
	assert.Contains(t, authorization, "secret")
	assert.Equal(t, "abc123", requestBody["head_sha"])
	assert.Equal(t, "failure", requestBody["conclusion"])
	assert.Equal(t, "completed", requestBody["status"])
	assert.EqualValues(t, 42, result.CheckRunID)
	assert.Equal(t, "https://github.com/octo/repo/runs/42", result.URL)
}

type capturedRequest struct {
	Method      string
	Path        string
	Annotations int
}

func TestPublishBatchesAnnotationsAcrossCreateAndPatch(t *testing.T) {
	var captured []capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload struct {
			Output struct {
				Annotations []any `json:"annotations"`
			} `json:"output"`
		}
		rawBody, _ := io.ReadAll(request.Body)
		_ = json.Unmarshal(rawBody, &payload)
		captured = append(captured, capturedRequest{request.Method, request.URL.Path, len(payload.Output.Annotations)})
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"id": 7, "html_url": "https://example.test/7"}`))
	}))
	defer server.Close()

	publisher, err := github.NewPublisher(github.Config{Token: "secret", Repo: "octo/repo", BaseURL: server.URL})
	require.NoError(t, err)

	annotations := make([]checks.Annotation, 51)
	for index := range annotations {
		annotations[index] = checks.Annotation{Path: "a.go", StartLine: 1, EndLine: 1, Level: checks.LevelWarning, Message: "m"}
	}
	_, err = publisher.Publish(context.Background(), checks.CheckRun{
		Name: "gavel", HeadSHA: "head", Conclusion: checks.ConclusionFailure, Annotations: annotations,
	})
	require.NoError(t, err)

	require.Len(t, captured, 2)
	assert.Equal(t, http.MethodPost, captured[0].Method)
	assert.Equal(t, "/repos/octo/repo/check-runs", captured[0].Path)
	assert.Equal(t, 50, captured[0].Annotations)
	assert.Equal(t, http.MethodPatch, captured[1].Method)
	assert.Equal(t, "/repos/octo/repo/check-runs/7", captured[1].Path)
	assert.Equal(t, 1, captured[1].Annotations)
}

func TestPublishReturnsErrorOnFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = writer.Write([]byte(`{"message":"invalid"}`))
	}))
	defer server.Close()

	publisher, err := github.NewPublisher(github.Config{Token: "secret", Repo: "octo/repo", BaseURL: server.URL})
	require.NoError(t, err)
	_, err = publisher.Publish(context.Background(), checks.CheckRun{Name: "gavel", HeadSHA: "abc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "422")
}

func TestPublishReturnsErrorOnMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`not json`))
	}))
	defer server.Close()

	publisher, err := github.NewPublisher(github.Config{Token: "secret", Repo: "octo/repo", BaseURL: server.URL})
	require.NoError(t, err)
	_, err = publisher.Publish(context.Background(), checks.CheckRun{Name: "gavel", HeadSHA: "abc"})
	require.Error(t, err)
}

func TestPublishReturnsErrorWhenServerUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	unreachable := server.URL
	server.Close()

	publisher, err := github.NewPublisher(github.Config{Token: "secret", Repo: "octo/repo", BaseURL: unreachable})
	require.NoError(t, err)
	_, err = publisher.Publish(context.Background(), checks.CheckRun{Name: "gavel", HeadSHA: "abc"})
	require.Error(t, err)
}

func TestNewPublisherRejectsInvalidConfig(t *testing.T) {
	_, err := github.NewPublisher(github.Config{Token: "secret", Repo: "noslash"})
	assert.Error(t, err, "repo without owner/name must be rejected")

	_, err = github.NewPublisher(github.Config{Repo: "octo/repo"})
	assert.Error(t, err, "empty token must be rejected")
}
