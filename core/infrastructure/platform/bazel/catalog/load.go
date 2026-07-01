package catalog

import (
	"fmt"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// catalogRunfile is the runfiles path to the gavel-tools catalog, the single
// source of truth for which tools exist per language.
const catalogRunfile = "gavel_tools/lint/catalog.yaml"

// loaded caches the active catalog. It is process-global like modulePrefix; the
// catalog is immutable for a run.
var loaded *Catalog

// active returns the loaded catalog, reading it from runfiles on first use. A
// missing or malformed catalog is fatal — gavel cannot decide what to run
// without it.
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

// SetCatalog installs a catalog explicitly, bypassing runfiles. Tests use it to
// exercise the selection logic with a fixed catalog.
func SetCatalog(catalog *Catalog) {
	loaded = catalog
}
