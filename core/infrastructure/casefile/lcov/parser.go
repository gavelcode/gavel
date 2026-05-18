package lcov

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Parser struct{}

var _ ingestcoverage.Parser = (*Parser)(nil)

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParsePerLine(data []byte) (map[string]map[int]int, error) {
	return ParsePerLine(data)
}

func (p *Parser) Parse(_ context.Context, data []byte) (ingestcoverage.Parsed, error) {
	if len(data) == 0 {
		return ingestcoverage.Parsed{}, nil
	}

	perLine, err := ParsePerLine(data)
	if err != nil {
		return ingestcoverage.Parsed{}, err
	}

	totalLines, coveredLines, langAccum, langOrder := tallyFromPerLine(perLine)

	suppTotal, suppCovered, suppLangs, suppOrder, err := supplementFromLFLH(data, perLine)
	if err != nil {
		return ingestcoverage.Parsed{}, err
	}
	totalLines += suppTotal
	coveredLines += suppCovered
	mergeLangAccum(langAccum, &langOrder, suppLangs, suppOrder)

	byLanguage, err := buildLanguageStats(langAccum, langOrder)
	if err != nil {
		return ingestcoverage.Parsed{}, err
	}

	return ingestcoverage.Parsed{
		TotalLines:   totalLines,
		CoveredLines: coveredLines,
		ByLanguage:   byLanguage,
		ByFile:       evidencedto.FileCoverageFromPerLine(perLine),
	}, nil
}

func tallyFromPerLine(perLine map[string]map[int]int) (int, int, map[string]*langTotals, []string) {
	accum := make(map[string]*langTotals)
	var order []string
	var totalLines, coveredLines int

	for filePath, lines := range perLine {
		lang := languageFromPath(filePath)
		if _, exists := accum[lang]; !exists {
			accum[lang] = &langTotals{}
			order = append(order, lang)
		}
		acc := accum[lang]
		for _, hitCount := range lines {
			totalLines++
			acc.total++
			if hitCount > 0 {
				coveredLines++
				acc.covered++
			}
		}
	}
	sort.Strings(order)
	return totalLines, coveredLines, accum, order
}

func supplementFromLFLH(data []byte, perLine map[string]map[int]int) (int, int, map[string]*langTotals, []string, error) {
	accum := make(map[string]*langTotals)
	var order []string
	var totalLines, coveredLines int

	seen := make(map[string]bool)
	var currentFile string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "SF:"):
			currentFile = strings.TrimPrefix(line, "SF:")
		case strings.HasPrefix(line, "LF:"):
			if seen[currentFile] {
				continue
			}
			if currentFile != "" && len(perLine[currentFile]) > 0 {
				seen[currentFile] = true
				continue
			}
			count, err := parseCount(line, "LF:")
			if err != nil {
				return 0, 0, nil, nil, fmt.Errorf("line count: %w", err)
			}
			totalLines += count
			lang := languageFromPath(currentFile)
			if currentFile != "" {
				if _, exists := accum[lang]; !exists {
					accum[lang] = &langTotals{}
					order = append(order, lang)
				}
				accum[lang].total += count
			}
			seen[currentFile] = true
		case strings.HasPrefix(line, "LH:"):
			if currentFile != "" && len(perLine[currentFile]) > 0 {
				continue
			}
			count, err := parseCount(line, "LH:")
			if err != nil {
				return 0, 0, nil, nil, fmt.Errorf("hit count: %w", err)
			}
			coveredLines += count
			lang := languageFromPath(currentFile)
			if currentFile != "" {
				if acc, ok := accum[lang]; ok {
					acc.covered += count
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, nil, nil, fmt.Errorf("%w: %w", ErrScanLCOV, err)
	}
	return totalLines, coveredLines, accum, order, nil
}

func mergeLangAccum(dst map[string]*langTotals, dstOrder *[]string, src map[string]*langTotals, srcOrder []string) {
	for _, lang := range srcOrder {
		acc := src[lang]
		if existing, ok := dst[lang]; ok {
			existing.total += acc.total
			existing.covered += acc.covered
		} else {
			dst[lang] = acc
			*dstOrder = append(*dstOrder, lang)
		}
	}
}

func buildLanguageStats(accum map[string]*langTotals, order []string) ([]coverage.LanguageStats, error) {
	result := make([]coverage.LanguageStats, 0, len(order))
	for _, langName := range order {
		acc := accum[langName]
		lang, err := coverage.NewLanguage(langName)
		if err != nil {
			return nil, fmt.Errorf("language %q: %w", langName, err)
		}
		stats, err := coverage.NewLanguageStats(lang, acc.total, acc.covered)
		if err != nil {
			return nil, fmt.Errorf("language %q: %w", langName, err)
		}
		result = append(result, stats)
	}
	return result, nil
}
