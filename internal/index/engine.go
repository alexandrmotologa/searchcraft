package index

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
	"github.com/alexandrmotologa/searchcraft/internal/facets"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
	"github.com/alexandrmotologa/searchcraft/internal/typo"
)

// Engine orchestrates document indexing, full-text FTS5 search, vector retrieval, and RRF ranking.
type Engine struct {
	db          *storage.DB
	embedder    embedding.Embedder
	mu          sync.RWMutex
	vectorCaches map[string]*VectorIndex
}

// NewEngine initializes the search engine.
func NewEngine(db *storage.DB, emb embedding.Embedder) (*Engine, error) {
	if emb == nil {
		emb = embedding.NewFastLocalEmbedder(256)
	}

	eng := &Engine{
		db:           db,
		embedder:     emb,
		vectorCaches: make(map[string]*VectorIndex),
	}

	// Warm up vector caches for existing indexes
	indexes, err := db.ListIndexes()
	if err == nil {
		for _, idx := range indexes {
			_ = eng.warmupVectorCache(idx.Name)
		}
	}

	return eng, nil
}

// warmupVectorCache preloads stored vectors from SQLite into memory for fast cosine search.
func (e *Engine) warmupVectorCache(indexName string) error {
	vecs, err := e.db.LoadVectors(indexName)
	if err != nil {
		return err
	}

	vIndex := NewVectorIndex(e.embedder.Dimension())
	vIndex.SetBatch(vecs)

	e.mu.Lock()
	e.vectorCaches[indexName] = vIndex
	e.mu.Unlock()

	return nil
}

func (e *Engine) getVectorCache(indexName string) *VectorIndex {
	e.mu.RLock()
	cache, exists := e.vectorCaches[indexName]
	e.mu.RUnlock()

	if exists {
		return cache
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	// Double check
	if cache, exists = e.vectorCaches[indexName]; exists {
		return cache
	}

	cache = NewVectorIndex(e.embedder.Dimension())
	e.vectorCaches[indexName] = cache
	return cache
}

// CreateIndex registers a new index and prepares its vector cache.
func (e *Engine) CreateIndex(name string, fieldBoosts map[string]float64, defaultMode, provider string) error {
	if err := e.db.CreateIndex(name, fieldBoosts, defaultMode, provider); err != nil {
		return err
	}

	e.mu.Lock()
	e.vectorCaches[name] = NewVectorIndex(e.embedder.Dimension())
	e.mu.Unlock()

	return nil
}

// GetIndex returns index metadata.
func (e *Engine) GetIndex(name string) (*storage.IndexMeta, error) {
	return e.db.GetIndex(name)
}

// ListIndexes returns all active indexes.
func (e *Engine) ListIndexes() ([]storage.IndexMeta, error) {
	return e.db.ListIndexes()
}

// DeleteIndex drops the index and flushes its vector cache.
func (e *Engine) DeleteIndex(name string) error {
	if err := e.db.DeleteIndex(name); err != nil {
		return err
	}

	e.mu.Lock()
	delete(e.vectorCaches, name)
	e.mu.Unlock()

	return nil
}

// IndexDocuments batches document ingestion, computes dense embeddings, and saves to storage.
func (e *Engine) IndexDocuments(ctx context.Context, indexName string, docs []storage.Document) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	// Ensure IDs exist and prepare textual corpus for embedding
	textsToEmbed := make([]string, len(docs))
	for i := range docs {
		if docs[i].ID == "" {
			docs[i].ID = fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), i)
		}
		textsToEmbed[i] = buildEmbeddableText(docs[i].Data)
	}

	// Generate dense vectors
	embeddings, err := e.embedder.EmbedBatch(ctx, textsToEmbed)
	if err != nil {
		return 0, fmt.Errorf("embed batch: %w", err)
	}

	vectorMap := make(map[string][]float32, len(docs))
	for i, doc := range docs {
		if i < len(embeddings) && len(embeddings[i]) > 0 {
			vectorMap[doc.ID] = embeddings[i]
		}
	}

	// Save to SQLite
	if err := e.db.UpsertDocuments(indexName, docs); err != nil {
		return 0, fmt.Errorf("upsert sqlite docs: %w", err)
	}

	if err := e.db.SaveVectors(indexName, vectorMap); err != nil {
		return 0, fmt.Errorf("save vectors: %w", err)
	}

	// Update in-memory vector cache
	vCache := e.getVectorCache(indexName)
	vCache.SetBatch(vectorMap)

	return len(docs), nil
}

// GetDocument returns a single document.
func (e *Engine) GetDocument(indexName, docID string) (*storage.Document, error) {
	return e.db.GetDocument(indexName, docID)
}

// ListDocuments returns paginated documents.
func (e *Engine) ListDocuments(indexName string, limit, offset int) ([]storage.Document, int64, error) {
	return e.db.ListDocuments(indexName, limit, offset)
}

// DeleteDocument removes a document from storage and the vector cache.
func (e *Engine) DeleteDocument(indexName, docID string) error {
	if err := e.db.DeleteDocument(indexName, docID); err != nil {
		return err
	}

	vCache := e.getVectorCache(indexName)
	vCache.Delete(docID)

	return nil
}

// Search executes a hybrid search query combining BM25 and vector similarity with RRF.
func (e *Engine) Search(ctx context.Context, indexName string, req storage.SearchRequest) (*storage.SearchResponse, error) {
	startTime := time.Now()

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	alpha := 0.5
	if req.HybridAlpha != nil {
		alpha = *req.HybridAlpha
	}

	mode := strings.ToLower(req.Mode)
	if mode == "" {
		mode = "hybrid"
	}

	cleanQuery := typo.CleanQuery(req.Query)
	queryTokens := typo.Tokenize(cleanQuery)

	// Fetch candidate multiplier to allow post-filtering by structured facets
	candidateLimit := limit * 4
	if candidateLimit < 50 {
		candidateLimit = 50
	}

	var bm25Hits []storage.FTSResult
	var vectorHits []VectorMatch

	// 1. BM25 Search
	if mode == "hybrid" || mode == "bm25" {
		if cleanQuery != "" {
			ftsResults, _ := ExecuteBM25(e.db, indexName, cleanQuery, candidateLimit)
			bm25Hits = ftsResults
		}
	}

	// 2. Vector Search
	if mode == "hybrid" || mode == "vector" {
		if cleanQuery != "" {
			queryVec, err := e.embedder.Embed(ctx, cleanQuery)
			if err == nil && len(queryVec) > 0 {
				vCache := e.getVectorCache(indexName)
				vectorHits = vCache.Search(queryVec, candidateLimit)
			}
		}
	}

	// 3. Combine rankings via RRF and Hybrid linear scoring
	var candidates []CandidateScore
	switch mode {
	case "bm25":
		candidates = CombineRRF(bm25Hits, nil, 1.0, DefaultRRFK)
	case "vector":
		candidates = CombineRRF(nil, vectorHits, 0.0, DefaultRRFK)
	default:
		candidates = CombineRRF(bm25Hits, vectorHits, alpha, DefaultRRFK)
	}

	// 4. Load full documents, evaluate filters, compute facets, and build final hits
	var matchedDocs []map[string]any
	var hits []storage.SearchHit

	for _, cand := range candidates {
		doc, err := e.db.GetDocument(indexName, cand.DocID)
		if err != nil || doc == nil {
			continue
		}

		// Apply structured filters if provided
		if len(req.Filters) > 0 && !facets.MatchesFilter(doc.Data, req.Filters) {
			continue
		}

		matchedDocs = append(matchedDocs, doc.Data)

		var highlights map[string]string
		if req.Highlight && len(queryTokens) > 0 {
			highlights = GenerateHighlights(doc.Data, queryTokens)
		}

		hits = append(hits, storage.SearchHit{
			ID:          cand.DocID,
			Score:       cand.FinalScore,
			RRFScore:    cand.RRFScore,
			BM25Score:   cand.BM25Score,
			VectorScore: cand.VectorScore,
			Document:    doc.Data,
			Highlights:  highlights,
		})
	}

	// 5. Dynamic Facet Aggregation across matching documents
	var facetResults map[string]map[string]int
	if len(req.Facets) > 0 {
		facetResults = facets.AggregateFacets(matchedDocs, req.Facets)
	}

	total := len(hits)

	// Apply pagination (offset & limit)
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > len(hits) {
		hits = []storage.SearchHit{}
	} else {
		end := offset + limit
		if end > len(hits) {
			end = len(hits)
		}
		hits = hits[offset:end]
	}

	duration := time.Since(startTime)
	tookMs := float64(duration.Microseconds()) / 1000.0

	// 6. Record query analytics asynchronously
	go e.db.RecordSearch(storage.SearchLog{
		IndexName: indexName,
		Query:     req.Query,
		Mode:      mode,
		TotalHits: total,
		TookMs:    tookMs,
	})

	return &storage.SearchResponse{
		Hits:   hits,
		Total:  total,
		TookMs: tookMs,
		Query:  req.Query,
		Mode:   mode,
		Facets: facetResults,
	}, nil
}

// buildEmbeddableText concatenates representative textual attributes for embedding vector generation.
func buildEmbeddableText(data map[string]any) string {
	var parts []string
	keys := []string{"title", "name", "description", "summary", "category", "brand", "tags", "content"}

	for _, k := range keys {
		val, exists := data[k]
		if !exists || val == nil {
			continue
		}
		switch v := val.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				parts = append(parts, v)
			}
		case []string:
			parts = append(parts, strings.Join(v, " "))
		case []any:
			var strItems []string
			for _, item := range v {
				strItems = append(strItems, fmt.Sprintf("%v", item))
			}
			parts = append(parts, strings.Join(strItems, " "))
		}
	}

	return strings.Join(parts, " ")
}
