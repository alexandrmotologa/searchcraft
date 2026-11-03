package server

import (
	"encoding/json"
	"net/http"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

type createIndexRequest struct {
	Name              string             `json:"name"`
	FieldBoosts       map[string]float64 `json:"field_boosts"`
	DefaultMode       string             `json:"default_mode"`
	EmbeddingProvider string             `json:"embedding_provider"`
}

func (s *Server) handleListIndexes(w http.ResponseWriter, _ *http.Request) {
	indexes, err := s.engine.ListIndexes()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if indexes == nil {
		indexes = []storage.IndexMeta{}
	}
	writeJSON(w, http.StatusOK, indexes)
}

func (s *Server) handleCreateIndex(w http.ResponseWriter, r *http.Request) {
	var req createIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "index name is required")
		return
	}

	err := s.engine.CreateIndex(req.Name, req.FieldBoosts, req.DefaultMode, req.EmbeddingProvider)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Index created successfully",
		"name":    req.Name,
	})
}

func (s *Server) handleGetIndex(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	idx, err := s.engine.GetIndex(name)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, idx)
}

func (s *Server) handleDeleteIndex(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.engine.DeleteIndex(name); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Index deleted successfully",
		"name":    name,
	})
}
