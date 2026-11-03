package server

import (
	"encoding/json"
	"net/http"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	indexName := r.PathValue("name")
	if indexName == "" {
		writeJSONError(w, http.StatusBadRequest, "index name is required")
		return
	}

	var req storage.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid search request json: "+err.Error())
		return
	}

	resp, err := s.engine.Search(r.Context(), indexName, req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
