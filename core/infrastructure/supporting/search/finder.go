package search

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type Finder struct {
	db *database.DB
}

func NewFinder(db *database.DB) *Finder {
	return &Finder{db: db}
}

func (q *Finder) Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	pattern := "%" + database.EscapeLike(query) + "%"

	rows, err := q.db.QueryContext(ctx, searchSQL, pattern, pattern, pattern, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []search.SearchResult
	for rows.Next() {
		var r search.SearchResult
		if err := rows.Scan(&r.Type, &r.ID, &r.Title, &r.Subtitle, &r.URL); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

const searchSQL = `
SELECT type, id, title, subtitle, url FROM (
    SELECT
        'project' AS type,
        CAST(p.id AS TEXT),
        p.name AS title,
        p.key AS subtitle,
        '/projects/' || p.key AS url
    FROM projects p
    WHERE p.name LIKE ? ESCAPE '\' OR p.key LIKE ? ESCAPE '\'

    UNION ALL

    SELECT
        'casefile' AS type,
        CAST(c.id AS TEXT),
        SUBSTR(c.commit_sha, 1, 7) AS title,
        c.branch || ' · ' || COALESCE(c.verdict_outcome, 'pending') AS subtitle,
        '/casefiles/' || c.id AS url
    FROM casefiles c
    WHERE c.commit_sha LIKE ? ESCAPE '\' OR c.branch LIKE ? ESCAPE '\'

    UNION ALL

    SELECT
        'finding' AS type,
        CAST(f.id AS TEXT),
        f.rule_id AS title,
        f.file_path || ':' || CAST(f.line AS TEXT) AS subtitle,
        '/casefiles/' || f.casefile_id AS url
    FROM findings f
    WHERE f.rule_id LIKE ? ESCAPE '\' OR f.file_path LIKE ? ESCAPE '\'
) results
LIMIT ?`
