package spa

import (
	"io/fs"
	"net/http"
	"strings"
)

func Handler(frontend fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(frontend))
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(frontend, path); err == nil {
			fileServer.ServeHTTP(writer, req)
			return
		}

		req.URL.Path = "/"
		fileServer.ServeHTTP(writer, req)
	})
}
