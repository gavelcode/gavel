package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/userinterface/cli/report/checks"
)

const (
	defaultBaseURL    = "https://api.github.com"
	apiVersion        = "2022-11-28"
	statusCompleted   = "completed"
	ownerRepoParts    = 2
	acceptContentType = "application/vnd.github+json"
	jsonContentType   = "application/json"
	requestTimeout    = 30 * time.Second
	commentMarker     = "<!-- gavel-report -->"
)

var (
	errEmptyToken  = errors.New("github: token is required")
	errInvalidRepo = errors.New("github: repo must be in owner/name form")
)

type Config struct {
	Token   string
	Repo    string
	BaseURL string
}

type Publisher struct {
	client  *http.Client
	token   string
	owner   string
	repo    string
	baseURL string
}

type Result struct {
	CheckRunID int64
	URL        string
}

func NewPublisher(config Config) (*Publisher, error) {
	if config.Token == "" {
		return nil, errEmptyToken
	}
	segments := strings.SplitN(config.Repo, "/", ownerRepoParts)
	if len(segments) != ownerRepoParts || segments[0] == "" || segments[1] == "" {
		return nil, fmt.Errorf("%w: %q", errInvalidRepo, config.Repo)
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Publisher{
		client:  &http.Client{Timeout: requestTimeout},
		token:   config.Token,
		owner:   segments[0],
		repo:    segments[1],
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, checkRun checks.CheckRun) (Result, error) {
	batches := checks.BatchAnnotations(checkRun.Annotations, checks.MaxAnnotationsPerRequest)

	var firstBatch []checks.Annotation
	if len(batches) > 0 {
		firstBatch = batches[0]
	}

	created, err := p.create(ctx, checkRun, firstBatch)
	if err != nil {
		return Result{}, err
	}

	for _, batch := range remainingBatches(batches) {
		if err := p.update(ctx, created.CheckRunID, checkRun, batch); err != nil {
			return Result{}, err
		}
	}
	return created, nil
}

func remainingBatches(batches [][]checks.Annotation) [][]checks.Annotation {
	if len(batches) <= 1 {
		return nil
	}
	return batches[1:]
}

func (p *Publisher) create(ctx context.Context, checkRun checks.CheckRun, annotations []checks.Annotation) (Result, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/check-runs", p.baseURL, p.owner, p.repo)
	payload := checkRunPayload{
		Name:       checkRun.Name,
		HeadSHA:    checkRun.HeadSHA,
		Status:     statusCompleted,
		Conclusion: checkRun.Conclusion,
		Output:     buildOutput(checkRun, annotations),
	}
	decoded, err := p.send(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return Result{}, err
	}
	return Result{CheckRunID: decoded.ID, URL: decoded.HTMLURL}, nil
}

func (p *Publisher) update(ctx context.Context, checkRunID int64, checkRun checks.CheckRun, annotations []checks.Annotation) error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/check-runs/%d", p.baseURL, p.owner, p.repo, checkRunID)
	payload := updatePayload{Output: buildOutput(checkRun, annotations)}
	_, err := p.send(ctx, http.MethodPatch, endpoint, payload)
	return err
}

func (p *Publisher) send(ctx context.Context, method, endpoint string, payload any) (checkRunResponse, error) {
	bodyBytes, err := p.request(ctx, method, endpoint, payload)
	if err != nil {
		return checkRunResponse{}, err
	}
	var decoded checkRunResponse
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		return checkRunResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return decoded, nil
}

func (p *Publisher) request(ctx context.Context, method, endpoint string, payload any) ([]byte, error) {
	var reader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+p.token)
	request.Header.Set("Accept", acceptContentType)
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	if payload != nil {
		request.Header.Set("Content-Type", jsonContentType)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call github: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if !isSuccessStatus(response.StatusCode) {
		return nil, fmt.Errorf("github %s %s: status %d: %s",
			method, endpoint, response.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	return bodyBytes, nil
}

func (p *Publisher) UpsertComment(ctx context.Context, prNumber int, body string) error {
	fullBody := commentMarker + "\n" + body
	existing, err := p.findComment(ctx, prNumber)
	if err != nil {
		return err
	}
	if existing > 0 {
		endpoint := fmt.Sprintf("%s/repos/%s/%s/issues/comments/%d", p.baseURL, p.owner, p.repo, existing)
		_, err := p.request(ctx, http.MethodPatch, endpoint, commentPayload{Body: fullBody})
		return err
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", p.baseURL, p.owner, p.repo, prNumber)
	_, err = p.request(ctx, http.MethodPost, endpoint, commentPayload{Body: fullBody})
	return err
}

func (p *Publisher) findComment(ctx context.Context, prNumber int) (int64, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", p.baseURL, p.owner, p.repo, prNumber)
	bodyBytes, err := p.request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	var comments []commentResponse
	if err := json.Unmarshal(bodyBytes, &comments); err != nil {
		return 0, fmt.Errorf("decode comments: %w", err)
	}
	for _, comment := range comments {
		if strings.Contains(comment.Body, commentMarker) {
			return comment.ID, nil
		}
	}
	return 0, nil
}

func isSuccessStatus(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func buildOutput(checkRun checks.CheckRun, annotations []checks.Annotation) outputData {
	converted := make([]annotationData, 0, len(annotations))
	for _, annotation := range annotations {
		converted = append(converted, annotationData{
			Path:            annotation.Path,
			StartLine:       annotation.StartLine,
			EndLine:         annotation.EndLine,
			AnnotationLevel: string(annotation.Level),
			Title:           annotation.Title,
			Message:         annotation.Message,
		})
	}
	return outputData{Title: checkRun.Title, Summary: checkRun.Summary, Annotations: converted}
}

type checkRunPayload struct {
	Name       string     `json:"name"`
	HeadSHA    string     `json:"head_sha"`
	Status     string     `json:"status"`
	Conclusion string     `json:"conclusion"`
	Output     outputData `json:"output"`
}

type updatePayload struct {
	Output outputData `json:"output"`
}

type outputData struct {
	Title       string           `json:"title"`
	Summary     string           `json:"summary"`
	Annotations []annotationData `json:"annotations,omitempty"`
}

type annotationData struct {
	Path            string `json:"path"`
	StartLine       int    `json:"start_line"`
	EndLine         int    `json:"end_line"`
	AnnotationLevel string `json:"annotation_level"`
	Title           string `json:"title,omitempty"`
	Message         string `json:"message"`
}

type checkRunResponse struct {
	ID      int64  `json:"id"`
	HTMLURL string `json:"html_url"`
}

type commentPayload struct {
	Body string `json:"body"`
}

type commentResponse struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}
