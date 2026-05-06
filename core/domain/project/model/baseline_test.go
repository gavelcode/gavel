package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/project/model"
)

func TestNewBaseline_SortsAndDedups(t *testing.T) {
	baseline := model.NewBaseline(
		[]string{"c", "a", "b", "a"},
		[]string{"z:x:y", "a:b:c", "z:x:y"},
		nil,
		nil,
	)

	assert.Equal(t, []string{"a", "b", "c"}, baseline.Fingerprints())
	assert.Equal(t, []string{"a:b:c", "z:x:y"}, baseline.ArchIDs())
}

func TestNewBaseline_Empty(t *testing.T) {
	b := model.NewBaseline(nil, nil, nil, nil)

	assert.Empty(t, b.Fingerprints())
	assert.Empty(t, b.ArchIDs())
	assert.False(t, b.HasPrevious())
}

func TestBaseline_HasPrevious(t *testing.T) {
	b := model.NewBaseline([]string{"fp1"}, nil, nil, nil)
	assert.True(t, b.HasPrevious())

	b2 := model.NewBaseline(nil, []string{"arch1"}, nil, nil)
	assert.True(t, b2.HasPrevious())
}

func TestBaseline_DefensiveCopy(t *testing.T) {
	fps := []string{"b", "a"}
	baseline := model.NewBaseline(fps, nil, nil, nil)

	fps[0] = "mutated"
	assert.Equal(t, []string{"a", "b"}, baseline.Fingerprints())

	got := baseline.Fingerprints()
	got[0] = "mutated"
	assert.Equal(t, []string{"a", "b"}, baseline.Fingerprints())
}

func TestNewBaseline_WithCoverage(t *testing.T) {
	cov := 85.5
	b := model.NewBaseline([]string{"fp1"}, nil, &cov, nil)

	require.NotNil(t, b.CoveragePercent())
	assert.InDelta(t, 85.5, *b.CoveragePercent(), 0.001)
}

func TestNewBaseline_NilCoverage(t *testing.T) {
	b := model.NewBaseline([]string{"fp1"}, nil, nil, nil)

	assert.Nil(t, b.CoveragePercent())
}

func TestBaseline_CoverageDefensiveCopy(t *testing.T) {
	cov := 85.5
	baseline := model.NewBaseline(nil, nil, &cov, nil)

	cov = 99.0
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 85.5, *baseline.CoveragePercent(), 0.001)

	got := baseline.CoveragePercent()
	*got = 99.0
	assert.InDelta(t, 85.5, *baseline.CoveragePercent(), 0.001)
}

func TestBaseline_HasPreviousWithOnlyCoverage(t *testing.T) {
	cov := 50.0
	b := model.NewBaseline(nil, nil, &cov, nil)

	assert.True(t, b.HasPrevious())
}

func TestRatchet_IntersectsFingerprints(t *testing.T) {
	previous := model.NewBaseline([]string{"fp1", "fp2", "fp3"}, nil, nil, nil)

	result := previous.Ratchet([]string{"fp2", "fp3", "fp4"}, nil)

	assert.Equal(t, []string{"fp2", "fp3"}, result.Fingerprints())
}

func TestRatchet_IntersectsArchIDs(t *testing.T) {
	previous := model.NewBaseline(nil, []string{"a1", "a2", "a3"}, nil, nil)

	result := previous.Ratchet(nil, []string{"a2", "a4"})

	assert.Equal(t, []string{"a2"}, result.ArchIDs())
}

func TestRatchet_PreservesCoverage(t *testing.T) {
	cov := 75.0
	previous := model.NewBaseline([]string{"fp1"}, nil, &cov, nil)

	result := previous.Ratchet([]string{"fp1"}, nil)

	require.NotNil(t, result.CoveragePercent())
	assert.InDelta(t, 75.0, *result.CoveragePercent(), 0.001)
}

func TestRatchet_EmptyPrevious(t *testing.T) {
	previous := model.NewBaseline(nil, nil, nil, nil)

	result := previous.Ratchet([]string{"fp1"}, []string{"a1"})

	assert.Empty(t, result.Fingerprints())
	assert.Empty(t, result.ArchIDs())
}

func TestRatchet_EmptyCurrent(t *testing.T) {
	previous := model.NewBaseline([]string{"fp1"}, []string{"a1"}, nil, nil)

	result := previous.Ratchet(nil, nil)

	assert.Empty(t, result.Fingerprints())
	assert.Empty(t, result.ArchIDs())
}

func TestRatchet_NoOverlap(t *testing.T) {
	previous := model.NewBaseline([]string{"fp1", "fp2"}, []string{"a1"}, nil, nil)

	result := previous.Ratchet([]string{"fp3"}, []string{"a2"})

	assert.Empty(t, result.Fingerprints())
	assert.Empty(t, result.ArchIDs())
}

func TestRatchet_FullOverlap(t *testing.T) {
	previous := model.NewBaseline([]string{"fp1", "fp2"}, []string{"a1"}, nil, nil)

	result := previous.Ratchet([]string{"fp1", "fp2"}, []string{"a1"})

	assert.Equal(t, []string{"fp1", "fp2"}, result.Fingerprints())
	assert.Equal(t, []string{"a1"}, result.ArchIDs())
}

func TestProject_Baseline_DefaultEmpty(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	baseline := project.Baseline("main")
	assert.False(t, baseline.HasPrevious())
	assert.Empty(t, baseline.Fingerprints())
}

func TestProject_UpdateBaseline(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	cov := 85.0
	project.UpdateBaseline("main", []string{"fp2", "fp1"}, []string{"arch1"}, &cov, nil)

	baseline := project.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp1", "fp2"}, baseline.Fingerprints())
	assert.Equal(t, []string{"arch1"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 85.0, *baseline.CoveragePercent(), 0.001)
}

func TestProject_UpdateBaselineNilCoverage(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	project.UpdateBaseline("main", []string{"fp1"}, nil, nil, nil)

	baseline := project.Baseline("main")
	assert.Nil(t, baseline.CoveragePercent())
}

func TestProject_UpdateBaseline_MultipleBranches(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	project.UpdateBaseline("main", []string{"fp1"}, nil, nil, nil)
	project.UpdateBaseline("develop", []string{"fp2"}, nil, nil, nil)

	assert.Equal(t, []string{"fp1"}, project.Baseline("main").Fingerprints())
	assert.Equal(t, []string{"fp2"}, project.Baseline("develop").Fingerprints())
}

func TestProject_UpdateBaseline_Replaces(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	project.UpdateBaseline("main", []string{"old"}, nil, nil, nil)
	project.UpdateBaseline("main", []string{"new"}, nil, nil, nil)

	assert.Equal(t, []string{"new"}, project.Baseline("main").Fingerprints())
}

func TestProject_RatchetBaseline(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	require.NoError(t, err)

	cov := 80.0
	project.UpdateBaseline("main", []string{"fp1", "fp2", "fp3"}, []string{"a1", "a2"}, &cov, nil)

	project.RatchetBaseline("main", []string{"fp2", "fp3", "fp4"}, []string{"a2", "a3"})

	baseline := project.Baseline("main")
	assert.Equal(t, []string{"fp2", "fp3"}, baseline.Fingerprints())
	assert.Equal(t, []string{"a2"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 80.0, *baseline.CoveragePercent(), 0.001)
}

func TestProject_RatchetBaselineNoExistingBaseline(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	require.NoError(t, err)

	project.RatchetBaseline("main", []string{"fp1"}, []string{"a1"})

	baseline := project.Baseline("main")
	assert.False(t, baseline.HasPrevious())
}

func TestProject_SeedBaselineIfAbsent_SeedsWhenNone(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	require.NoError(t, err)

	cov := 57.5
	seeded := project.SeedBaselineIfAbsent("main", []string{"fp1", "fp2"}, []string{"a1"}, &cov, nil)

	assert.True(t, seeded)
	baseline := project.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp1", "fp2"}, baseline.Fingerprints())
	assert.Equal(t, []string{"a1"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 57.5, *baseline.CoveragePercent(), 0.001)
}

func TestProject_SeedBaselineIfAbsent_NoOpWhenExists(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	require.NoError(t, err)

	cov := 80.0
	project.UpdateBaseline("main", []string{"existing"}, []string{"a-existing"}, &cov, nil)

	newCov := 90.0
	seeded := project.SeedBaselineIfAbsent("main", []string{"new"}, []string{"a-new"}, &newCov, nil)

	assert.False(t, seeded)
	baseline := project.Baseline("main")
	assert.Equal(t, []string{"existing"}, baseline.Fingerprints())
	assert.Equal(t, []string{"a-existing"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 80.0, *baseline.CoveragePercent(), 0.001)
}

func TestNewFileCoverageEntry(t *testing.T) {
	entry, err := model.NewFileCoverageEntry("src/main.go", []int{1, 2, 3}, []int{4, 5})
	require.NoError(t, err)

	assert.Equal(t, "src/main.go", entry.FilePath())
	assert.Equal(t, []int{1, 2, 3}, entry.Covered())
	assert.Equal(t, []int{4, 5}, entry.Uncovered())
}

func TestNewFileCoverageEntryRejectsEmptyPath(t *testing.T) {
	_, err := model.NewFileCoverageEntry("", []int{1}, []int{2})
	assert.Error(t, err)
}

func TestFileCoverageEntryDefensiveCopy(t *testing.T) {
	covered := []int{1, 2, 3}
	uncovered := []int{4, 5}
	entry, err := model.NewFileCoverageEntry("f.go", covered, uncovered)
	require.NoError(t, err)

	covered[0] = 99
	uncovered[0] = 99
	assert.Equal(t, []int{1, 2, 3}, entry.Covered())
	assert.Equal(t, []int{4, 5}, entry.Uncovered())

	got := entry.Covered()
	got[0] = 99
	assert.Equal(t, []int{1, 2, 3}, entry.Covered())
}

func TestNewBaseline_WithFileCoverage(t *testing.T) {
	entry, err := model.NewFileCoverageEntry("f.go", []int{1, 2}, []int{3})
	require.NoError(t, err)
	cov := 66.7

	b := model.NewBaseline(nil, nil, &cov, []model.FileCoverageEntry{entry})

	require.Len(t, b.FileCoverage(), 1)
	assert.Equal(t, "f.go", b.FileCoverage()[0].FilePath())
}

func TestBaseline_FileCoverageDefensiveCopy(t *testing.T) {
	entry, err := model.NewFileCoverageEntry("f.go", []int{1}, nil)
	require.NoError(t, err)

	b := model.NewBaseline(nil, nil, nil, []model.FileCoverageEntry{entry})

	got := b.FileCoverage()
	got[0] = model.FileCoverageEntry{}
	assert.Equal(t, "f.go", b.FileCoverage()[0].FilePath())
}

func TestNewBaseline_NilFileCoverage(t *testing.T) {
	b := model.NewBaseline(nil, nil, nil, nil)
	assert.Nil(t, b.FileCoverage())
}

func TestProject_Baselines(t *testing.T) {
	project, err := model.NewProject("test", "test", "//test/...")
	assert.NoError(t, err)

	project.UpdateBaseline("main", []string{"fp1"}, nil, nil, nil)
	project.UpdateBaseline("dev", []string{"fp2"}, nil, nil, nil)

	all := project.Baselines()
	assert.Len(t, all, 2)
	assert.Equal(t, []string{"fp1"}, all["main"].Fingerprints())
	assert.Equal(t, []string{"fp2"}, all["dev"].Fingerprints())
}
