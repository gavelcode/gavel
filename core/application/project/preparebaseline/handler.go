package preparebaseline

import (
	"context"
	"io"
	"log/slog"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type Handler struct {
	projects ProjectRepository
	seeder   FingerprintSeeder
	fetcher  BaselineFetcher
	log      *slog.Logger
}

func NewHandler(projects ProjectRepository, seeder FingerprintSeeder, opts ...HandlerOption) *Handler {
	if projects == nil {
		panic("preparebaseline: projects repository must not be nil")
	}
	if seeder == nil {
		panic("preparebaseline: fingerprint seeder must not be nil")
	}
	h := &Handler{projects: projects, seeder: seeder, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type HandlerOption func(*Handler)

func WithFetcher(f BaselineFetcher) HandlerOption {
	return func(h *Handler) { h.fetcher = f }
}

func WithBaselineLogger(log *slog.Logger) HandlerOption {
	return func(h *Handler) { h.log = log }
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	var baselines []ProjectBaseline

	for _, input := range cmd.Projects() {
		bl := h.prepareProject(ctx, input)
		baselines = append(baselines, bl)
	}

	return Result{Baselines: baselines}, nil
}

func (h *Handler) prepareProject(ctx context.Context, input ProjectInput) ProjectBaseline {
	if h.fetcher != nil {
		remote, err := h.fetcher.FetchBaseline(ctx, input.Name, input.DefaultBranch)
		if err != nil {
			h.log.Warn("server baseline unavailable", "project", input.Name, "error", err)
		} else if remote != nil && remote.HasPrevious {
			return h.applyRemoteBaseline(ctx, input, remote)
		}
	}

	return h.applyLocalBaseline(ctx, input)
}

func (h *Handler) applyRemoteBaseline(ctx context.Context, input ProjectInput, remote *RemoteBaseline) ProjectBaseline {
	project, err := h.projects.FindByName(ctx, input.Name)
	if err != nil {
		h.log.Warn("project not found for baseline", "project", input.Name, "error", err)
		return ProjectBaseline{ProjectName: input.Name}
	}

	project.UpdateBaseline(input.DefaultBranch, remote.Fingerprints, remote.ArchViolationIDs, nil, nil)
	if err := h.projects.Save(ctx, project); err != nil {
		h.log.Warn("failed to save server baseline", "error", err)
	}

	fps := stringsToFingerprints(remote.Fingerprints)
	if len(fps) > 0 {
		h.seeder.PreloadFingerprints(project.ID(), input.DefaultBranch, fps)
	}

	return ProjectBaseline{
		ProjectName:      input.Name,
		FingerprintCount: len(remote.Fingerprints),
		ArchIDCount:      len(remote.ArchViolationIDs),
		HasPrevious:      true,
		Source:           "server",
	}
}

func (h *Handler) applyLocalBaseline(ctx context.Context, input ProjectInput) ProjectBaseline {
	project, err := h.projects.FindByName(ctx, input.Name)
	if err != nil {
		h.log.Debug("project not found for baseline", "project", input.Name)
		return ProjectBaseline{ProjectName: input.Name}
	}

	baseline := project.Baseline(input.DefaultBranch)
	if !baseline.HasPrevious() {
		h.log.Debug("no baseline found", "project", input.Name)
		return ProjectBaseline{ProjectName: input.Name}
	}

	fps := stringsToFingerprints(baseline.Fingerprints())
	if len(fps) > 0 {
		h.seeder.PreloadFingerprints(project.ID(), input.DefaultBranch, fps)
	}

	return ProjectBaseline{
		ProjectName:      input.Name,
		FingerprintCount: len(baseline.Fingerprints()),
		ArchIDCount:      len(baseline.ArchIDs()),
		HasPrevious:      true,
		Source:           "local",
	}
}

func stringsToFingerprints(raw []string) []finding.FingerprintID {
	fps := make([]finding.FingerprintID, 0, len(raw))
	for _, s := range raw {
		fp, err := finding.NewFingerprintID(s)
		if err != nil {
			continue
		}
		fps = append(fps, fp)
	}
	return fps
}
