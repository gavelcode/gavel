package tree

import (
	"path"
	"sort"
)

type FindingInput struct {
	FilePath string
	Severity string
}

type FindingsFile struct {
	Name       string         `json:"name"`
	Count      int            `json:"count"`
	BySeverity map[string]int `json:"by_severity"`
}

type FindingsNode struct {
	Path       string          `json:"path"`
	Count      int             `json:"count"`
	BySeverity map[string]int  `json:"by_severity"`
	Children   []*FindingsNode `json:"children,omitempty"`
	Files      []FindingsFile  `json:"files,omitempty"`
}

func BuildFindingsTree(findings []FindingInput) *FindingsNode {
	root := &FindingsNode{BySeverity: make(map[string]int)}
	if len(findings) == 0 {
		return root
	}

	nodes := make(map[string]*FindingsNode)
	nodes[""] = root

	fileIndex := make(map[string]map[string]*FindingsFile)

	for _, finding := range findings {
		dir := path.Dir(finding.FilePath)
		name := path.Base(finding.FilePath)

		ensureFindingsDir(nodes, dir)

		if fileIndex[dir] == nil {
			fileIndex[dir] = make(map[string]*FindingsFile)
		}
		findings, ok := fileIndex[dir][name]
		if !ok {
			findings = &FindingsFile{Name: name, BySeverity: make(map[string]int)}
			fileIndex[dir][name] = findings
			nodes[dir].Files = append(nodes[dir].Files, FindingsFile{})
		}
		findings.Count++
		findings.BySeverity[finding.Severity]++
	}

	for dir, fileMap := range fileIndex {
		node := nodes[dir]
		node.Files = node.Files[:0]
		for _, findings := range fileMap {
			node.Files = append(node.Files, *findings)
		}
	}

	aggregateFindings(root)
	sortFindingsTree(root)
	return root
}

func ensureFindingsDir(nodes map[string]*FindingsNode, dir string) {
	if _, ok := nodes[dir]; ok {
		return
	}
	parent := path.Dir(dir)
	if parent == dir || parent == "." {
		parent = ""
	}
	ensureFindingsDir(nodes, parent)

	node := &FindingsNode{Path: dir, BySeverity: make(map[string]int)}
	nodes[dir] = node
	nodes[parent].Children = append(nodes[parent].Children, node)
}

func aggregateFindings(node *FindingsNode) {
	for _, finding := range node.Files {
		node.Count += finding.Count
		for sev, cnt := range finding.BySeverity {
			node.BySeverity[sev] += cnt
		}
	}
	for _, child := range node.Children {
		aggregateFindings(child)
		node.Count += child.Count
		for sev, cnt := range child.BySeverity {
			node.BySeverity[sev] += cnt
		}
	}
}

func sortFindingsTree(node *FindingsNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Path < node.Children[j].Path
	})
	sort.Slice(node.Files, func(i, j int) bool {
		return node.Files[i].Name < node.Files[j].Name
	})
	for _, child := range node.Children {
		sortFindingsTree(child)
	}
}
