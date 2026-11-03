package server

import (
	"net/http"
	"strconv"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func (s *Server) handleAnalyticsOverview(w http.ResponseWriter, _ *http.Request) {
	overview, err := s.db.GetAnalyticsOverview()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handlePopularQueries(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if qLimit := r.URL.Query().Get("limit"); qLimit != "" {
		if l, err := strconv.Atoi(qLimit); err == nil && l > 0 {
			limit = l
		}
	}

	popular, err := s.db.GetPopularQueries(limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if popular == nil {
		popular = []storage.PopularQuery{}
	}
	writeJSON(w, http.StatusOK, popular)
}

func (s *Server) handleZeroResults(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if qLimit := r.URL.Query().Get("limit"); qLimit != "" {
		if l, err := strconv.Atoi(qLimit); err == nil && l > 0 {
			limit = l
		}
	}

	zeros, err := s.db.GetZeroResults(limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if zeros == nil {
		zeros = []storage.ZeroResultQuery{}
	}
	writeJSON(w, http.StatusOK, zeros)
}
