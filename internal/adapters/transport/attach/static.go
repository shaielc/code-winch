package attach

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func newStaticHandler(root string) (http.Handler, error) {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("static page unavailable")
	}
	info, err := os.Stat(filepath.Join(root, "index.html"))
	if err != nil || info.IsDir() {
		return nil, fmt.Errorf("static page unavailable")
	}
	files := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := filepath.Clean("/" + r.URL.Path)
		candidate := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
		if rel, err := filepath.Rel(root, candidate); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			resolved, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			rel, err := filepath.Rel(realRoot, resolved)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				http.NotFound(w, r)
				return
			}
			clone := r.Clone(r.Context())
			clone.URL.Path = clean
			files.ServeHTTP(w, clone)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	}), nil
}
