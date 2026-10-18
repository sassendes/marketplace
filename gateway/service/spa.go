package service

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type SPA struct {
	files   fs.FS
	index   []byte
	modTime time.Time
	server  http.Handler
}

func NewSPA(files fs.FS) *SPA {
	index, _ := fs.ReadFile(files, "index.html")
	return &SPA{files: files, index: index, modTime: time.Now(), server: http.FileServerFS(files)}
}

func (s *SPA) Available() bool { return len(s.index) > 0 }

func (s *SPA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if !s.Available() {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("storefront assets were not built into this binary\n"))
		return
	}
	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if p != "" && p != "index.html" {
		if st, err := fs.Stat(s.files, p); err == nil && !st.IsDir() {
			if strings.HasPrefix(p, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}
			s.server.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(p, "assets/") || path.Ext(p) != "" {
			http.NotFound(w, r)
			return
		}
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", s.modTime, bytes.NewReader(s.index))
}
