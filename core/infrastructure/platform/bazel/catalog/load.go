package catalog

import (
	"fmt"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

const catalogRunfile = "gavel_tools/lint/catalog.yaml"

var loaded *Catalog

func active() *Catalog {
	if loaded == nil {
		catalog, err := loadFromRunfiles()
		if err != nil {
			panic(fmt.Sprintf("catalog: %v", err))
		}
		loaded = catalog
	}
	return loaded
}

func loadFromRunfiles() (*Catalog, error) {
	path, err := runfiles.Rlocation(catalogRunfile)
	if err != nil {
		return nil, fmt.Errorf("locate %s: %w", catalogRunfile, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return ParseCatalog(data)
}

func SetCatalog(catalog *Catalog) {
	loaded = catalog
}
