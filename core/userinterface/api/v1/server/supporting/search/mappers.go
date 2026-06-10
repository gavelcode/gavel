package search

import (
	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
)

func resultFromQuery(result searchquery.SearchResult) gen.SearchResult {
	return gen.SearchResult{
		Type:     result.Type,
		Id:       result.ID,
		Title:    result.Title,
		Subtitle: result.Subtitle,
		Url:      result.URL,
	}
}
