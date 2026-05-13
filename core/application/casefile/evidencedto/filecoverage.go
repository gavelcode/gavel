package evidencedto

import "sort"

type FileCoverage struct {
	FilePath  string
	Covered   []int
	Uncovered []int
}

func FileCoverageFromPerLine(perLine map[string]map[int]int) []FileCoverage {
	out := make([]FileCoverage, 0, len(perLine))
	for path, lines := range perLine {
		out = append(out, fileCoverage(path, lines))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FilePath < out[j].FilePath })
	return out
}

func fileCoverage(path string, lines map[int]int) FileCoverage {
	covered := make([]int, 0, len(lines))
	uncovered := make([]int, 0)
	for line, hits := range lines {
		if hits > 0 {
			covered = append(covered, line)
			continue
		}
		uncovered = append(uncovered, line)
	}
	sort.Ints(covered)
	sort.Ints(uncovered)
	return FileCoverage{FilePath: path, Covered: covered, Uncovered: uncovered}
}
