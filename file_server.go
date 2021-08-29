package mux

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

// CacheControlFS represents the ability to associate
// a Cache-Control response header with a file name.
type CacheControlFS interface {
	fs.FS
	CacheControl(name string) string
}

// assetCacheFS implements the CacheControlFS interface.
type assetCacheFS struct {
	fs.FS
}

// CacheControl implements the CacheControlFS interface.
func (c *assetCacheFS) CacheControl(name string) string {
	return "public, max-age=31536000"
}

// AssetCacheFS returns the fs as an implementation of CacheControlFS.
// All files will be cached with the "public, max-age=31536000" policy.
func AssetCacheFS(fs fs.FS) fs.FS {
	return &assetCacheFS{fs}
}

// fileServer is a simplified file server.
type fileServer struct {
	h  *Handler
	fs fs.FS
}

// ServeHTTP implements the http.Handler interface.
func (h *fileServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := path.Clean(req.URL.Path)
	name = strings.TrimPrefix(name, "/")
	f, err := h.fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			h.h.Abort(w, req, ErrNotFound)
			return
		}
		h.h.Abort(w, req, err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		h.h.Abort(w, req, err)
		return
	}
	if fi.IsDir() {
		h.h.Abort(w, req, ErrNotFound)
		return
	}
	cfs, ok := h.fs.(CacheControlFS)
	if ok {
		w.Header().Set("Cache-Control", cfs.CacheControl(name))
	}
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
	w.WriteHeader(http.StatusOK)
	if req.Method == http.MethodHead {
		return
	}
	io.Copy(w, f)
}
