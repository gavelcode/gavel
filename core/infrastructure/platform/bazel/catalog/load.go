package catalog

import (
	"fmt"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

const catalogRunfile = "gavel_tools/lint/catalog.yaml"

var loaded *Catalog

var loader = loadFromRunfiles

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

func loadFromRunfiles() (*Catalog, error) {
	return loadCatalog(runfiles.Rlocation, os.ReadFile)
}

func loadCatalog(resolvePath func(string) (string, error), readFile func(string) ([]byte, error)) (*Catalog, error) {
	path, err := resolvePath(catalogRunfile)
	if err != nil {
		return nil, fmt.Errorf("locate %s: %w", catalogRunfile, err)
	}
	data, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return ParseCatalog(data)
}

func SetCatalog(catalog *Catalog) {
	loaded = catalog
}
