package catalog

import (
	_ "embed"
	"fmt"
)

// catalog.yaml is synced from @gavel_tools//lint/catalog.yaml by `make
// catalog-sync` and kept in lockstep by `make catalog-check`. Embedding it —
// rather than reading it from Bazel runfiles — is what lets the distributed
// standalone binary run `init` and `judge` off a Bazel workspace it did not
// build itself.
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
