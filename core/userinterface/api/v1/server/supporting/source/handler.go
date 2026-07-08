package source

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	"github.com/usegavel/gavel/core/application/shared/apperr"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type Blobs interface {
	Fetch(ctx context.Context, projectID, commitSHA, filePath string) ([]byte, string, error)
	Save(ctx context.Context, projectID, commitSHA, filePath string, content []byte, contentType string) error
}

type FileCoverageFetcher interface {
	Fetch(ctx context.Context, caseFileID, filePath string) (*evidencedto.FileCoverage, error)
}

type FindingsByFileFetcher interface {
	ListByFile(ctx context.Context, tenantID, caseFileID, filePath string) ([]findinglist.FindingView, error)
}

type Deps struct {
	Blobs               Blobs
	ResolveProjectByKey *projectgetbykey.Handler
	FileCoverage        FileCoverageFetcher
	Findings            FindingsByFileFetcher
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) UploadProjectSource(ctx context.Context, req gen.UploadProjectSourceRequestObject) (gen.UploadProjectSourceResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.UploadProjectSource401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	if req.Body == nil {
		return gen.UploadProjectSource400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("request body is required")}, nil
	}
	commit := strings.TrimSpace(req.Body.Commit)
	if commit == "" {
		return gen.UploadProjectSource400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("commit is required")}, nil
	}
	if len(req.Body.Files) == 0 {
		return gen.UploadProjectSource204Response{}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.UploadProjectSource404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	for _, finding := range req.Body.Files {
		path := strings.TrimSpace(finding.Path)
		if path == "" || !isSafeSourcePath(path) {
			continue
		}
		contentType := sourceContentType(path)
		if err := h.deps.Blobs.Save(ctx, detail.ID, commit, path, []byte(finding.Content), contentType); err != nil {
			return nil, err
		}
	}
	return gen.UploadProjectSource204Response{}, nil
}

func (h *Handler) GetProjectSource(ctx context.Context, req gen.GetProjectSourceRequestObject) (gen.GetProjectSourceResponseObject, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return gen.GetProjectSource401JSONResponse{UnauthorizedJSONResponse: httpx.Unauthorized("unauthenticated")}, nil
	}
	commit := req.Params.Commit
	path := req.Params.Path
	if strings.TrimSpace(commit) == "" || strings.TrimSpace(path) == "" {
		return gen.GetProjectSource400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("commit and path are required")}, nil
	}
	if !isSafeSourcePath(path) {
		return gen.GetProjectSource400JSONResponse{BadRequestJSONResponse: httpx.BadRequest("path must be a relative repository path")}, nil
	}
	detail, err := h.fetchProjectDetail(ctx, principal.TenantID, string(req.Key))
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetProjectSource404JSONResponse{NotFoundJSONResponse: httpx.NotFound("project not found")}, nil
		}
		return nil, err
	}
	content, contentType, err := h.deps.Blobs.Fetch(ctx, detail.ID, commit, path)
	if err != nil {
		if apperr.Of(err) == apperr.NotFound {
			return gen.GetProjectSource404JSONResponse{NotFoundJSONResponse: httpx.NotFound("source not found")}, nil
		}
		return nil, err
	}

	if req.Params.Casefile != nil {
		return h.sourceWithContext(ctx, principal.TenantID, content, req.Params.Casefile.String(), path)
	}

	if contentType == "" {
		contentType = sourceContentType(path)
	}
	return gen.GetProjectSource200AsteriskResponse{
		Body:          io.NopCloser(bytes.NewReader(content)),
		ContentType:   contentType,
		ContentLength: int64(len(content)),
	}, nil
}

func (h *Handler) sourceWithContext(ctx context.Context, tenantID string, content []byte, caseFileID, filePath string) (gen.GetProjectSourceResponseObject, error) {
	resp := gen.GetProjectSource200JSONResponse{
		Content: string(content),
	}

	if h.deps.FileCoverage != nil {
		cov, err := h.deps.FileCoverage.Fetch(ctx, caseFileID, filePath)
		if err != nil {
			return nil, err
		}
		if cov != nil {
			resp.Coverage = &gen.FileCoverageData{
				CoveredLines:   toInt32Slice(cov.Covered),
				UncoveredLines: toInt32Slice(cov.Uncovered),
			}
		}
	}

	if h.deps.Findings != nil {
		views, err := h.deps.Findings.ListByFile(ctx, tenantID, caseFileID, filePath)
		if err != nil {
			return nil, err
		}
		if len(views) > 0 {
			findings := make([]gen.Finding, 0, len(views))
			for _, violation := range views {
				findings = append(findings, findingViewToGen(violation))
			}
			resp.Findings = &findings
		}
	}

	return resp, nil
}

func findingViewToGen(violation findinglist.FindingView) gen.Finding {
	return gen.Finding{
		Tool:        violation.Tool,
		RuleId:      violation.RuleID,
		Severity:    violation.Severity,
		FilePath:    violation.FilePath,
		Line:        int32(violation.Line),
		Message:     violation.Message,
		Fingerprint: violation.FingerprintID,
		Status:      violation.Status,
		Source:      violation.Source,
		CommitSha:   violation.CommitSHA,
		ProjectKey:  violation.ProjectKey,
		CasefileId:  httpx.ParseUUIDOrZero(violation.CaseFileID),
	}
}

func toInt32Slice(ints []int) []int32 {
	out := make([]int32, len(ints))
	for i, violation := range ints {
		out[i] = int32(violation)
	}
	return out
}

func (h *Handler) fetchProjectDetail(ctx context.Context, tenantID, key string) (*projectgetbykey.ProjectDetail, error) {
	q, err := projectgetbykey.NewQuery(tenantID, key)
	if err != nil {
		return nil, err
	}
	return h.deps.ResolveProjectByKey.Execute(ctx, q)
}

func isSafeSourcePath(p string) bool {
	if p == "" {
		return false
	}
	cleaned := filepath.ToSlash(filepath.Clean(p))
	if strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return false
	}
	return true
}

func sourceContentType(p string) string {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".json":
		return "application/json; charset=utf-8"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}
