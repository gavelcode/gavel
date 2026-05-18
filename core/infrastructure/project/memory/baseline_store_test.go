package memory_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/project/memory"
)

func TestBaselineStore_SaveAndLoad(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	bl := model.NewBaseline([]string{"fp2", "fp1"}, []string{"arch1"}, nil, nil)
	err := store.Save("backend", bl)
	require.NoError(t, err)

	loaded := store.Load("backend")
	assert.True(t, loaded.HasPrevious())
	assert.Equal(t, []string{"fp1", "fp2"}, loaded.Fingerprints())
	assert.Equal(t, []string{"arch1"}, loaded.ArchIDs())
}

func TestBaselineStore_SaveAndLoadWithCoverage(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	cov := 57.5
	bl := model.NewBaseline([]string{"fp1"}, nil, &cov, nil)
	err := store.Save("backend", bl)
	require.NoError(t, err)

	loaded := store.Load("backend")
	assert.True(t, loaded.HasPrevious())
	require.NotNil(t, loaded.CoveragePercent())
	assert.InDelta(t, 57.5, *loaded.CoveragePercent(), 0.001)
}

func TestBaselineStore_LoadMissingCoverage(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	bl := model.NewBaseline([]string{"fp1"}, nil, nil, nil)
	err := store.Save("backend", bl)
	require.NoError(t, err)

	loaded := store.Load("backend")
	assert.Nil(t, loaded.CoveragePercent())
}

func TestBaselineStore_LoadMissing(t *testing.T) {
	store := memory.NewBaselineStore(t.TempDir())

	loaded := store.Load("nonexistent")
	assert.False(t, loaded.HasPrevious())
}

func TestBaselineStore_SaveOnlyDefaultBranchBaseline(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	defaultBL := model.NewBaseline([]string{"main-fp1", "main-fp2"}, []string{"arch-main"}, nil, nil)
	err := store.Save("backend", defaultBL)
	require.NoError(t, err)

	loaded := store.Load("backend")
	assert.True(t, loaded.HasPrevious())
	assert.Equal(t, []string{"main-fp1", "main-fp2"}, loaded.Fingerprints())
	assert.Equal(t, []string{"arch-main"}, loaded.ArchIDs())
}

func TestBaselineStore_LoadWithoutBranchParameter(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	bl := model.NewBaseline([]string{"fp-a", "fp-b"}, nil, nil, nil)
	require.NoError(t, store.Save("myproject", bl))

	loaded := store.Load("myproject")
	assert.True(t, loaded.HasPrevious())
	assert.Equal(t, []string{"fp-a", "fp-b"}, loaded.Fingerprints())
}

func TestBaselineStore_SaveReturnsErrorWhenMkdirFails(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "nonexistent", "deep", "path")
	require.NoError(t, os.MkdirAll(workspace, 0o755))
	require.NoError(t, os.Chmod(workspace, 0o444))
	t.Cleanup(func() { _ = os.Chmod(workspace, 0o755) })

	store := memory.NewBaselineStore(workspace)
	bl := model.NewBaseline([]string{"fp1"}, nil, nil, nil)
	err := store.Save("proj", bl)
	require.Error(t, err)
}

func TestBaselineStore_SaveReturnsErrorWhenFindingsWriteFails(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	bl := model.NewBaseline([]string{"fp1"}, nil, nil, nil)
	err := store.Save("proj", bl)
	require.Error(t, err)
}

func TestBaselineStore_SaveReturnsErrorWhenArchitectureWriteFails(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	bl := model.NewBaseline(nil, []string{"arch1"}, nil, nil)

	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := store.Save("proj", bl)
	require.Error(t, err)
}

func TestBaselineStore_SaveReturnsErrorWhenCoverageWriteFails(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	cov := 85.0
	bl := model.NewBaseline(nil, nil, &cov, nil)

	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := store.Save("proj", bl)
	require.Error(t, err)
}

func TestBaselineStore_LoadReturnsNilCoverageForEmptyFile(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage"), []byte("   \n"), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	assert.Nil(t, loaded.CoveragePercent())
}

func TestBaselineStore_LoadReturnsNilCoverageForNonNumericContent(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage"), []byte("abc\n"), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	assert.Nil(t, loaded.CoveragePercent())
}

func TestBaselineStore_LoadReturnsNilFingerprintsForEmptyFile(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "findings"), []byte("  \n"), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	assert.Nil(t, loaded.Fingerprints())
}

func TestBaselineStore_SaveAndLoadFileCoverage(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	entry1, err := model.NewFileCoverageEntry("pkg/a.go", []int{1, 2, 3}, []int{4, 5})
	require.NoError(t, err)
	entry2, err := model.NewFileCoverageEntry("pkg/b.go", []int{10}, nil)
	require.NoError(t, err)

	cov := 75.0
	bl := model.NewBaseline([]string{"fp1"}, nil, &cov, []model.FileCoverageEntry{entry1, entry2})
	require.NoError(t, store.Save("backend", bl))

	loaded := store.Load("backend")
	require.NotNil(t, loaded.CoveragePercent())
	assert.InDelta(t, 75.0, *loaded.CoveragePercent(), 0.001)

	fileCoverage := loaded.FileCoverage()
	require.Len(t, fileCoverage, 2)
	assert.Equal(t, "pkg/a.go", fileCoverage[0].FilePath())
	assert.Equal(t, []int{1, 2, 3}, fileCoverage[0].Covered())
	assert.Equal(t, []int{4, 5}, fileCoverage[0].Uncovered())
	assert.Equal(t, "pkg/b.go", fileCoverage[1].FilePath())
	assert.Equal(t, []int{10}, fileCoverage[1].Covered())
	assert.Nil(t, fileCoverage[1].Uncovered())
}

func TestBaselineStore_LoadFileCoverageMissing(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	bl := model.NewBaseline([]string{"fp1"}, nil, nil, nil)
	require.NoError(t, store.Save("backend", bl))

	loaded := store.Load("backend")
	assert.Nil(t, loaded.FileCoverage())
}

func TestBaselineStore_SaveClearsStaleArchWhenEmpty(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	require.NoError(t, store.Save("backend", model.NewBaseline([]string{"fp1"}, []string{"v1"}, nil, nil)))
	require.NoError(t, store.Save("backend", model.NewBaseline([]string{"fp1"}, nil, nil, nil)))

	loaded := store.Load("backend")
	assert.Nil(t, loaded.ArchIDs())
	assert.Equal(t, []string{"fp1"}, loaded.Fingerprints())
}

func TestBaselineStore_SaveClearsStaleFindingsWhenEmpty(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	require.NoError(t, store.Save("backend", model.NewBaseline([]string{"fp1"}, []string{"v1"}, nil, nil)))
	require.NoError(t, store.Save("backend", model.NewBaseline(nil, []string{"v1"}, nil, nil)))

	loaded := store.Load("backend")
	assert.Nil(t, loaded.Fingerprints())
	assert.Equal(t, []string{"v1"}, loaded.ArchIDs())
}

func TestBaselineStore_LoadLegacyCoverageFile(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "findings"), []byte("fp1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage"), []byte("85.5\n"), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	require.NotNil(t, loaded.CoveragePercent())
	assert.InDelta(t, 85.5, *loaded.CoveragePercent(), 0.001)
	assert.Nil(t, loaded.FileCoverage())
}

func TestBaselineStore_SaveReturnsErrorWhenFileCoverageWriteFails(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	entry, err := model.NewFileCoverageEntry("f.go", []int{1}, nil)
	require.NoError(t, err)

	cov := 50.0
	bl := model.NewBaseline(nil, nil, &cov, []model.FileCoverageEntry{entry})

	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err = store.Save("proj", bl)
	require.Error(t, err)
}

func TestBaselineStore_LoadCorruptedCoverageJSON(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage.json"), []byte("{invalid json"), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	assert.False(t, loaded.HasPrevious())
}

func TestBaselineStore_LoadCoverageJSONWithInvalidFilePath(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "baseline", "proj")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	data := `{"percent":80.0,"by_file":[{"file_path":"","covered":[1],"uncovered":[2]}]}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage.json"), []byte(data), 0o644))

	store := memory.NewBaselineStore(workspace)
	loaded := store.Load("proj")
	require.NotNil(t, loaded.CoveragePercent())
	assert.InDelta(t, 80.0, *loaded.CoveragePercent(), 0.001)
	assert.Nil(t, loaded.FileCoverage())
}

func TestProjectRepo_WithBaseline_PersistsAcrossInstances(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)

	project, err := model.NewProject("backend", "backend", "//server/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp1", "fp2"}, []string{"arch1"}, nil, nil)

	repo1 := memory.NewProjectRepositoryWithBaseline(store)
	require.NoError(t, repo1.Save(context.Background(), project))

	data, err := os.ReadFile(filepath.Join(workspace, ".gavel", "baseline", "backend", "findings"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "fp1")

	repo2 := memory.NewProjectRepositoryWithBaseline(store)
	project2, err := model.NewProject("backend", "backend", "//server/...")
	require.NoError(t, err)
	require.NoError(t, repo2.Save(context.Background(), project2))

	loaded, err := repo2.FindByID(context.Background(), project2.ID())
	require.NoError(t, err)
	bl := loaded.Baseline("main")
	assert.True(t, bl.HasPrevious())
	assert.Equal(t, []string{"fp1", "fp2"}, bl.Fingerprints())
}
