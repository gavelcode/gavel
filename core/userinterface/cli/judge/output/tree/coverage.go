package tree

import (
	"math"
	"path"
	"sort"
)

const (
	percentScale   = 100
	decimalQuantum = 10
)

type FileCoverage struct {
	FilePath     string
	CoveredLines int
	TotalLines   int
}

type CoverageFile struct {
	Name         string  `json:"name"`
	CoveredLines int     `json:"covered_lines"`
	TotalLines   int     `json:"total_lines"`
	Percent      float64 `json:"percent"`
}

type CoverageNode struct {
	Path         string          `json:"path"`
	CoveredLines int             `json:"covered_lines"`
	TotalLines   int             `json:"total_lines"`
	Percent      float64         `json:"percent"`
	Children     []*CoverageNode `json:"children,omitempty"`
	Files        []CoverageFile  `json:"files,omitempty"`
}

func BuildCoverageTree(files []FileCoverage) *CoverageNode {
	root := &CoverageNode{}
	if len(files) == 0 {
		return root
	}

	nodes := make(map[string]*CoverageNode)
	nodes[""] = root

	for _, fileCov := range files {
		dir := path.Dir(fileCov.FilePath)
		name := path.Base(fileCov.FilePath)

		ensureDir(nodes, dir)
		node := nodes[dir]

		var pct float64
		if fileCov.TotalLines > 0 {
			pct = roundPercent(float64(fileCov.CoveredLines) / float64(fileCov.TotalLines) * percentScale)
		}
		node.Files = append(node.Files, CoverageFile{
			Name:         name,
			CoveredLines: fileCov.CoveredLines,
			TotalLines:   fileCov.TotalLines,
			Percent:      pct,
		})
	}

	aggregate(root)
	sortTree(root)
	return root
}

func ensureDir(nodes map[string]*CoverageNode, dir string) {
	if _, ok := nodes[dir]; ok {
		return
	}
	parent := path.Dir(dir)
	if parent == dir || parent == "." {
		parent = ""
	}
	ensureDir(nodes, parent)

	node := &CoverageNode{Path: dir}
	nodes[dir] = node
	nodes[parent].Children = append(nodes[parent].Children, node)
}

func aggregate(node *CoverageNode) {
	for _, fileCov := range node.Files {
		node.CoveredLines += fileCov.CoveredLines
		node.TotalLines += fileCov.TotalLines
	}
	for _, child := range node.Children {
		aggregate(child)
		node.CoveredLines += child.CoveredLines
		node.TotalLines += child.TotalLines
	}
	if node.TotalLines > 0 {
		node.Percent = roundPercent(float64(node.CoveredLines) / float64(node.TotalLines) * percentScale)
	}
}

func sortTree(node *CoverageNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Path < node.Children[j].Path
	})
	sort.Slice(node.Files, func(i, j int) bool {
		return node.Files[i].Name < node.Files[j].Name
	})
	for _, child := range node.Children {
		sortTree(child)
	}
}

func roundPercent(p float64) float64 {
	return math.Round(p*decimalQuantum) / decimalQuantum
}
