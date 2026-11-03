package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func (s *Server) handleUpsertDocuments(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	if indexName == "" {
		writeJSONError(w, http.StatusBadRequest, "index name is required")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "read request body: "+err.Error())
		return
	}

	var rawItems []map[string]any
	// Try parsing as array
	if err := json.Unmarshal(bodyBytes, &rawItems); err != nil {
		// Try parsing as single object
		var single map[string]any
		if errSingle := json.Unmarshal(bodyBytes, &single); errSingle != nil {
			writeJSONError(w, http.StatusBadRequest, "expected JSON array or object of documents")
			return
		}
		rawItems = []map[string]any{single}
	}

	if len(rawItems) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"indexed": 0, "took_ms": 0.0})
		return
	}

	// Ensure the index exists, auto-create if not present
	if _, err := s.engine.GetIndex(indexName); err != nil {
		_ = s.engine.CreateIndex(indexName, nil, "hybrid", "fast_local")
	}

	docs := make([]storage.Document, len(rawItems))
	for i, item := range rawItems {
		docID := extractDocID(item)
		if docID == "" {
			docID = fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), i)
			item["id"] = docID
		}
		docs[i] = storage.Document{
			ID:   docID,
			Data: item,
		}
	}

	start := time.Now()
	indexedCount, err := s.engine.IndexDocuments(r.Context(), indexName, docs)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "index documents: "+err.Error())
		return
	}

	tookMs := float64(time.Since(start).Microseconds()) / 1000.0
	writeJSON(w, http.StatusOK, map[string]any{
		"indexed": indexedCount,
		"failed":  0,
		"took_ms": tookMs,
	})
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	limit := 20
	offset := 0

	qLimit := r.URL.Query().Get("limit")
	if qLimit != "" {
		if l, err := strconv.Atoi(qLimit); err == nil && l > 0 {
			limit = l
		}
	}

	qOffset := r.URL.Query().Get("offset")
	if qOffset != "" {
		if o, err := strconv.Atoi(qOffset); err == nil && o >= 0 {
			offset = o
		}
	}

	docs, total, err := s.engine.ListDocuments(indexName, limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"documents": docs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	docID := r.PathValue("id")

	doc, err := s.engine.GetDocument(indexName, docID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	docID := r.PathValue("id")

	if err := s.engine.DeleteDocument(indexName, docID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Document deleted successfully",
		"id":      docID,
	})
}

func extractDocID(data map[string]any) string {
	keys := []string{"id", "doc_id", "key", "_id"}
	for _, k := range keys {
		if val, exists := data[k]; exists && val != nil {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}
