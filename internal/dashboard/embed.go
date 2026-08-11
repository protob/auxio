package dashboard

import (
	"io/fs"
	"net/http"
	"strings"
)

type SPAHandler struct {
	FS      fs.FS
	Handler http.Handler
}

func (h SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	if _, err := h.FS.Open(strings.TrimPrefix(path, "/")); err != nil {
		r.URL.Path = "/"
	}

	h.Handler.ServeHTTP(w, r)
}
