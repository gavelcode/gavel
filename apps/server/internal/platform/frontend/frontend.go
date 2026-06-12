package frontend

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func Load() (fs.FS, error) {
	path, err := runfiles.Rlocation("_main/apps/web/dist/index.html")
	if err != nil {
		return nil, fmt.Errorf("locate frontend assets: %w", err)
	}
	dir := path[:len(path)-len("/index.html")]
	return os.DirFS(dir), nil
}
