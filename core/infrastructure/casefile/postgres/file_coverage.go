package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type FileCoverageStore struct {
	db *database.DB
}

func NewFileCoverageStore(db *database.DB) *FileCoverageStore {
	return &FileCoverageStore{db: db}
}

func (s *FileCoverageStore) Save(ctx context.Context, caseFileID string, entries []evidencedto.FileCoverage) error {
	if len(entries) == 0 {
		return nil
	}
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	for _, entry := range entries {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO casefile_file_coverage (casefile_id, file_path, covered_lines, uncovered_lines)
			VALUES (?, ?, ?::integer[], ?::integer[])
			ON CONFLICT (casefile_id, file_path) DO UPDATE SET
				covered_lines = EXCLUDED.covered_lines,
				uncovered_lines = EXCLUDED.uncovered_lines
		`, caseFileID, entry.FilePath, formatPgArray(entry.Covered), formatPgArray(entry.Uncovered))
		if err != nil {
			return fmt.Errorf("upsert file coverage %s: %w", entry.FilePath, err)
		}
	}
	return transaction.Commit()
}

func (s *FileCoverageStore) Fetch(ctx context.Context, caseFileID, filePath string) (*evidencedto.FileCoverage, error) {
	var coveredStr, uncoveredStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT covered_lines::text, uncovered_lines::text FROM casefile_file_coverage
		WHERE casefile_id = ? AND file_path = ?
	`, caseFileID, filePath).Scan(&coveredStr, &uncoveredStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetch file coverage: %w", err)
	}
	covered, err := parsePgArray(coveredStr)
	if err != nil {
		return nil, fmt.Errorf("parse covered_lines: %w", err)
	}
	uncovered, err := parsePgArray(uncoveredStr)
	if err != nil {
		return nil, fmt.Errorf("parse uncovered_lines: %w", err)
	}
	return &evidencedto.FileCoverage{
		FilePath:  filePath,
		Covered:   covered,
		Uncovered: uncovered,
	}, nil
}

func formatPgArray(ints []int) string {
	if len(ints) == 0 {
		return "{}"
	}
	parts := make([]string, len(ints))
	for i, v := range ints {
		parts[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func parsePgArray(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "{}" || s == "" {
		return nil, nil
	}
	inner := strings.Trim(s, "{}")
	parts := strings.Split(inner, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("parse array element %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}
