package server

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("GET /health", s.handleHealth)

	// Indexes
	mux.HandleFunc("GET /api/v1/indexes", s.handleListIndexes)
	mux.HandleFunc("POST /api/v1/indexes", s.handleCreateIndex)
	mux.HandleFunc("GET /api/v1/indexes/{name}", s.handleGetIndex)
	mux.HandleFunc("DELETE /api/v1/indexes/{name}", s.handleDeleteIndex)

	// Documents
	mux.HandleFunc("POST /api/v1/indexes/{name}/documents", s.handleUpsertDocuments)
	mux.HandleFunc("GET /api/v1/indexes/{name}/documents", s.handleListDocuments)
	mux.HandleFunc("GET /api/v1/indexes/{name}/documents/{id}", s.handleGetDocument)
	mux.HandleFunc("DELETE /api/v1/indexes/{name}/documents/{id}", s.handleDeleteDocument)

	// Search
	mux.HandleFunc("POST /api/v1/indexes/{name}/search", s.handleSearch)

	// Analytics
	mux.HandleFunc("GET /api/v1/analytics/overview", s.handleAnalyticsOverview)
	mux.HandleFunc("GET /api/v1/analytics/popular", s.handlePopularQueries)
	mux.HandleFunc("GET /api/v1/analytics/zero-results", s.handleZeroResults)

	// Widget & Static Admin UI
	mux.HandleFunc("GET /widget.js", s.handleServeWidget)
	mux.HandleFunc("GET /", s.handleServeWeb)

	return s.corsMiddleware(s.authMiddleware(mux))
}

// corsMiddleware allows cross-origin requests from any client website embedding the search widget.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware enforces bearer token validation if configured.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AuthToken == "" || r.URL.Path == "/health" || r.URL.Path == "/widget.js" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Allow search and static assets without auth, protect mutations
		if r.Method == http.MethodGet && !containsString(r.URL.Path, "/analytics/") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+s.cfg.AuthToken {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized: invalid or missing bearer token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	indexes, _ := s.engine.ListIndexes()
	var totalDocs int64
	for _, idx := range indexes {
		totalDocs += idx.DocumentCount
	}

	res := map[string]any{
		"status":          "ok",
		"version":         s.cfg.Version,
		"uptime_seconds":  int64(time.Since(s.startTime).Seconds()),
		"memory_bytes":    m.Alloc,
		"total_indexes":   len(indexes),
		"total_documents": totalDocs,
	}

	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) > 0 && indexOf(str, substr) >= 0)
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
