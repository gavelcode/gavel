package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	// maxBuildEventBytes caps one JSON line of the build event stream; a NamedSetOfFiles event can be large.
	maxBuildEventBytes     = 64 << 20
	submissionsOutputGroup = "gavel_submissions"
	failureOutputTailLines = 15
)

// BridgeConfig describes one project-declared aspect to build and collect.
type BridgeConfig struct {
	Workspace    string
	Targets      []string
	Aspect       string
	SARIFSuffix  string
	BuildFlags   []string
	BazelConfigs []string
}

type BridgeFile struct {
	// Name is the output's path relative to bazel-bin, for messages.
	Name string
	Data []byte
}

type BridgeResult struct {
	Files []BridgeFile
}

// RunBridge builds the bridge's aspect and returns the SARIF files that this
// very build reported in its gavel_submissions output group. It never globs
// bazel-bin, so a file left over from an earlier build cannot leak in. Any
// failure (build error, no matching output) is returned as an error whose
// message is meant to be shown to the user as the tool execution failure.
func RunBridge(ctx context.Context, config BridgeConfig) (*BridgeResult, error) {
	return runBridge(ctx, &ExecRunner{}, config)
}

func runBridge(ctx context.Context, cmd CommandRunner, config BridgeConfig) (*BridgeResult, error) {
	workDir, err := os.MkdirTemp("", "gavel-bridge-")
	if err != nil {
		return nil, fmt.Errorf("create bridge scratch dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()
	bepPath := filepath.Join(workDir, "bep.json")

	args, err := buildBridgeArgs(config, bepPath)
	if err != nil {
		return nil, err
	}
	_, stderr, runErr := cmd.Run(ctx, config.Workspace, "bazel", args...)
	if runErr != nil {
		return nil, fmt.Errorf("bazel build failed (%v):\n%s", runErr, tailLines(string(stderr), failureOutputTailLines))
	}

	reported, err := submissionFiles(bepPath)
	if err != nil {
		return nil, err
	}
	var result BridgeResult
	for _, file := range reported {
		if !strings.HasSuffix(file.name, config.SARIFSuffix) {
			continue
		}
		data, readErr := os.ReadFile(file.path)
		if readErr != nil {
			return nil, fmt.Errorf("read %s reported by the build: %w", file.name, readErr)
		}
		result.Files = append(result.Files, BridgeFile{Name: file.name, Data: data})
	}
	if len(result.Files) == 0 {
		return nil, fmt.Errorf("the build produced no SARIF files ending in %q in output group %s; the aspect must emit one per analyzed target (even when it has no findings)", config.SARIFSuffix, submissionsOutputGroup)
	}
	return &result, nil
}

func buildBridgeArgs(config BridgeConfig, bepPath string) ([]string, error) {
	args := []string{"build"}
	configArgs, err := bazelConfigArgs(config.BazelConfigs)
	if err != nil {
		return nil, err
	}
	args = append(args, configArgs...)
	args = append(args,
		"--aspects="+config.Aspect,
		"--output_groups="+submissionsOutputGroup,
		"--keep_going",
		"--build_event_json_file="+bepPath,
		// The reported outputs are read from disk, so a remote cache must materialize them.
		"--remote_download_toplevel",
	)
	args = append(args, config.BuildFlags...)
	args = append(args, "--")
	args = append(args, config.Targets...)
	return args, nil
}

type reportedFile struct {
	name string
	path string
}

type bepEvent struct {
	ID struct {
		NamedSet struct {
			ID string `json:"id"`
		} `json:"namedSet"`
		TargetCompleted *struct {
			Aspect string `json:"aspect"`
		} `json:"targetCompleted"`
	} `json:"id"`
	NamedSetOfFiles *struct {
		Files []struct {
			Name string `json:"name"`
			URI  string `json:"uri"`
		} `json:"files"`
		FileSets []struct {
			ID string `json:"id"`
		} `json:"fileSets"`
	} `json:"namedSetOfFiles"`
	Completed *struct {
		OutputGroup []struct {
			Name     string `json:"name"`
			FileSets []struct {
				ID string `json:"id"`
			} `json:"fileSets"`
		} `json:"outputGroup"`
	} `json:"completed"`
}

// submissionFiles resolves the files the build reported, for aspects, in the gavel_submissions group.
func submissionFiles(bepPath string) ([]reportedFile, error) {
	data, err := os.ReadFile(bepPath)
	if err != nil {
		return nil, fmt.Errorf("read build event stream: %w", err)
	}
	sets := make(map[string]*bepEvent)
	var roots []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 1<<20), maxBuildEventBytes)
	for scanner.Scan() {
		var event bepEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode build event: %w", err)
		}
		switch {
		case event.NamedSetOfFiles != nil:
			sets[event.ID.NamedSet.ID] = &event
		case event.ID.TargetCompleted != nil && event.ID.TargetCompleted.Aspect != "" && event.Completed != nil:
			for _, group := range event.Completed.OutputGroup {
				if group.Name != submissionsOutputGroup {
					continue
				}
				for _, set := range group.FileSets {
					roots = append(roots, set.ID)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read build event stream: %w", err)
	}

	var files []reportedFile
	seenSets := make(map[string]bool)
	seenFiles := make(map[string]bool)
	var visit func(setID string) error
	visit = func(setID string) error {
		if seenSets[setID] {
			return nil
		}
		seenSets[setID] = true
		set, ok := sets[setID]
		if !ok {
			return fmt.Errorf("build event stream references unknown file set %q", setID)
		}
		for _, file := range set.NamedSetOfFiles.Files {
			path, err := localPath(file.URI)
			if err != nil {
				return err
			}
			if !seenFiles[path] {
				seenFiles[path] = true
				files = append(files, reportedFile{name: file.Name, path: path})
			}
		}
		for _, child := range set.NamedSetOfFiles.FileSets {
			if err := visit(child.ID); err != nil {
				return err
			}
		}
		return nil
	}
	for _, root := range roots {
		if err := visit(root); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func localPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "file" {
		return "", fmt.Errorf("build reported an output that is not a local file: %q", uri)
	}
	return parsed.Path, nil
}

func tailLines(text string, count int) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}
	return strings.Join(lines, "\n")
}
