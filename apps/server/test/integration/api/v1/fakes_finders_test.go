package v1integration

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	pleadingmodel "github.com/usegavel/gavel/core/domain/pleading/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

type projectStore struct {
	mu      sync.Mutex
	byKey   map[string]*projectgetbykey.ProjectDetail
	byID    map[string]*projectgetbykey.ProjectDetail
	created map[string]time.Time
	clock   func() time.Time
}

func newProjectStore(clock func() time.Time) *projectStore {
	return &projectStore{
		byKey:   make(map[string]*projectgetbykey.ProjectDetail),
		byID:    make(map[string]*projectgetbykey.ProjectDetail),
		created: make(map[string]time.Time),
		clock:   clock,
	}
}

func (s *projectStore) put(project *projectgetbykey.ProjectDetail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byKey[project.Key] = project
	s.byID[project.ID] = project
	if _, ok := s.created[project.Key]; !ok {
		s.created[project.Key] = s.clock()
	}
}

func (s *projectStore) GetByKey(_ context.Context, key string) (*projectgetbykey.ProjectDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.byKey[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", failure.New("project not found", failure.NotFound), key)
	}
	return p, nil
}

func (s *projectStore) List(_ context.Context, limit, offset int) ([]projectlist.ProjectSummary, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.byKey))
	for k := range s.byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return s.created[keys[i]].After(s.created[keys[j]]) })
	total := len(keys)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	out := make([]projectlist.ProjectSummary, 0, end-offset)
	for _, k := range keys[offset:end] {
		detail := s.byKey[k]
		out = append(out, projectlist.ProjectSummary{
			ID:            detail.ID,
			Key:           detail.Key,
			Name:          detail.Name,
			DefaultBranch: detail.DefaultBranch,
			LatestVerdict: detail.LatestVerdict,
			TotalFindings: detail.TotalFindings,
			CreatedAt:     detail.CreatedAt,
		})
	}
	return out, total, nil
}

type casefileStore struct {
	mu        sync.Mutex
	byID      map[string]*casefileget.CaseFileDetail
	byProject map[string][]string
	findings  []findinglist.FindingView
}

func newCaseFileStore() *casefileStore {
	return &casefileStore{
		byID:      make(map[string]*casefileget.CaseFileDetail),
		byProject: make(map[string][]string),
	}
}

func (s *casefileStore) putDetail(d *casefileget.CaseFileDetail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[d.ID] = d
	s.byProject[d.ProjectID] = append(s.byProject[d.ProjectID], d.ID)
}

func (s *casefileStore) putFinding(f findinglist.FindingView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings = append(s.findings, f)
}

func (s *casefileStore) GetByID(_ context.Context, id string) (*casefileget.CaseFileDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", failure.New("casefile not found", failure.NotFound), id)
	}
	return d, nil
}

func (s *casefileStore) ListByProject(_ context.Context, projectID, _ string, limit, offset int) ([]casefilelist.CaseFileSummary, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.byProject[projectID]
	total := len(ids)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	out := make([]casefilelist.CaseFileSummary, 0, end-offset)
	for _, id := range ids[offset:end] {
		d := s.byID[id]
		out = append(out, casefilelist.CaseFileSummary{
			ID: d.ID, ProjectID: d.ProjectID, CommitSHA: d.CommitSHA, Branch: d.Branch,
			StartedAt: d.StartedAt, VerdictOutcome: d.VerdictOutcome, TotalFindings: d.TotalFindings,
			NewFindings: d.NewFindings, ExistingFindings: d.ExistingFindings, ResolvedFindings: d.ResolvedFindings,
			CoveragePercent: d.CoveragePercent, CreatedAt: d.CreatedAt,
		})
	}
	return out, total, nil
}

func (s *casefileStore) List(_ context.Context, filters findinglist.Filters, limit, offset int) ([]findinglist.FindingView, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	matches := make([]findinglist.FindingView, 0, len(s.findings))
	for _, finding := range s.findings {
		if filters.Severity != "" && finding.Severity != filters.Severity {
			continue
		}
		if filters.Tool != "" && finding.Tool != filters.Tool {
			continue
		}
		matches = append(matches, finding)
	}
	total := len(matches)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matches[offset:end], total, nil
}

type pleadingStore struct {
	mu        sync.Mutex
	byID      map[string]*pleadingget.PleadingDetail
	byProject map[string][]string
}

func newPleadingStore() *pleadingStore {
	return &pleadingStore{
		byID:      make(map[string]*pleadingget.PleadingDetail),
		byProject: make(map[string][]string),
	}
}

func (s *pleadingStore) putDetail(d *pleadingget.PleadingDetail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[d.ID] = d
	s.byProject[d.ProjectID] = append(s.byProject[d.ProjectID], d.ID)
}

func (s *pleadingStore) GetByID(_ context.Context, id string) (*pleadingget.PleadingDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", failure.New("pleading not found", failure.NotFound), id)
	}
	return d, nil
}

func (s *pleadingStore) ListByProject(_ context.Context, projectID, status, _ string, limit, offset int) ([]pleadinglist.PleadingSummary, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.byProject[projectID]
	if projectID == "" {
		ids = nil
		for _, list := range s.byProject {
			ids = append(ids, list...)
		}
	}
	out := make([]pleadinglist.PleadingSummary, 0, len(ids))
	for _, id := range ids {
		detail := s.byID[id]
		if status != "" && detail.Status != status {
			continue
		}
		out = append(out, pleadinglist.PleadingSummary{
			ID: detail.ID, ProjectID: detail.ProjectID, Number: detail.Number, Title: detail.Title, Petitioner: detail.Petitioner,
			SourceBranch: detail.SourceBranch, TargetBranch: detail.TargetBranch, CommitSHA: detail.CommitSHA, Status: detail.Status,
			CreatedAt: detail.CreatedAt, UpdatedAt: detail.UpdatedAt,
		})
	}
	total := len(out)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return out[offset:end], total, nil
}


type searchStore struct {
	results []searchquery.SearchResult
}

func newSearchStore() *searchStore {
	return &searchStore{}
}

func (s *searchStore) Search(_ context.Context, _ string, limit int) ([]searchquery.SearchResult, error) {
	if limit > len(s.results) {
		limit = len(s.results)
	}
	return s.results[:limit], nil
}

type pleadingMemRepo struct {
	mu   sync.Mutex
	byID map[string]pleadingmodel.Pleading
}

func newPleadingMemRepo() *pleadingMemRepo {
	return &pleadingMemRepo{byID: make(map[string]pleadingmodel.Pleading)}
}

func (r *pleadingMemRepo) Save(_ context.Context, p pleadingmodel.Pleading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID().String()] = p
	return nil
}

func (r *pleadingMemRepo) FindByID(_ context.Context, id pleadingmodel.PleadingID) (pleadingmodel.Pleading, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id.String()]
	if !ok {
		return pleadingmodel.Pleading{}, fmt.Errorf("%w: %s", failure.New("pleading not found", failure.NotFound), id.String())
	}
	return p, nil
}

type blobStore struct {
	byKey map[string][]byte
}

func newBlobStore() *blobStore {
	return &blobStore{byKey: make(map[string][]byte)}
}

func (b *blobStore) put(projectID, commit, path string, content []byte) {
	b.byKey[projectID+":"+commit+":"+path] = content
}

func (b *blobStore) Save(_ context.Context, projectID, commit, path string, content []byte, _ string) error {
	b.put(projectID, commit, path, content)
	return nil
}

func (b *blobStore) Fetch(_ context.Context, projectID, commit, path string) ([]byte, string, error) {
	if b == nil {
		return nil, "", fmt.Errorf("%w: blob storage unavailable", failure.New("not configured", failure.NotFound))
	}
	content, ok := b.byKey[projectID+":"+commit+":"+path]
	if !ok {
		return nil, "", fmt.Errorf("%w: %s@%s/%s", failure.New("blob not found", failure.NotFound), projectID, commit, path)
	}
	return content, "", nil
}
