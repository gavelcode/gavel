package sarif

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileSourceReader struct {
	workspace string
}

func NewFileSourceReader(workspace string) *FileSourceReader {
	return &FileSourceReader{workspace: workspace}
}

func (r *FileSourceReader) ReadLine(filePath string, line int) (string, error) {
	if line <= 0 {
		return "", fmt.Errorf("line must be positive, got %d", line)
	}
	file, err := os.Open(filepath.Join(r.workspace, filePath))
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	current := 0
	for scanner.Scan() {
		current++
		if current == line {
			return strings.TrimSpace(scanner.Text()), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("line %d out of range (file has %d lines)", line, current)
}
