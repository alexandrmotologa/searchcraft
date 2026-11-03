package index

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func TestEngineHybridSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sc_engine_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "engine.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	emb := embedding.NewFastLocalEmbedder(256)
	engine, err := NewEngine(db, emb)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	ctx := context.Background()

	// 1. Create Index
	err = engine.CreateIndex("articles", nil, "hybrid", "fast_local")
	if err != nil {
		t.Fatalf("create index: %v", err)
	}

	// 2. Ingest Sample Documents
	docs := []storage.Document{
		{
			ID: "art_1",
			Data: map[string]any{
				"title":       "Getting Started with Go and SQLite",
				"description": "Learn how to use SQLite FTS5 for high speed full-text search in Go.",
				"category":    "Databases",
				"tags":        []string{"golang", "sqlite", "fts5"},
				"views":       1500,
			},
		},
		{
			ID: "art_2",
			Data: map[string]any{
				"title":       "Vector Embeddings and Reciprocal Rank Fusion",
				"description": "An introduction to cosine similarity and combining search algorithms.",
				"category":    "Search",
				"tags":        []string{"vector", "rrf", "ai"},
				"views":       3200,
			},
		},
		{
			ID: "art_3",
			Data: map[string]any{
				"title":       "Building High-Speed Web Applications with React 19",
				"description": "A deep dive into Server Components, client state, and responsive UIs.",
				"category":    "Frontend",
				"tags":        []string{"react", "javascript", "web"},
				"views":       2100,
			},
		},
	}

	indexed, err := engine.IndexDocuments(ctx, "articles", docs)
	if err != nil {
		t.Fatalf("index docs: %v", err)
	}
	if indexed != 3 {
		t.Errorf("expected 3 docs indexed, got %d", indexed)
	}

	// 3. Test BM25-only query
	resBM25, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query: "SQLite FTS5",
		Mode:  "bm25",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("search bm25: %v", err)
	}
	if len(resBM25.Hits) == 0 {
		t.Fatalf("expected bm25 hits, got 0")
	}
	if resBM25.Hits[0].ID != "art_1" {
		t.Errorf("expected top hit art_1, got %s", resBM25.Hits[0].ID)
	}

	// 4. Test Vector-only query
	resVector, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query: "cosine similarity search algorithms",
		Mode:  "vector",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("search vector: %v", err)
	}
	if len(resVector.Hits) == 0 {
		t.Fatalf("expected vector hits, got 0")
	}
	if resVector.Hits[0].ID != "art_2" {
		t.Errorf("expected top vector hit art_2, got %s", resVector.Hits[0].ID)
	}

	// 5. Test Hybrid RRF Search with Highlighting and Facets
	alpha := 0.5
	resHybrid, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query:       "React web components",
		Mode:        "hybrid",
		Limit:       10,
		HybridAlpha: &alpha,
		Highlight:   true,
		Facets:      []string{"category", "tags"},
	})
	if err != nil {
		t.Fatalf("search hybrid: %v", err)
	}
	if len(resHybrid.Hits) == 0 {
		t.Fatalf("expected hybrid hits, got 0")
	}
	if resHybrid.Hits[0].ID != "art_3" {
		t.Errorf("expected top hybrid hit art_3, got %s", resHybrid.Hits[0].ID)
	}
	if resHybrid.Hits[0].Highlights["title"] == "" {
		t.Errorf("expected highlight in title for art_3")
	}
	if resHybrid.Facets["category"]["Frontend"] != 1 {
		t.Errorf("expected category Frontend facet count 1, got %d", resHybrid.Facets["category"]["Frontend"])
	}

	// 6. Test Filter evaluation during search
	resFiltered, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query: "search",
		Mode:  "hybrid",
		Filters: map[string]any{
			"category": "Databases",
		},
	})
	if err != nil {
		t.Fatalf("search filtered: %v", err)
	}
	for _, hit := range resFiltered.Hits {
		if hit.Document["category"] != "Databases" {
			t.Errorf("expected only Databases category, got %v", hit.Document["category"])
		}
	}

	// 7. Test Synonyms Expansion
	err = db.AddSynonym(storage.SynonymRule{
		IndexName: "articles",
		Term:      "golang",
		Synonyms:  []string{"google-go"},
		IsTwoWay:  true,
	})
	if err != nil {
		t.Fatalf("add synonym: %v", err)
	}

	resSyn, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query: "google-go",
		Mode:  "bm25",
	})
	if err != nil || len(resSyn.Hits) == 0 {
		t.Fatalf("expected hits for expanded synonym 'google-go'")
	}
	if resSyn.Hits[0].ID != "art_1" {
		t.Errorf("expected art_1 via synonym, got %s", resSyn.Hits[0].ID)
	}

	// 8. Test Merchandising / Pinned Promotions
	err = db.AddPromotion(storage.PromotionRule{
		IndexName:    "articles",
		QueryPattern: "overview",
		PinnedDocIDs: []string{"art_3"},
		Position:     1,
	})
	if err != nil {
		t.Fatalf("add promotion: %v", err)
	}

	resPromo, err := engine.Search(ctx, "articles", storage.SearchRequest{
		Query: "system overview",
		Mode:  "hybrid",
	})
	if err != nil || len(resPromo.Hits) == 0 {
		t.Fatalf("expected hits with promotion")
	}
	if resPromo.Hits[0].ID != "art_3" {
		t.Errorf("expected pinned art_3 at position 1, got %s", resPromo.Hits[0].ID)
	}

	// 9. Test Autocomplete Suggestions
	suggestions, err := engine.Suggest(ctx, "articles", "re", 5)
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) == 0 {
		t.Errorf("expected suggestions for prefix 're', got 0")
	}
}
