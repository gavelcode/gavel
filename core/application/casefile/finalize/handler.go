package finalize

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/shared/event"
	casefilemodel "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	caseservice "github.com/usegavel/gavel/core/domain/casefile/service"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type CounterWriter interface {
	WriteCounters(ctx context.Context, caseFileID string, counters Counters) error
}

type Handler struct {
	caseFiles     caseservice.CaseFileRepository
	projects      projectservice.ProjectRepository
	classify      *classify.Handler
	judge         *judge.Handler
	counterWriter CounterWriter
	log           *slog.Logger
}

type HandlerOption func(*Handler)

func WithLogger(log *slog.Logger) HandlerOption {
	return func(h *Handler) { h.log = log }
}

func NewHandler(
	caseFiles caseservice.CaseFileRepository,
	projects projectservice.ProjectRepository,
	classifyH *classify.Handler,
	judgeH *judge.Handler,
	counterWriter CounterWriter,
	opts ...HandlerOption,
) *Handler {
	if caseFiles == nil {
		panic("finalize: caseFiles repository must not be nil")
	}
	if projects == nil {
		panic("finalize: projects repository must not be nil")
	}
	if classifyH == nil {
		panic("finalize: classify handler must not be nil")
	}
	if judgeH == nil {
		panic("finalize: judge handler must not be nil")
	}
	handler := &Handler{
		caseFiles:     caseFiles,
		projects:      projects,
		classify:      classifyH,
		judge:         judgeH,
		counterWriter: counterWriter,
		log:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	cfID, err := casefilemodel.ParseCaseFileID(cmd.CaseFileID())
	if err != nil {
		return Result{}, fmt.Errorf("case file id: %w", err)
	}
	caseFile, err := h.caseFiles.FindByID(ctx, cfID)
	if err != nil {
		return Result{}, fmt.Errorf("load case file: %w", err)
	}

	project, err := h.projects.FindByID(ctx, caseFile.TenantID(), caseFile.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("load project: %w", err)
	}

	findings := extractFindings(caseFile.Evidences())

	fingerprints := cmd.Fingerprints()
	if fingerprints == nil {
		fingerprints = extractFingerprintStrings(findings)
	}
	archIDs := cmd.ArchIDs()
	if archIDs == nil {
		archIDs = extractArchIDStrings(caseFile.Evidences())
	}

	coveragePercent := extractCoveragePercent(caseFile.Evidences())

	var verdictView judge.VerdictView
	var events []event.Event
	var trackingPtr *evidencedto.Tracking

	if precomputed := cmd.PrecomputedVerdict(); precomputed != nil {
		vv, evts, err := h.recordPrecomputedVerdict(ctx, &caseFile, precomputed)
		if err != nil {
			return Result{}, err
		}
		verdictView = vv
		events = evts
	} else {
		if !cmd.Absolute() && !caseFile.IsFreshEvaluation() {
			classifyCmd, err := classify.NewCommand(project.ID().String(), project.DefaultBranch(), findings)
			if err != nil {
				return Result{}, fmt.Errorf("classify command: %w", err)
			}
			classifyRes, err := h.classify.Execute(ctx, classifyCmd)
			if err != nil {
				return Result{}, fmt.Errorf("classify: %w", err)
			}
			t := classifyRes.Tracking
			trackingPtr = &t
		}

		var judgeOpts []judge.CommandOption
		if !cmd.Absolute() {
			delta := computeDelta(project, caseFile.Branch(), fingerprints, cmd.ArchDelta())
			deltaInput := buildDeltaInput(delta, project, caseFile.Branch(), coveragePercent)
			judgeOpts = append(judgeOpts, judge.WithDeltaInput(&deltaInput))
		}

		judgeCmd, err := judge.NewCommand(caseFile.ID().String(), trackingPtr, judgeOpts...)
		if err != nil {
			return Result{}, fmt.Errorf("judge command: %w", err)
		}
		judgeRes, err := h.judge.Execute(ctx, judgeCmd)
		if err != nil {
			return Result{}, fmt.Errorf("judge: %w", err)
		}
		verdictView = judgeRes.Verdict
		events = judgeRes.Events
	}

	var delta Delta
	if !cmd.Absolute() {
		delta = computeDelta(project, caseFile.Branch(), fingerprints, cmd.ArchDelta())
		updateBaseline(ctx, h.projects, project, caseFile.Branch(), verdictView.Outcome, fingerprints, archIDs, cmd.Quick(), coveragePercent, cmd.FileCoverage(), h.log)
	}

	counters := buildCounters(caseFile.Evidences(), fingerprints, trackingPtr, delta)
	h.writeCounters(ctx, caseFile.ID().String(), counters)

	return Result{
		CaseFileID: caseFile.ID().String(),
		Verdict:    verdictView,
		Counters:   counters,
		Delta:      delta,
		Events:     events,
	}, nil
}

func (h *Handler) recordPrecomputedVerdict(ctx context.Context, caseFile *casefilemodel.CaseFile, precomputed *PrecomputedVerdict) (judge.VerdictView, []event.Event, error) {
	var rulings []verdict.Ruling
	for _, r := range precomputed.Rulings {
		subtype, err := evidence.NewSubtype(r.Subtype)
		if err != nil {
			return judge.VerdictView{}, nil, fmt.Errorf("ruling subtype: %w", err)
		}
		rulings = append(rulings, verdict.NewRuling(subtype, r.Passed, r.Detail))
	}

	v, err := verdict.ReconstituteResult(precomputed.Outcome, rulings, precomputed.EvaluatedAt)
	if err != nil {
		return judge.VerdictView{}, nil, fmt.Errorf("reconstitute verdict: %w", err)
	}

	if err := caseFile.RecordVerdict(v); err != nil {
		return judge.VerdictView{}, nil, fmt.Errorf("record verdict: %w", err)
	}

	events := event.EventsFromDomain(caseFile.Events())
	caseFile.ClearEvents()

	if err := h.caseFiles.Save(ctx, *caseFile); err != nil {
		return judge.VerdictView{}, nil, fmt.Errorf("save case file: %w", err)
	}

	rulingViews := make([]judge.RulingView, 0, len(precomputed.Rulings))
	for _, r := range precomputed.Rulings {
		rulingViews = append(rulingViews, judge.RulingView{
			Subtype: r.Subtype,
			Passed:  r.Passed,
			Detail:  r.Detail,
		})
	}

	return judge.VerdictView{
		Outcome:     precomputed.Outcome,
		Rulings:     rulingViews,
		EvaluatedAt: precomputed.EvaluatedAt,
	}, events, nil
}

func (h *Handler) writeCounters(ctx context.Context, caseFileID string, counters Counters) {
	if h.counterWriter == nil {
		return
	}
	if err := h.counterWriter.WriteCounters(ctx, caseFileID, counters); err != nil {
		h.log.Warn("failed to write counters", "caseFileID", caseFileID, "error", err)
	}
}

func buildDeltaInput(delta Delta, project projectmodel.Project, branch string, coveragePercent float64) casefilemodel.DeltaInput {
	deltaInput := casefilemodel.DeltaInput{
		FindingsResolved: delta.FixedCount,
		ArchResolved:     delta.FixedViolationsCount,
		CurrentCoverage:  coveragePercent,
	}
	baseline := project.Baseline(branch)
	if cp := baseline.CoveragePercent(); cp != nil {
		deltaInput.PreviousCoverage = cp
	}
	return deltaInput
}

func computeDelta(project projectmodel.Project, branch string, fingerprints []string, archDelta ArchDeltaInput) Delta {
	baseline := project.Baseline(branch)

	classified := tracking.ClassifyIdentifiers(fingerprints, baseline.Fingerprints())

	return Delta{
		NewCount:        classified.NewCount(),
		FixedCount:      classified.ResolvedCount(),
		ExistingCount:   classified.ExistingCount(),
		NewFingerprints: classified.NewIdentifiers(),
		HasPrevious:     baseline.HasPrevious(),

		PreviousCoveragePercent: baseline.CoveragePercent(),
		PreviousFileCoverage:    toEvidenceFileCoverage(baseline.FileCoverage()),

		NewViolationsCount:      archDelta.NewCount,
		FixedViolationsCount:    archDelta.FixedCount,
		ExistingViolationsCount: archDelta.ExistingCount,
		NewViolationIDs:         archDelta.NewIDs,
		HasArchPrevious:         archDelta.NewCount > 0 || archDelta.FixedCount > 0 || archDelta.ExistingCount > 0,
	}
}

func toEvidenceFileCoverage(entries []projectmodel.FileCoverageEntry) []evidencedto.FileCoverage {
	if len(entries) == 0 {
		return nil
	}
	out := make([]evidencedto.FileCoverage, 0, len(entries))
	for _, e := range entries {
		out = append(out, evidencedto.FileCoverage{
			FilePath:  e.FilePath(),
			Covered:   e.Covered(),
			Uncovered: e.Uncovered(),
		})
	}
	return out
}

func updateBaseline(ctx context.Context, projects projectservice.ProjectRepository, project projectmodel.Project, branch, outcomeStr string, fingerprints, archIDs []string, quick bool, coveragePercent float64, fileCoverage []projectmodel.FileCoverageEntry, log *slog.Logger) {
	outcome, err := verdict.NewOutcome(outcomeStr)
	if err != nil {
		return
	}

	effectiveArchIDs := archIDs
	if quick {
		effectiveArchIDs = project.Baseline(branch).ArchIDs()
	}

	if project.SeedBaselineIfAbsent(branch, fingerprints, effectiveArchIDs, &coveragePercent, fileCoverage) {
		if branch != project.DefaultBranch() && !project.Baseline(project.DefaultBranch()).HasPrevious() {
			project.SeedBaselineIfAbsent(project.DefaultBranch(), fingerprints, effectiveArchIDs, &coveragePercent, fileCoverage)
			log.Info("seeded default branch baseline from feature branch analysis",
				"branch", branch, "defaultBranch", project.DefaultBranch())
		}
		if err := projects.Save(ctx, project); err != nil {
			log.Warn("failed to save baseline", "error", err)
		}
		return
	}

	if outcome.ShouldRecordAsBaseline() {
		project.UpdateBaseline(branch, fingerprints, effectiveArchIDs, &coveragePercent, fileCoverage)
	} else {
		project.RatchetBaseline(branch, fingerprints, effectiveArchIDs)
	}

	if err := projects.Save(ctx, project); err != nil {
		log.Warn("failed to save baseline", "error", err)
	}
}

func extractFindings(evidences []evidence.Evidence) []finding.Finding {
	var all []finding.Finding
	for _, ev := range evidences {
		fc, ok := ev.Content().(finding.Content)
		if !ok {
			continue
		}
		all = append(all, fc.Findings()...)
	}
	return all
}

const percentageMultiplier = 100

func extractCoveragePercent(evidences []evidence.Evidence) float64 {
	for _, ev := range evidences {
		c, ok := ev.Content().(coverage.Content)
		if ok && c.TotalLines() > 0 {
			return float64(c.CoveredLines()) / float64(c.TotalLines()) * percentageMultiplier
		}
	}
	return 0
}

func extractFingerprintStrings(findings []finding.Finding) []string {
	if len(findings) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(findings))
	fps := make([]string, 0, len(findings))
	for _, f := range findings {
		v := f.ID().Value()
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		fps = append(fps, v)
	}
	return fps
}

func extractArchIDStrings(evidences []evidence.Evidence) []string {
	for _, ev := range evidences {
		ac, ok := ev.Content().(architecture.Content)
		if !ok {
			continue
		}
		ids := make([]string, 0, len(ac.Violations()))
		for _, v := range ac.Violations() {
			ids = append(ids, v.Rule()+":"+v.SourcePkg()+":"+v.TargetPkg())
		}
		return ids
	}
	return nil
}

func buildCounters(evidences []evidence.Evidence, fingerprints []string, tracking *evidencedto.Tracking, delta Delta) Counters {
	out := Counters{
		FindingsCount:   len(fingerprints),
		CoveragePercent: extractCoveragePercent(evidences),
	}
	if tracking != nil {
		out.HasTracking = true
		out.NewCount = len(tracking.NewFindings)
		out.ExistingCount = len(tracking.ExistingFindings)
		out.ResolvedCount = tracking.ResolvedCount
	} else if delta.HasPrevious {
		out.HasTracking = true
		out.NewCount = delta.NewCount
		out.ExistingCount = delta.ExistingCount
		out.ResolvedCount = delta.FixedCount
	}
	return out
}
