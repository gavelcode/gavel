package catalog

import (
	_ "embed"
	"fmt"
)

// Embedded (not read from Bazel runfiles) so the standalone binary runs off a workspace it did not build.
//
//go:embed catalog.yaml
var embeddedCatalog []byte

var loaded *Catalog

var loader = loadEmbedded

func active() *Catalog {
	if loaded == nil {
		catalog, err := loader()
		if err != nil {
			panic(fmt.Sprintf("catalog: %v", err))
		}
		loaded = catalog
	}
	return loaded
}

func loadEmbedded() (*Catalog, error) {
	return ParseCatalog(embeddedCatalog)
}

func SetCatalog(catalog *Catalog) {
	loaded = catalog
}
