package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/crawler"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

// --- API Keys Handlers ---

func (s *Server) handleListAPIKeys(w http.ResponseWriter, _ *http.Request) {
	keys, err := s.db.ListAPIKeys()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if keys == nil {
		keys = []storage.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	rawKey, keyMeta, err := s.db.CreateAPIKey(req.Name, req.Role)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"key":        rawKey, // returned only once at creation!
		"prefix":     keyMeta.Prefix,
		"name":       keyMeta.Name,
		"role":       keyMeta.Role,
		"created_at": keyMeta.CreatedAt,
	})
}

func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	prefix := r.PathValue("prefix")
	if err := s.db.DeleteAPIKey(prefix); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "API key deleted"})
}

// --- Synonyms Handlers ---

func (s *Server) handleListSynonyms(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	synonyms, err := s.db.ListSynonyms(indexName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if synonyms == nil {
		synonyms = []storage.SynonymRule{}
	}
	writeJSON(w, http.StatusOK, synonyms)
}

func (s *Server) handleAddSynonym(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	var rule storage.SynonymRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	rule.IndexName = indexName

	if err := s.db.AddSynonym(rule); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Synonym added successfully"})
}

func (s *Server) handleDeleteSynonym(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	term := r.PathValue("term")

	if err := s.db.DeleteSynonym(indexName, term); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Synonym deleted"})
}

// --- Promotions Handlers ---

func (s *Server) handleListPromotions(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	promos, err := s.db.ListPromotions(indexName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if promos == nil {
		promos = []storage.PromotionRule{}
	}
	writeJSON(w, http.StatusOK, promos)
}

func (s *Server) handleAddPromotion(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	var rule storage.PromotionRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	rule.IndexName = indexName

	if err := s.db.AddPromotion(rule); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Promotion rule added successfully"})
}

func (s *Server) handleDeletePromotion(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	queryPattern := r.PathValue("query")

	if err := s.db.DeletePromotion(indexName, queryPattern); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Promotion rule deleted"})
}

// --- Suggestions / Autocomplete Handler ---

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	q := r.URL.Query().Get("q")
	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	suggestions, err := s.engine.Suggest(r.Context(), indexName, q, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if suggestions == nil {
		suggestions = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query":       q,
		"suggestions": suggestions,
	})
}

// --- Web Crawler Handler ---

func (s *Server) handleCrawl(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	var req storage.CrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	// Auto create index if it doesn't exist
	if _, err := s.engine.GetIndex(indexName); err != nil {
		_ = s.engine.CreateIndex(indexName, nil, "hybrid", "fast_local")
	}

	c := crawler.NewCrawler()
	result, err := c.Crawl(r.Context(), s.engine, indexName, req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "crawl failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// --- Data Portability: Export & Backup Handlers ---

func (s *Server) handleExportDocuments(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	docs, err := s.db.ExportAllDocuments(indexName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_export.json", indexName))
	writeJSON(w, http.StatusOK, docs)
}

func (s *Server) handleSystemBackup(w http.ResponseWriter, r *http.Request) {
	target := fmt.Sprintf("./backup_%d.db", time.Now().Unix())
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Path != "" {
		target = req.Path
	}

	if err := s.db.BackupTo(target); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "backup failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Backup created successfully",
		"path":    target,
	})
}
