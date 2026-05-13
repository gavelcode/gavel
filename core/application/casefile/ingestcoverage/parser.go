package ingestcoverage

import "context"

type Parser interface {
	Parse(ctx context.Context, data []byte) (Parsed, error)
}
