package v1integration

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

type gavelspaceStore struct {
	mu        sync.Mutex
	byName    map[string]gsmodel.Gavelspace
	createdAt map[string]time.Time
	clock     func() time.Time
}

func newGavelspaceStore(clock func() time.Time) *gavelspaceStore {
	return &gavelspaceStore{
		byName:    make(map[string]gsmodel.Gavelspace),
		createdAt: make(map[string]time.Time),
		clock:     clock,
	}
}

func (s *gavelspaceStore) Save(_ context.Context, gs gsmodel.Gavelspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := gs.ID().String()
	s.byName[name] = gs
	if _, seen := s.createdAt[name]; !seen {
		s.createdAt[name] = s.clock()
	}
	return nil
}

func (s *gavelspaceStore) FindByName(_ context.Context, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	gs, ok := s.byName[name.String()]
	if !ok {
		return gsmodel.Gavelspace{}, fmt.Errorf("%w: %s", failure.New("gavelspace not found", failure.NotFound), name.String())
	}
	return gs, nil
}

func (s *gavelspaceStore) List(_ context.Context, limit, offset int) ([]gslist.GavelspaceSummary, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.byName))
	for n := range s.byName {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		return s.createdAt[names[i]].After(s.createdAt[names[j]])
	})
	total := len(names)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := names[offset:end]

	out := make([]gslist.GavelspaceSummary, 0, len(page))
	for _, n := range page {
		gs := s.byName[n]
		out = append(out, gslist.GavelspaceSummary{
			Name:         n,
			ProjectCount: len(gs.Projects()),
			CreatedAt:    s.createdAt[n],
		})
	}
	return out, total, nil
}

func (s *gavelspaceStore) GetByName(_ context.Context, name string) (*gsget.GavelspaceDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	gs, ok := s.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", failure.New("gavelspace not found", failure.NotFound), name)
	}
	projects := make([]gsget.ProjectRefView, 0, len(gs.Projects()))
	for _, p := range gs.Projects() {
		projects = append(projects, gsget.ProjectRefView{
			ID:            p.ID().String(),
			Key:           p.ID().String(),
			Name:          p.ID().String(),
			LatestVerdict: "",
		})
	}
	return &gsget.GavelspaceDetail{
		Name:      name,
		Projects:  projects,
		CreatedAt: s.createdAt[name],
	}, nil
}
