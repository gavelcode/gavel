package sourceblob

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var ErrNotFound = failure.New("source blob not found", failure.NotFound)

type Storage struct {
	db *database.DB
}

func NewStorage(db *database.DB) *Storage {
	return &Storage{db: db}
}

func (r *Storage) Save(ctx context.Context, projectID, commitSHA, filePath string, content []byte, contentType string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO source_blobs (project_id, commit_sha, file_path, content, content_type, size_bytes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (project_id, commit_sha, file_path) DO NOTHING
	`, projectID, commitSHA, filePath, content, contentType, len(content))
	if err != nil {
		return fmt.Errorf("insert source blob: %w", err)
	}
	return nil
}

func (r *Storage) Fetch(ctx context.Context, projectID, commitSHA, filePath string) ([]byte, string, error) {
	var content []byte
	var contentType string
	err := r.db.QueryRowContext(ctx, `
		SELECT content, content_type FROM source_blobs
		WHERE project_id = ? AND commit_sha = ? AND file_path = ?
	`, projectID, commitSHA, filePath).Scan(&content, &contentType)
	if err == sql.ErrNoRows {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("query source blob: %w", err)
	}
	return content, contentType, nil
}
