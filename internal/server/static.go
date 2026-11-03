package server

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed dist/*
var embeddedDist embed.FS

func (s *Server) handleServeWidget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Try reading embedded widget
	data, err := embeddedDist.ReadFile("dist/widget/widget.js")
	if err == nil && len(data) > 0 {
		_, _ = w.Write(data)
		return
	}

	// Fallback minimal stub if not yet compiled
	stub := []byte(`console.log("[SearchCraft] Widget loaded");`)
	_, _ = w.Write(stub)
}

func (s *Server) handleServeWeb(w http.ResponseWriter, r *http.Request) {
	// If path is API or health, should have been handled by mux
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	webFS, err := fs.Sub(embeddedDist, "dist/web")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if file exists in webFS
	f, err := webFS.Open(path)
	if err == nil {
		_ = f.Close()
		ext := filepath.Ext(path)
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
		http.FileServer(http.FS(webFS)).ServeHTTP(w, r)
		return
	}

	// SPA fallback: serve index.html for client-side routing
	indexData, err := embeddedDist.ReadFile("dist/web/index.html")
	if err == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexData)
		return
	}

	http.NotFound(w, r)
}
