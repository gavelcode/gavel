package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.CaseFileRepository = (*CaseFileRepository)(nil)

var ErrCaseFileNotFound = failure.New("case file not found", failure.NotFound)

type CaseFileRepository struct {
	mu           sync.RWMutex
	byID         map[string]model.CaseFile
	byBranch     map[branchKey][]model.CaseFile
	preloadedFPs map[branchKey][]finding.FingerprintID
}

func NewCaseFileRepository() *CaseFileRepository {
	return &CaseFileRepository{
		byID:         make(map[string]model.CaseFile),
		byBranch:     make(map[branchKey][]model.CaseFile),
		preloadedFPs: make(map[branchKey][]finding.FingerprintID),
	}
}

func (r *CaseFileRepository) Save(_ context.Context, caseFile model.CaseFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := caseFile.ID().String()
	r.byID[id] = caseFile

	key := branchKey{projectID: caseFile.ProjectID().String(), branch: caseFile.Branch()}
	caseFiles := r.byBranch[key]

	for i, cf := range caseFiles {
		if cf.ID().Equal(caseFile.ID()) {
			caseFiles[i] = caseFile
			r.byBranch[key] = caseFiles
			return nil
		}
	}
	r.byBranch[key] = append(caseFiles, caseFile)
	return nil
}

func (r *CaseFileRepository) FindByID(_ context.Context, tenantID tenant.TenantID, caseFileID model.CaseFileID) (model.CaseFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cf, ok := r.byID[caseFileID.String()]
	if !ok || !cf.TenantID().Equal(tenantID) {
		return model.CaseFile{}, fmt.Errorf("%w: %s", ErrCaseFileNotFound, caseFileID)
	}
	return cf, nil
}

func (r *CaseFileRepository) FindByProject(_ context.Context, projectID projectmodel.ProjectID) ([]model.CaseFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.CaseFile
	for _, cf := range r.byID {
		if cf.ProjectID().Equal(projectID) {
			result = append(result, cf)
		}
	}
	return result, nil
}

func (r *CaseFileRepository) FindLatestByBranch(_ context.Context, projectID projectmodel.ProjectID, branch string) (model.CaseFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := branchKey{projectID: projectID.String(), branch: branch}
	caseFiles := r.byBranch[key]
	if len(caseFiles) == 0 {
		return model.CaseFile{}, fmt.Errorf("%w: project %s branch %s", ErrCaseFileNotFound, projectID, branch)
	}

	return latestByStartedAt(caseFiles), nil
}

func (r *CaseFileRepository) PreloadFingerprints(projectID projectmodel.ProjectID, branch string, fingerprints []finding.FingerprintID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := branchKey{projectID: projectID.String(), branch: branch}
	r.preloadedFPs[key] = fingerprints
}

func (r *CaseFileRepository) FindFingerprintIDsByBranch(_ context.Context, projectID projectmodel.ProjectID, branch string) ([]finding.FingerprintID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := branchKey{projectID: projectID.String(), branch: branch}

	caseFiles := r.byBranch[key]
	judged := filterJudged(caseFiles)
	if len(judged) == 0 {
		if preloaded, ok := r.preloadedFPs[key]; ok {
			return preloaded, nil
		}
		return nil, nil
	}

	latest := latestByStartedAt(judged)

	seen := make(map[string]bool)
	var fingerprints []finding.FingerprintID
	for _, ev := range latest.Evidences() {
		fc, ok := ev.Content().(finding.Content)
		if !ok {
			continue
		}
		for _, f := range fc.Findings() {
			fp := f.ID()
			if !seen[fp.Value()] {
				seen[fp.Value()] = true
				fingerprints = append(fingerprints, fp)
			}
		}
	}
	return fingerprints, nil
}

func filterJudged(caseFiles []model.CaseFile) []model.CaseFile {
	var result []model.CaseFile
	for _, cf := range caseFiles {
		if cf.IsJudged() {
			result = append(result, cf)
		}
	}
	return result
}

func latestByStartedAt(caseFiles []model.CaseFile) model.CaseFile {
	latest := caseFiles[0]
	for _, cf := range caseFiles[1:] {
		if cf.StartedAt().After(latest.StartedAt()) {
			latest = cf
		}
	}
	return latest
}
