package collectevidence_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classifyarch"
	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/ingestncc"
	"github.com/usegavel/gavel/core/infrastructure/casefile/lcov"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
)

type fakeFindingsCollector struct {
	evidences    []evidencedto.Evidence
	rawFiles     []collectevidence.RawFile
	buildWarning string
	err          error
}

func (f *fakeFindingsCollector) CollectFindings(_ context.Context, _ string, _ []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	return f.evidences, f.rawFiles, f.buildWarning, f.err
}

type fakeCoverageCollector struct {
	data []byte
	err  error
}

func (f *fakeCoverageCollector) CollectCoverage(_ context.Context, _ string, _ []string, _ []string) ([]byte, error) {
	return f.data, f.err
}

type fakeArchCollector struct {
	evidence *evidencedto.Evidence
	docs     [][]byte
	err      error
}

func (f *fakeArchCollector) CollectViolations(_ context.Context, _ string, _ []string, _ map[string][]string) (*evidencedto.Evidence, [][]byte, error) {
	return f.evidence, f.docs, f.err
}

type fakeChangedLinesSource struct {
	lines map[string][]int
	err   error
}

func (f *fakeChangedLinesSource) ChangedLines(_ context.Context, _, _ string) (map[string][]int, error) {
	return f.lines, f.err
}

type capturingFindingsCollector struct {
	capturedTargets []string
	evidences       []evidencedto.Evidence
	rawFiles        []collectevidence.RawFile
}

func (c *capturingFindingsCollector) CollectFindings(_ context.Context, _ string, targets []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	c.capturedTargets = targets
	return c.evidences, c.rawFiles, "", nil
}

func newHandler(findings *fakeFindingsCollector, coverage *fakeCoverageCollector, arch *fakeArchCollector) *collectevidence.Handler {
	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser})
	coverageHandler := ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser})
	classifyArchHandler := classifyarch.NewHandler()
	nccHandler := ingestncc.NewHandler(lcovParser)

	return collectevidence.NewHandler(
		findings, coverage, arch,
		findingsHandler, coverageHandler,
		classifyArchHandler, nccHandler,
	)
}

func TestExecute_FindingsOnly_QuickMode(t *testing.T) {
	findings := &fakeFindingsCollector{
		evidences: []evidencedto.Evidence{
			{Subtype: "code_quality", Findings: []evidencedto.Finding{{RuleID: "r1", FingerprintID: "fp1"}}},
		},
		rawFiles: []collectevidence.RawFile{{Format: "sarif", Source: "test.sarif"}},
	}

	handler := newHandler(findings, &fakeCoverageCollector{}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Len(t, result.Fingerprints, 1)
	assert.Equal(t, "fp1", result.Fingerprints[0])
}

func TestExecute_PropagatesBuildWarning(t *testing.T) {
	findings := &fakeFindingsCollector{
		evidences:    []evidencedto.Evidence{{Subtype: "code_quality"}},
		buildWarning: "bazel build had failures",
	}

	handler := newHandler(findings, &fakeCoverageCollector{}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "bazel build had failures", result.BuildWarning)
}

func TestExecute_NoBuildWarning_WhenCleanBuild(t *testing.T) {
	findings := &fakeFindingsCollector{
		evidences: []evidencedto.Evidence{{Subtype: "code_quality"}},
	}

	handler := newHandler(findings, &fakeCoverageCollector{}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Empty(t, result.BuildWarning)
}

func TestExecute_WithCoverage(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nDA:2,1\nDA:3,0\nLF:3\nLH:2\nend_of_record\n")
	findings := &fakeFindingsCollector{}
	coverage := &fakeCoverageCollector{data: lcovData}

	handler := newHandler(findings, coverage, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.InDelta(t, 66.6, result.CovPercent, 1.0)
}

func TestExecute_WithArchitecture(t *testing.T) {
	archEv := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
			},
		},
	}

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, &fakeArchCollector{evidence: archEv, docs: [][]byte{[]byte("{}")}})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ViolationsCount)
	assert.Len(t, result.ArchIDs, 1)
}

func TestExecute_WithArchitectureFiltersToNewViolations(t *testing.T) {
	archEv := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "web", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "ui", TargetPkg: "infra", Message: "forbidden"},
			},
		},
	}

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, &fakeArchCollector{evidence: archEv, docs: [][]byte{[]byte("{}")}})

	baselineArchIDs := []string{"layer:api:domain", "layer:web:domain"}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, baselineArchIDs)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ViolationsCount, "only new violations counted")
	assert.Equal(t, 1, result.ArchDelta.NewCount)
	assert.Equal(t, 2, result.ArchDelta.ExistingCount)
	assert.Equal(t, 0, result.ArchDelta.FixedCount)
}

func TestExecute_ArchIDsCarryAllCurrentViolationsForRatchet(t *testing.T) {
	archEv := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "web", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "ui", TargetPkg: "infra", Message: "forbidden"},
			},
		},
	}

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, &fakeArchCollector{evidence: archEv, docs: [][]byte{[]byte("{}")}})

	baselineArchIDs := []string{"layer:api:domain", "layer:web:domain"}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, baselineArchIDs)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"layer:api:domain", "layer:web:domain", "layer:ui:infra"}, result.ArchIDs)
}

func TestExecute_WithArchitectureNoBaselineReportsAllNew(t *testing.T) {
	archEv := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "web", TargetPkg: "domain", Message: "forbidden"},
			},
		},
	}

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, &fakeArchCollector{evidence: archEv, docs: [][]byte{[]byte("{}")}})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 2, result.ViolationsCount, "all violations are new when no baseline")
	assert.Equal(t, 2, result.ArchDelta.NewCount)
	assert.Equal(t, 0, result.ArchDelta.ExistingCount)
}

func TestExecute_WithScopedTargets_PassesToCollectors(t *testing.T) {
	capturing := &capturingFindingsCollector{}

	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		capturing, &fakeCoverageCollector{}, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
	)

	scoped := []string{"//core/domain:model", "//core/application:app"}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil,
		collectevidence.WithScopedTargets(scoped),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, scoped, capturing.capturedTargets)
}

func TestExecute_WithChangedLinesSource_InvokesNCC(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nDA:2,1\nDA:3,0\nLF:3\nLH:2\nend_of_record\n")
	cls := &fakeChangedLinesSource{
		lines: map[string][]int{"main.go": {1, 2, 3}},
	}

	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		&fakeFindingsCollector{}, &fakeCoverageCollector{data: lcovData}, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
		collectevidence.WithChangedLinesSource(cls),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Greater(t, result.NCCPercent, 0.0)
}

func TestExecute_WithPerLineParser_PopulatesCoverageByFile(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nDA:2,1\nDA:3,0\nLF:3\nLH:2\nend_of_record\n")

	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		&fakeFindingsCollector{}, &fakeCoverageCollector{data: lcovData}, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
		collectevidence.WithPerLineParser(lcovParser),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, result.CoverageByFile, 1)
	assert.Equal(t, "main.go", result.CoverageByFile[0].FilePath)
	assert.Equal(t, []int{1, 2}, result.CoverageByFile[0].Covered)
	assert.Equal(t, []int{3}, result.CoverageByFile[0].Uncovered)
}

func TestExecute_NoPerLineParser_EmptyCoverageByFile(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nDA:2,1\nDA:3,0\nLF:3\nLH:2\nend_of_record\n")

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{data: lcovData}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Empty(t, result.CoverageByFile)
}

func TestExecute_NilChangedLinesSource_SkipsNCC(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nDA:2,1\nDA:3,0\nLF:3\nLH:2\nend_of_record\n")

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{data: lcovData}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 0.0, result.NCCPercent)
}

func TestExecute_FindingsError(t *testing.T) {
	findings := &fakeFindingsCollector{err: errors.New("build failed")}
	handler := newHandler(findings, &fakeCoverageCollector{}, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "findings")
}

func TestExecute_CoverageCollectorError(t *testing.T) {
	coverage := &fakeCoverageCollector{err: errors.New("coverage failed")}
	handler := newHandler(&fakeFindingsCollector{}, coverage, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "coverage")
}

func TestExecute_ArchitectureCollectorError(t *testing.T) {
	arch := &fakeArchCollector{err: errors.New("arch failed")}
	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, arch)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "architecture")
}

func TestExecute_NilFindingsCollector(t *testing.T) {
	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		nil, &fakeCoverageCollector{}, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, true, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, 0, result.FindingsCount)
}

func TestExecute_NilCoverageCollector(t *testing.T) {
	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		&fakeFindingsCollector{}, nil, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.CovPercent)
}

func TestExecute_NilArchCollector(t *testing.T) {
	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	handler := collectevidence.NewHandler(
		&fakeFindingsCollector{}, &fakeCoverageCollector{}, nil,
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ViolationsCount)
}

func TestExecute_CoverageByFileParserError(t *testing.T) {
	lcovData := []byte("TN:\nSF:main.go\nDA:1,1\nLF:1\nLH:1\nend_of_record\n")

	sarifParser := sarif.NewParser()
	lcovParser := lcov.NewParser()
	failParser := &fakePerLineParser{err: errors.New("parse failed")}
	handler := collectevidence.NewHandler(
		&fakeFindingsCollector{}, &fakeCoverageCollector{data: lcovData}, &fakeArchCollector{},
		ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser}),
		ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser}),
		classifyarch.NewHandler(), ingestncc.NewHandler(lcovParser),
		collectevidence.WithPerLineParser(failParser),
	)

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "coverage by file")
}

func TestExecute_AbsoluteArchReturnsAll(t *testing.T) {
	archEv := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "web", TargetPkg: "domain", Message: "forbidden"},
			},
		},
	}

	handler := newHandler(&fakeFindingsCollector{}, &fakeCoverageCollector{}, &fakeArchCollector{evidence: archEv})

	baselineArchIDs := []string{"layer:api:domain"}
	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, true, baselineArchIDs)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, 2, result.ViolationsCount, "absolute mode counts all violations")
}

func TestExecute_CoverageEmptyData(t *testing.T) {
	coverage := &fakeCoverageCollector{data: []byte{}}
	handler := newHandler(&fakeFindingsCollector{}, coverage, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "coverage")
}

func TestExecute_CoverageParseError(t *testing.T) {
	malformedLCOV := []byte("TN:\nSF:main.go\nLF:notanumber\nend_of_record\n")
	coverage := &fakeCoverageCollector{data: malformedLCOV}
	handler := newHandler(&fakeFindingsCollector{}, coverage, &fakeArchCollector{})

	cmd, err := collectevidence.NewCommand("/ws", "//core/...", "core", "main", []string{"go"}, false, false, nil)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "coverage")
}

type fakePerLineParser struct {
	err error
}

func (f *fakePerLineParser) ParsePerLine(_ []byte) (map[string]map[int]int, error) {
	return nil, f.err
}
