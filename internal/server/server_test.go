package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
	"github.com/alexandrmotologa/searchcraft/internal/index"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	tmpDir, err := os.MkdirTemp("", "sc_server_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "server_test.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	emb := embedding.NewFastLocalEmbedder(256)
	eng, err := index.NewEngine(db, emb)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	srv := NewServer(Config{
		Port:    7700,
		Version: "1.0.0-test",
	}, eng, db)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return srv, cleanup
}

func TestServerAPI(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	handler := srv.Handler()

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected health 200, got %d", w.Code)
	}

	// 2. Create Index
	createPayload := []byte(`{
		"name": "products",
		"field_boosts": {"title": 3.0, "description": 1.0},
		"default_mode": "hybrid"
	}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/indexes", bytes.NewReader(createPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected create index 201, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Upsert Documents
	docPayload := []byte(`[
		{
			"id": "p1",
			"title": "Ultra-lightweight Ergonomic Mouse",
			"description": "High precision sensor with wireless connectivity",
			"category": "Mice",
			"price": 59.99
		},
		{
			"id": "p2",
			"title": "Mechanical Keyboard 75%",
			"description": "Hot-swappable switches with tactile feedback",
			"category": "Keyboards",
			"price": 109.99
		}
	]`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/indexes/products/documents", bytes.NewReader(docPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected upsert 200, got %d: %s", w.Code, w.Body.String())
	}

	// 4. List Documents
	req = httptest.NewRequest(http.MethodGet, "/api/v1/indexes/products/documents?limit=10", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected list docs 200, got %d", w.Code)
	}

	// 5. Search Documents
	searchPayload := []byte(`{
		"query": "ergonomic mouse",
		"mode": "hybrid",
		"limit": 5,
		"facets": ["category"]
	}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/indexes/products/search", bytes.NewReader(searchPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected search 200, got %d: %s", w.Code, w.Body.String())
	}

	var searchResp storage.SearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &searchResp); err != nil {
		t.Fatalf("unmarshal search resp: %v", err)
	}
	if len(searchResp.Hits) == 0 {
		t.Fatalf("expected search hits, got 0")
	}
	if searchResp.Hits[0].ID != "p1" {
		t.Errorf("expected top hit p1, got %s", searchResp.Hits[0].ID)
	}

	// 6. Analytics endpoints
	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected analytics overview 200, got %d", w.Code)
	}

	// 7. Widget script serving
	req = httptest.NewRequest(http.MethodGet, "/widget.js", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected widget.js 200, got %d", w.Code)
	}
}
