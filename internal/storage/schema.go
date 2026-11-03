package storage

import "time"

// IndexMeta represents configuration and statistics for a search index.
type IndexMeta struct {
	Name              string             `json:"name"`
	CreatedAt         time.Time          `json:"created_at"`
	DocumentCount     int64              `json:"document_count"`
	FieldBoosts       map[string]float64 `json:"field_boosts"`
	DefaultMode       string             `json:"default_mode"`
	EmbeddingProvider string             `json:"embedding_provider"`
}

// Document represents an indexed JSON record.
type Document struct {
	ID        string         `json:"id"`
	Index     string         `json:"index,omitempty"`
	Data      map[string]any `json:"data"`
	RawJSON   string         `json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SearchRequest holds parameters for executing a search query.
type SearchRequest struct {
	Query       string         `json:"query"`
	Mode        string         `json:"mode,omitempty"`         // "hybrid", "bm25", "vector"
	Limit       int            `json:"limit,omitempty"`        // default 20
	Offset      int            `json:"offset,omitempty"`       // default 0
	HybridAlpha *float64       `json:"hybrid_alpha,omitempty"` // 0.0 (vector) to 1.0 (bm25), default 0.5
	Highlight   bool           `json:"highlight,omitempty"`
	Facets      []string       `json:"facets,omitempty"`
	Filters     map[string]any `json:"filters,omitempty"`
}

// SearchHit represents a single matched document in search results.
type SearchHit struct {
	ID          string            `json:"id"`
	Score       float64           `json:"score"`
	RRFScore    float64           `json:"rrf_score"`
	BM25Score   float64           `json:"bm25_score"`
	VectorScore float64           `json:"vector_score"`
	Document    map[string]any    `json:"document"`
	Highlights  map[string]string `json:"highlights,omitempty"`
}

// SearchResponse is returned to search clients.
type SearchResponse struct {
	Hits        []SearchHit               `json:"hits"`
	Total       int                       `json:"total"`
	TookMs      float64                   `json:"took_ms"`
	Query       string                    `json:"query"`
	Mode        string                    `json:"mode"`
	Facets      map[string]map[string]int `json:"facets,omitempty"`
	DidYouMean  string                    `json:"did_you_mean,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
}

// SearchLog records individual query executions for latency and volume analytics.
type SearchLog struct {
	ID        int64     `json:"id"`
	IndexName string    `json:"index_name"`
	Query     string    `json:"query"`
	Mode      string    `json:"mode"`
	TotalHits int       `json:"total_hits"`
	TookMs    float64   `json:"took_ms"`
	CreatedAt time.Time `json:"created_at"`
}

// ZeroResultQuery records queries that produced 0 hits.
type ZeroResultQuery struct {
	Query        string    `json:"query"`
	Count        int64     `json:"count"`
	LastSearched time.Time `json:"last_searched"`
}

// AnalyticsOverview provides high-level search engine performance stats.
type AnalyticsOverview struct {
	TotalSearches      int64   `json:"total_searches"`
	ZeroResultSearches int64   `json:"zero_result_searches"`
	ZeroResultRate     float64 `json:"zero_result_rate"`
	AvgLatencyMs       float64 `json:"avg_latency_ms"`
	P50LatencyMs       float64 `json:"p50_latency_ms"`
	P95LatencyMs       float64 `json:"p95_latency_ms"`
	P99LatencyMs       float64 `json:"p99_latency_ms"`
}

// PopularQuery represents a frequently searched query with average hit counts.
type PopularQuery struct {
	Query   string  `json:"query"`
	Count   int64   `json:"count"`
	AvgHits float64 `json:"avg_hits"`
}

// APIKey represents an authentication key for SearchCraft.
type APIKey struct {
	Prefix    string    `json:"prefix"`
	Name      string    `json:"name"`
	Role      string    `json:"role"` // "admin" or "search"
	CreatedAt time.Time `json:"created_at"`
}

// SynonymRule defines synonym mappings for query expansion.
type SynonymRule struct {
	IndexName string   `json:"index_name"`
	Term      string   `json:"term"`
	Synonyms  []string `json:"synonyms"`
	IsTwoWay  bool     `json:"is_two_way"`
}

// PromotionRule pins specific documents to top ranks for designated queries.
type PromotionRule struct {
	IndexName    string   `json:"index_name"`
	QueryPattern string   `json:"query_pattern"`
	PinnedDocIDs []string `json:"pinned_doc_ids"`
	Position     int      `json:"position"` // 1-based start position
}

// CrawlRequest parameters for the integrated web crawler.
type CrawlRequest struct {
	URL      string `json:"url"`
	MaxPages int    `json:"max_pages,omitempty"` // default 50, max 200
}

// CrawlResult summary of a website crawl run.
type CrawlResult struct {
	URL          string  `json:"url"`
	PagesCrawled int     `json:"pages_crawled"`
	DocsIndexed  int     `json:"docs_indexed"`
	TookMs       float64 `json:"took_ms"`
}

