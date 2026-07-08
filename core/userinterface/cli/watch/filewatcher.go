package watch

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	root     string
	debounce time.Duration
	onChange func(context.Context, []string)
	onStart  func()
	log      *slog.Logger
}

func NewFileWatcher(root string, debounce time.Duration, onChange func(context.Context, []string), log *slog.Logger) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{watcher: w, root: root, debounce: debounce, onChange: onChange, log: log}, nil
}

// onStart fires only after the watch is registered, so no event is missed between the start signal and the watch.
func (fw *FileWatcher) Run(ctx context.Context) error {
	if err := fw.addRecursive(fw.root); err != nil {
		return err
	}
	if fw.onStart != nil {
		fw.onStart()
	}

	pending := make(map[string]struct{})
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	armed := false

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return fw.watcher.Close()

		case <-timer.C:
			armed = false
			files := make([]string, 0, len(pending))
			for f := range pending {
				files = append(files, f)
			}
			pending = make(map[string]struct{})
			if len(files) > 0 {
				fw.onChange(ctx, files)
			}

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			if shouldIgnore(event.Name, fw.root) {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				fw.tryAddDirectory(event.Name)
			}
			pending[event.Name] = struct{}{}
			if !armed {
				timer.Reset(fw.debounce)
				armed = true
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			fw.log.Warn("watcher error", "err", err)
		}
	}
}

func (fw *FileWatcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldIgnoreDir(d.Name()) {
				return filepath.SkipDir
			}
			return fw.watcher.Add(path)
		}
		return nil
	})
}

func (fw *FileWatcher) tryAddDirectory(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}
	if shouldIgnoreDir(filepath.Base(path)) {
		return
	}
	if err := fw.watcher.Add(path); err != nil {
		fw.log.Debug("failed to watch new directory", "path", path, "err", err)
	}
}

func shouldIgnore(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return true
	}
	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		if shouldIgnoreDir(p) {
			return true
		}
	}
	return false
}

func shouldIgnoreDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if strings.HasPrefix(name, "bazel-") {
		return true
	}
	return name == "node_modules"
}
