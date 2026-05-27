package spectest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIRefsAreResolvable(t *testing.T) {
	r, err := runfiles.New()
	if err != nil {
		t.Fatalf("init runfiles: %v", err)
	}
	specDir, err := r.Rlocation("_main/openapi/v1")
	if err != nil {
		t.Fatalf("locate openapi/v1: %v", err)
	}

	entries, err := os.ReadDir(specDir)
	if err != nil {
		t.Fatalf("read spec dir: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, err := filepath.Glob(filepath.Join(specDir, entry.Name(), "*.yaml"))
		if err != nil {
			t.Fatalf("glob %s: %v", entry.Name(), err)
		}
		for _, file := range files {
			t.Run(filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file), func(t *testing.T) {
				validateFileRefs(t, file)
			})
		}
	}
}

func validateFileRefs(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	localComponents := collectLocalComponents(&doc)
	refs := collectRefs(&doc)

	for _, ref := range refs {
		if strings.HasPrefix(ref.value, "#/") {
			component := extractComponentName(ref.value)
			if _, ok := localComponents[component]; !ok {
				t.Errorf("line %d: broken internal ref %q — component not defined in this file", ref.line, ref.value)
			}
		} else if strings.Contains(ref.value, "#/") {
			filePart, _ := splitRef(ref.value)
			targetPath := filepath.Join(filepath.Dir(path), filePart)
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				t.Errorf("line %d: broken cross-file ref %q — file %s does not exist", ref.line, ref.value, targetPath)
			}
		}
	}
}

type refEntry struct {
	value string
	line  int
}

func collectRefs(node *yaml.Node) []refEntry {
	var refs []refEntry
	walkRefs(node, &refs)
	return refs
}

func walkRefs(node *yaml.Node, refs *[]refEntry) {
	if node == nil {
		return
	}
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			walkRefs(child, refs)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, val := node.Content[i], node.Content[i+1]
			if key.Value == "$ref" && val.Kind == yaml.ScalarNode {
				*refs = append(*refs, refEntry{value: val.Value, line: val.Line})
			} else {
				walkRefs(val, refs)
			}
		}
	}
}

func collectLocalComponents(doc *yaml.Node) map[string]bool {
	components := make(map[string]bool)

	root := doc
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		root = doc.Content[0]
	}

	compNode := mapLookup(root, "components")
	if compNode == nil {
		return components
	}

	for _, section := range []string{"schemas", "parameters", "responses"} {
		sectionNode := mapLookup(compNode, section)
		if sectionNode == nil || sectionNode.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i+1 < len(sectionNode.Content); i += 2 {
			name := sectionNode.Content[i].Value
			components["#/components/"+section+"/"+name] = true
		}
	}

	return components
}

func mapLookup(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func extractComponentName(ref string) string {
	return ref
}

func splitRef(ref string) (file, pointer string) {
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return ref, ""
}
