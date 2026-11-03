package server

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
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

	// Search & Suggest
	mux.HandleFunc("POST /api/v1/indexes/{name}/search", s.handleSearch)
	mux.HandleFunc("GET /api/v1/indexes/{name}/suggest", s.handleSuggest)

	// Synonyms
	mux.HandleFunc("GET /api/v1/indexes/{name}/synonyms", s.handleListSynonyms)
	mux.HandleFunc("POST /api/v1/indexes/{name}/synonyms", s.handleAddSynonym)
	mux.HandleFunc("DELETE /api/v1/indexes/{name}/synonyms/{term}", s.handleDeleteSynonym)

	// Promotions (Merchandising)
	mux.HandleFunc("GET /api/v1/indexes/{name}/promotions", s.handleListPromotions)
	mux.HandleFunc("POST /api/v1/indexes/{name}/promotions", s.handleAddPromotion)
	mux.HandleFunc("DELETE /api/v1/indexes/{name}/promotions/{query}", s.handleDeletePromotion)

	// Web Crawler
	mux.HandleFunc("POST /api/v1/indexes/{name}/crawl", s.handleCrawl)

	// Data Portability
	mux.HandleFunc("GET /api/v1/indexes/{name}/export", s.handleExportDocuments)
	mux.HandleFunc("POST /api/v1/system/backup", s.handleSystemBackup)

	// API Keys Management
	mux.HandleFunc("GET /api/v1/keys", s.handleListAPIKeys)
	mux.HandleFunc("POST /api/v1/keys", s.handleCreateAPIKey)
	mux.HandleFunc("DELETE /api/v1/keys/{prefix}", s.handleDeleteAPIKey)

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

// authMiddleware enforces role-based access control (Search-Only vs Admin).
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Public assets and health checks are always open
		if path == "/health" || path == "/widget.js" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Static UI assets
		if !strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if authentication is enabled (either master token or stored API keys)
		hasMasterToken := s.cfg.AuthToken != ""
		keyCount, _ := s.db.CountAPIKeys()
		hasRegisteredKeys := keyCount > 0

		// Open development mode if no keys or tokens configured
		if !hasMasterToken && !hasRegisteredKeys {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Header or Query Param
		rawToken := ""
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			rawToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else if qToken := r.URL.Query().Get("api_key"); qToken != "" {
			rawToken = qToken
		}

		if rawToken == "" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized: missing api_key or bearer token")
			return
		}

		// Check master admin token
		if hasMasterToken && rawToken == s.cfg.AuthToken {
			next.ServeHTTP(w, r)
			return
		}

		// Validate against stored keys
		keyInfo, err := s.db.ValidateAPIKey(rawToken)
		if err != nil || keyInfo == nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized: invalid api key")
			return
		}

		// Role enforcement:
		if keyInfo.Role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		if keyInfo.Role == "search" {
			// Search key only permits search & suggest operations
			isSearchOp := strings.HasSuffix(path, "/search") && r.Method == http.MethodPost
			isSuggestOp := strings.HasSuffix(path, "/suggest") && r.Method == http.MethodGet

			if isSearchOp || isSuggestOp {
				next.ServeHTTP(w, r)
				return
			}

			writeJSONError(w, http.StatusForbidden, "forbidden: search key cannot perform admin operations")
			return
		}

		writeJSONError(w, http.StatusUnauthorized, "unauthorized: unrecognized role")
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
