package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sc_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// 1. Create Index
	err = db.CreateIndex("products", map[string]float64{
		"title":       3.0,
		"description": 1.0,
	}, "hybrid", "fast_local")
	if err != nil {
		t.Fatalf("create index: %v", err)
	}

	meta, err := db.GetIndex("products")
	if err != nil {
		t.Fatalf("get index: %v", err)
	}
	if meta.Name != "products" {
		t.Errorf("expected name products, got %s", meta.Name)
	}

	// 2. Upsert Documents
	docs := []Document{
		{
			ID: "doc_1",
			Data: map[string]any{
				"title":       "Ergonomic Wireless Mouse",
				"description": "Optical mouse with bluetooth and long battery life",
				"category":    "Mice",
				"price":       49.99,
				"tags":        []string{"wireless", "bluetooth"},
			},
		},
		{
			ID: "doc_2",
			Data: map[string]any{
				"title":       "Mechanical Keyboard RGB",
				"description": "Tactile switches with custom keycaps and backlight",
				"category":    "Keyboards",
				"price":       119.99,
				"tags":        []string{"mechanical", "rgb"},
			},
		},
	}

	err = db.UpsertDocuments("products", docs)
	if err != nil {
		t.Fatalf("upsert docs: %v", err)
	}

	// 3. Verify Document retrieval & count
	meta, err = db.GetIndex("products")
	if err != nil {
		t.Fatalf("get index post-upsert: %v", err)
	}
	if meta.DocumentCount != 2 {
		t.Errorf("expected 2 documents, got %d", meta.DocumentCount)
	}

	doc1, err := db.GetDocument("products", "doc_1")
	if err != nil {
		t.Fatalf("get doc_1: %v", err)
	}
	if doc1.Data["title"] != "Ergonomic Wireless Mouse" {
		t.Errorf("expected title Ergonomic Wireless Mouse, got %v", doc1.Data["title"])
	}

	// 4. Test FTS5 BM25 Query
	ftsRes, err := db.QueryFTS("products", "mouse", 10)
	if err != nil {
		t.Fatalf("query fts: %v", err)
	}
	if len(ftsRes) == 0 {
		t.Fatalf("expected fts hit for 'mouse', got 0")
	}
	if ftsRes[0].DocID != "doc_1" {
		t.Errorf("expected top hit doc_1, got %s", ftsRes[0].DocID)
	}

	// 5. Test Vector persistence
	vecs := map[string][]float32{
		"doc_1": {0.1, 0.2, 0.3, 0.4},
		"doc_2": {0.9, 0.8, 0.7, 0.6},
	}
	err = db.SaveVectors("products", vecs)
	if err != nil {
		t.Fatalf("save vectors: %v", err)
	}

	loadedVecs, err := db.LoadVectors("products")
	if err != nil {
		t.Fatalf("load vectors: %v", err)
	}
	if len(loadedVecs) != 2 {
		t.Errorf("expected 2 loaded vectors, got %d", len(loadedVecs))
	}
	if len(loadedVecs["doc_1"]) != 4 || loadedVecs["doc_1"][0] != 0.1 {
		t.Errorf("vector mismatch for doc_1: %v", loadedVecs["doc_1"])
	}

	// 6. Test Analytics & Zero Results logging
	db.RecordSearch(SearchLog{
		IndexName: "products",
		Query:     "mouse",
		Mode:      "hybrid",
		TotalHits: 1,
		TookMs:    2.4,
	})
	db.RecordSearch(SearchLog{
		IndexName: "products",
		Query:     "unmatched keyword",
		Mode:      "hybrid",
		TotalHits: 0,
		TookMs:    1.2,
	})

	overview, err := db.GetAnalyticsOverview()
	if err != nil {
		t.Fatalf("get analytics: %v", err)
	}
	if overview.TotalSearches != 2 {
		t.Errorf("expected 2 total searches, got %d", overview.TotalSearches)
	}
	if overview.ZeroResultSearches != 1 {
		t.Errorf("expected 1 zero-result search, got %d", overview.ZeroResultSearches)
	}

	zeros, err := db.GetZeroResults(10)
	if err != nil {
		t.Fatalf("get zero results: %v", err)
	}
	if len(zeros) != 1 || zeros[0].Query != "unmatched keyword" {
		t.Errorf("unexpected zero results: %v", zeros)
	}

	popular, err := db.GetPopularQueries(10)
	if err != nil {
		t.Fatalf("get popular: %v", err)
	}
	if len(popular) < 2 {
		t.Errorf("expected 2 popular items, got %d", len(popular))
	}

	// 7. Delete document
	err = db.DeleteDocument("products", "doc_1")
	if err != nil {
		t.Fatalf("delete doc_1: %v", err)
	}
	_, err = db.GetDocument("products", "doc_1")
	if err == nil {
		t.Errorf("expected error getting deleted doc_1, got nil")
	}

	// 8. Delete index
	err = db.DeleteIndex("products")
	if err != nil {
		t.Fatalf("delete index: %v", err)
	}
	_, err = db.GetIndex("products")
	if err == nil {
		t.Errorf("expected error getting deleted index, got nil")
	}
}
