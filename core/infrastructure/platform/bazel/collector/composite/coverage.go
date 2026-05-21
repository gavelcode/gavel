package composite

import (
	"context"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
)

type CoverageCollector struct {
	primary  collectevidence.CoverageCollector
	fallback collectevidence.CoverageCollector
}

func NewCoverageCollector(primary, fallback collectevidence.CoverageCollector) *CoverageCollector {
	return &CoverageCollector{primary: primary, fallback: fallback}
}

func (c *CoverageCollector) CollectCoverage(ctx context.Context, workspace string, targets []string, languages []string) ([]byte, error) {
	data, err := c.primary.CollectCoverage(ctx, workspace, targets, languages)
	if err != nil {
		return nil, err
	}

	if hasUsefulCoverage(data) {
		return data, nil
	}

	if !hasTypeScript(languages) || c.fallback == nil {
		return data, nil
	}

	fallbackData, err := c.fallback.CollectCoverage(ctx, workspace, targets, languages)
	if err != nil {
		return data, nil
	}
	if fallbackData != nil {
		return fallbackData, nil
	}
	return data, nil
}

func hasUsefulCoverage(data []byte) bool {
	if data == nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "DA:") {
			return true
		}
	}
	return false
}

func hasTypeScript(languages []string) bool {
	for _, l := range languages {
		if l == "typescript" {
			return true
		}
	}
	return false
}
