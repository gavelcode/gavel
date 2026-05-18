package lcov

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func ParsePerLine(data []byte) (map[string]map[int]int, error) {
	if len(data) == 0 {
		return nil, nil
	}

	result := make(map[string]map[int]int)
	var currentFile string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "SF:"):
			currentFile = strings.TrimPrefix(line, "SF:")
		case strings.HasPrefix(line, "DA:"):
			if currentFile == "" {
				continue
			}
			lineNum, hitCount, err := parseDARecord(line)
			if err != nil {
				return nil, err
			}
			if result[currentFile] == nil {
				result[currentFile] = make(map[int]int)
			}
			if existing, exists := result[currentFile][lineNum]; !exists || hitCount > existing {
				result[currentFile][lineNum] = hitCount
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrScanLCOV, err)
	}
	return result, nil
}

const (
	daSplitLimit = 3
	daMinParts   = 2
)

func parseDARecord(line string) (int, int, error) {
	parts := strings.SplitN(strings.TrimPrefix(line, "DA:"), ",", daSplitLimit)
	if len(parts) < daMinParts {
		return 0, 0, fmt.Errorf("%w: DA: expected line,count", ErrInvalidLine)
	}
	lineNum, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("%w: DA line number: %w", ErrInvalidLine, err)
	}
	hitCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("%w: DA hit count: %w", ErrInvalidLine, err)
	}
	return lineNum, hitCount, nil
}
