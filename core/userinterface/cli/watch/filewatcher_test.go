package watch

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldIgnoreDir(t *testing.T) {
	cases := []struct {
		name   string
		dir    string
		expect bool
	}{
		{name: "dot prefix", dir: ".git", expect: true},
		{name: "hidden dir", dir: ".idea", expect: true},
		{name: "bazel output", dir: "bazel-out", expect: true},
		{name: "bazel bin", dir: "bazel-bin", expect: true},
		{name: "node_modules", dir: "node_modules", expect: true},
		{name: "regular dir", dir: "src", expect: false},
		{name: "pkg dir", dir: "pkg", expect: false},
		{name: "core dir", dir: "core", expect: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, shouldIgnoreDir(tc.dir))
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	root := "/workspace"
	cases := []struct {
		name   string
		path   string
		expect bool
	}{
		{name: "git file", path: "/workspace/.git/config", expect: true},
		{name: "bazel output", path: "/workspace/bazel-bin/main", expect: true},
		{name: "node_modules file", path: "/workspace/node_modules/pkg/index.js", expect: true},
		{name: "nested hidden", path: "/workspace/src/.cache/data", expect: true},
		{name: "regular source", path: "/workspace/src/main.go", expect: false},
		{name: "nested source", path: "/workspace/core/pkg/lib.go", expect: false},
		{name: "unrelated path", path: "/other/path/file.go", expect: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, shouldIgnore(tc.path, root))
		})
	}
}

func TestNewFileWatcherSuccess(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	watcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = watcher.watcher.Close() })

	assert.Equal(t, dir, watcher.root)
}

func TestFileWatcherRunContextCancellation(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	watcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- watcher.Run(ctx) }()

	cancel()
	require.NoError(t, <-errCh)
}

func TestFileWatcherRunDetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	received := make(chan []string, 1)
	onChange := func(_ context.Context, files []string) {
		received <- files
	}

	watcher, err := NewFileWatcher(dir, 10*time.Millisecond, onChange, log)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- watcher.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	testFile := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package x"), 0o644))

	select {
	case files := <-received:
		assert.Contains(t, files, testFile)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for file change event")
	}

	cancel()
	require.NoError(t, <-errCh)
}

func TestFileWatcherRunCreatesSubdirectoryWatch(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	received := make(chan []string, 10)
	onChange := func(_ context.Context, files []string) {
		received <- files
	}

	watcher, err := NewFileWatcher(dir, 10*time.Millisecond, onChange, log)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- watcher.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)

	subDir := filepath.Join(dir, "newpkg")
	require.NoError(t, os.Mkdir(subDir, 0o755))
	time.Sleep(50 * time.Millisecond)

	testFile := filepath.Join(subDir, "new.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package newpkg"), 0o644))

	found := false
	timeout := time.After(2 * time.Second)
	for !found {
		select {
		case files := <-received:
			for _, f := range files {
				if f == testFile {
					found = true
				}
			}
		case <-timeout:
			t.Fatal("timed out waiting for file in new subdirectory")
		}
	}

	cancel()
	require.NoError(t, <-errCh)
}

func TestFileWatcherRunIgnoresHiddenDirectories(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	received := make(chan []string, 10)
	onChange := func(_ context.Context, files []string) {
		received <- files
	}

	watcher, err := NewFileWatcher(dir, 10*time.Millisecond, onChange, log)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- watcher.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)

	hiddenDir := filepath.Join(dir, ".hidden")
	require.NoError(t, os.Mkdir(hiddenDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(hiddenDir, "file.go"), []byte("x"), 0o644))

	visibleFile := filepath.Join(dir, "visible.go")
	require.NoError(t, os.WriteFile(visibleFile, []byte("package x"), 0o644))

	select {
	case files := <-received:
		for _, f := range files {
			assert.NotContains(t, f, ".hidden")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for visible file event")
	}

	cancel()
	require.NoError(t, <-errCh)
}

func TestFileWatcherRunAddRecursiveError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fw, err := NewFileWatcher(t.TempDir(), 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)

	require.NoError(t, fw.watcher.Close())

	err = fw.Run(context.Background())
	require.Error(t, err)
}

func TestAddRecursiveWithFilesInTree(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg", "lib.go"), []byte("package pkg"), 0o644))

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fileWatcher.watcher.Close() })

	require.NoError(t, fileWatcher.addRecursive(dir))

	watchList := fileWatcher.watcher.WatchList()
	assert.Contains(t, watchList, dir)
	assert.Contains(t, watchList, filepath.Join(dir, "pkg"))
}

func TestTryAddDirectoryIgnoresHiddenDir(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fileWatcher.watcher.Close() })

	hiddenDir := filepath.Join(dir, ".cache")
	require.NoError(t, os.Mkdir(hiddenDir, 0o755))

	fileWatcher.tryAddDirectory(hiddenDir)

	watchList := fileWatcher.watcher.WatchList()
	assert.NotContains(t, watchList, hiddenDir)
}

func TestTryAddDirectoryIgnoresNonDirectory(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fileWatcher.watcher.Close() })

	filePath := filepath.Join(dir, "file.go")
	require.NoError(t, os.WriteFile(filePath, []byte("package x"), 0o644))

	fileWatcher.tryAddDirectory(filePath)

	watchList := fileWatcher.watcher.WatchList()
	assert.NotContains(t, watchList, filePath)
}

func TestTryAddDirectoryIgnoresNonExistentPath(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fileWatcher.watcher.Close() })

	fileWatcher.tryAddDirectory(filepath.Join(dir, "nonexistent"))
}

func TestFileWatcherRunExitsOnEventsChannelClosed(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)

	errCh := make(chan error, 1)
	go func() { errCh <- fileWatcher.Run(context.Background()) }()

	time.Sleep(50 * time.Millisecond)
	require.NoError(t, fileWatcher.watcher.Close())

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Run to exit after watcher close")
	}
}

func TestTryAddDirectoryWatcherAddError(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	subDir := filepath.Join(dir, "pkg")
	require.NoError(t, os.Mkdir(subDir, 0o755))

	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)

	require.NoError(t, fileWatcher.watcher.Close())

	fileWatcher.tryAddDirectory(subDir)
}

func TestAddRecursiveSkipsIgnoredDirs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src", "pkg"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bazel-out", "k8"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0o755))

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	fileWatcher, err := NewFileWatcher(dir, 10*time.Millisecond, func(_ context.Context, _ []string) {}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fileWatcher.watcher.Close() })

	require.NoError(t, fileWatcher.addRecursive(dir))

	watchList := fileWatcher.watcher.WatchList()
	assert.Contains(t, watchList, dir)
	assert.Contains(t, watchList, filepath.Join(dir, "src"))
	assert.Contains(t, watchList, filepath.Join(dir, "src", "pkg"))
	assert.NotContains(t, watchList, filepath.Join(dir, ".git"))
	assert.NotContains(t, watchList, filepath.Join(dir, ".git", "objects"))
	assert.NotContains(t, watchList, filepath.Join(dir, "bazel-out"))
	assert.NotContains(t, watchList, filepath.Join(dir, "node_modules"))
}
