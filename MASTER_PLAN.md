# SearchCraft: Master Implementation Plan

High-performance single-binary instant search engine with BM25, dense vector hybrid search, typo-tolerance, dynamic faceting, and a drop-in 2KB command palette search widget.

---

## 1. Executive Summary & Problem Space

Modern web applications, documentation hubs, e-commerce stores, and SaaS platforms require instant search with keyboard navigation (Cmd+K). Existing solutions introduce operational challenges:
- Algolia charges per search operation and per record, causing costs to increase rapidly with scale.
- Elasticsearch and OpenSearch rely on the JVM, typically requiring several gigabytes of RAM and complex cluster maintenance.
- Meilisearch provides fast keyword matching, but hybrid vector search configurations require external embedding pipelines and higher memory footprints.
- Custom search implementations often take weeks to build, requiring manual handling of debouncing, keyboard focus, highlighting, typo handling, and facet aggregation.

SearchCraft provides a self-hosted search engine in a single Go binary with an embedded SQLite FTS5 engine and dense vector support:
- Memory footprint under 40MB RAM under normal workloads.
- Sub-10ms query execution across 500,000 documents.
- Hybrid search combining BM25 keyword matching and dense vector similarity through Reciprocal Rank Fusion (RRF).
- Adaptive typo-tolerance via Damerau-Levenshtein distance calculation.
- Dynamic facet aggregation and range filtering.
- Standalone 2KB Cmd+K search widget compiled with zero runtime dependencies.
- React 19 administration studio with a live query sandbox, document management, and latency analytics.

---

## 2. Technical Stack

- Backend: Go 1.23+ (`net/http` high-throughput routing, zero-dependency server architecture)
- Storage & Full-Text Search: Pure Go SQLite (`modernc.org/sqlite`) with FTS5, BM25 scoring, and WAL journaling
- Vector Search: Vector storage with cosine similarity, dot product calculations, and top-K ranking
- Embeddings:
  - Built-in deterministic character n-gram semantic vectorizer (128-256 dimensions, zero external downloads)
  - Local ONNX Runtime provider for `all-MiniLM-L6-v2` (384 dimensions)
  - External embedding provider support (OpenAI `text-embedding-3-small`, Ollama `nomic-embed-text`, Voyage AI, custom HTTP endpoints)
- Typo Tolerance: Damerau-Levenshtein automaton with adaptive thresholds based on token length
- Search Widget: Vanilla TypeScript compiled via esbuild to a minified standalone bundle under 3KB
- Admin Studio: React 19, TypeScript, Vite, Tailwind CSS, Lucide icons, embedded directly inside the Go binary via `embed.FS`

---

## 3. System Architecture

```
+-------------------------------------------------------------------------------+
|                               SEARCHCRAFT ENGINE                              |
+---------------------------------------+---------------------------------------+
                                        |
                +-----------------------+-----------------------+
                |                                               |
+---------------v---------------+               +---------------v---------------+
|     REST API & Ingestion      |               |     Hybrid Query Processor    |
| * POST /indexes/{id}/docs     |               | * BM25 Token Matching         |
| * Schema Inference            |               | * Typo Tolerance Engine       |
| * Batch Bulk Upsert           |               | * Dense Vector Similarity     |
| * Document Deletion           |               | * Reciprocal Rank Fusion      |
+---------------+---------------+               +---------------+---------------+
                |                                               |
                +-----------------------+-----------------------+
                                        |
+---------------------------------------v---------------------------------------+
|                              STORAGE LAYER                                    |
| * SQLite Engine (WAL mode, normal sync, memory cache)                         |
| * Documents Table: JSON payload and metadata                                  |
| * FTS5 Table: Porter stemmer, trigram tokenizer, column weights               |
| * Vectors Table: Packed float32 BLOB vectors with cosine index                |
| * Analytics Table: Query latencies, volume, zero-result queries               |
+---------------------------------------+---------------------------------------+
                                        |
                +-----------------------+-----------------------+
                |                                               |
+---------------v---------------+               +---------------v---------------+
|   2KB Standalone Search Widget|               |    React 19 Admin Studio      |
| * Cmd+K / Ctrl+K Modal        |               | * Live Search Sandbox         |
| * Zero External Dependencies  |               | * Document & Index Explorer   |
| * Keyboard Navigation         |               | * Analytics & Zero-Results    |
| * Highlight Match Offsets     |               | * Widget Embed Code Generator |
+-------------------------------+---------------+-------------------------------+
```

---

## 4. Directory Structure

```
searchcraft/
├── cmd/
│   └── searchcraft/
│       └── main.go                 # CLI entrypoint (serve, import, benchmark)
├── internal/
│   ├── config/
│   │   └── config.go               # Server and engine configuration flags
│   ├── storage/
│   │   ├── sqlite.go               # SQLite connection, pragmas, migrations
│   │   └── schema.go               # Document, Index, and Analytics schemas
│   ├── index/
│   │   ├── engine.go               # Orchestrator coordinating FTS5, Vectors, and RRF
│   │   ├── bm25.go                 # FTS5 BM25 querying, field weights, highlighting
│   │   ├── vector.go               # Cosine similarity vector search and top-K selection
│   │   └── rrf.go                  # Reciprocal Rank Fusion and hybrid score combiner
│   ├── typo/
│   │   ├── levenshtein.go          # Damerau-Levenshtein distance calculation
│   │   └── fuzzy.go                # Query token expansion and prefix candidate matching
│   ├── facets/
│   │   ├── aggregator.go           # Dynamic categorical facet counter
│   │   └── filter.go               # Equality, range, and array filter evaluator
│   ├── embedding/
│   │   ├── embedder.go             # Common Embedder interface
│   │   ├── fast_local.go           # Fast built-in deterministic n-gram vectorizer
│   │   ├── onnx_local.go           # ONNX Runtime MiniLM loader
│   │   └── client.go               # Remote API embedding client (OpenAI, Ollama)
│   ├── analytics/
│   │   └── tracker.go              # Query logs, latency percentiles, zero-result alerts
│   └── server/
│       ├── server.go               # HTTP server setup, middleware, graceful shutdown
│       ├── routes.go               # Route registration
│       ├── handlers_index.go       # Index management handlers
│       ├── handlers_doc.go         # Document CRUD and bulk ingestion
│       ├── handlers_search.go      # Search execution handler
│       ├── handlers_analytics.go   # Analytics queries handler
│       └── static.go               # Static file server with embed.FS
├── widget/
│   ├── src/
│   │   ├── index.ts                # Standalone Cmd+K search widget script
│   │   └── styles.css              # Minified scoped CSS styles
│   ├── package.json
│   └── esbuild.config.mjs          # Standalone bundle compiler
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Navbar.tsx
│   │   │   ├── SearchSandbox.tsx   # Live search playground with score breakdown
│   │   │   ├── DocumentManager.tsx # Document table, JSON editor, bulk importer
│   │   │   ├── IndexManager.tsx    # Index configuration and field weight boosting
│   │   │   ├── QueryAnalytics.tsx  # Latency charts, top queries, zero results
│   │   │   └── WidgetEmbed.tsx     # Embed code generator with interactive preview
│   │   ├── types/
│   │   │   └── search.ts
│   │   ├── api/
│   │   │   └── client.ts
│   │   ├── App.tsx
│   │   ├── index.css
│   │   └── main.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── docs/
│   ├── API_SPEC.md                 # Complete REST API reference
│   ├── HYBRID_SEARCH.md            # BM25 + Vector RRF algorithm deep dive
│   ├── WIDGET_INTEGRATION.md       # Integration guide for websites and SaaS apps
│   └── BENCHMARKS.md               # Latency and memory benchmarks
├── scripts/
│   └── seed_data.json              # Sample datasets for instant evaluation
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

---

## 5. Detailed Component Specifications

### 5.1 Storage Layer & FTS5 Configuration
- Driver: `modernc.org/sqlite` (Pure Go, FTS5 enabled, cross-platform without CGO)
- Pragmas:
  - `PRAGMA journal_mode = WAL;`
  - `PRAGMA synchronous = NORMAL;`
  - `PRAGMA temp_store = MEMORY;`
  - `PRAGMA cache_size = -64000;` (approx 64MB cache)
  - `PRAGMA busy_timeout = 5000;`
- Tables:
  - `sc_indexes`: Stores index configurations, field boost weights, default search mode, created timestamps.
  - `sc_documents`: Stores document ID, index name, raw JSON content, and timestamp metadata.
  - `sc_fts_<index>`: SQLite FTS5 virtual table indexing title, content, tags, and categorical text using porter stemmer and unicode61/trigram tokenizers.
  - `sc_vectors`: Stores document ID, index name, vector dimension, and packed IEEE 754 float32 byte slice.
  - `sc_search_logs`: Records query text, execution duration, total hits returned, search mode, timestamp.
  - `sc_zero_results`: Maintains aggregated counters of queries returning zero results with last seen timestamp.

### 5.2 Hybrid Search & Reciprocal Rank Fusion
1. BM25 Search:
   - Queries FTS5 table using column weights (e.g. `bm25(table, 3.0, 1.0, 1.5)`).
   - Extracts match highlights with offset markers for client rendering.
   - Normalizes BM25 scores between 0 and 1.
2. Vector Search:
   - Generates or receives query embedding vector.
   - Calculates cosine similarity against stored vectors.
   - Extracts top-K candidates using min-heap priority queue.
3. Reciprocal Rank Fusion (RRF):
   - Computes rank score for document $d$:
     $$RRF(d) = \frac{1}{k + \text{Rank}_{\text{BM25}}(d)} + \frac{1}{k + \text{Rank}_{\text{Vector}}(d)}$$
     where default smoothing parameter $k = 60$.
   - Combines with hybrid score interpolation using parameter $\alpha \in [0, 1]$:
     $$\text{HybridScore}(d) = \alpha \cdot \text{Score}_{\text{BM25}}(d) + (1 - \alpha) \cdot \text{Score}_{\text{Vector}}(d)$$
   - Returns deduplicated list sorted by composite score, accompanied by individual BM25 and vector score breakdowns.

### 5.3 Typo Tolerance Engine
- Calculates Damerau-Levenshtein distance supporting deletions, insertions, substitutions, and adjacent transpositions.
- Threshold policy:
  - Word length 1-3 characters: 0 typos allowed (exact match only).
  - Word length 4-7 characters: 1 typo allowed.
  - Word length > 7 characters: up to 2 typos allowed.
- Dynamic fallback: When exact FTS5 query yields fewer hits than requested limit, query is re-evaluated using token variants or trigram prefix expansion.

### 5.4 Dynamic Facet Aggregator & Filter Engine
- Extracts discrete values from document JSON payloads.
- Computes cardinality counts for attributes marked as facetable.
- Supports structured filter expressions:
  - Equality: `category = "electronics"`
  - Range: `price >= 50 AND price <= 500`
  - Set inclusion: `tags IN ["go", "react"]`
  - Negation: `in_stock != false`

### 5.5 Standalone 2KB Command Palette Widget
- Single `<script>` tag inclusion without dependencies.
- Global shortcut listener: `Cmd+K` on macOS, `Ctrl+K` on Windows/Linux, forward slash `/`, or click trigger.
- Accessible modal window with backdrop blur, focus trap, and clean animations.
- Input debounce (150ms) with animated loading indicator.
- Keyboard navigation: Arrow Up and Down to cycle items, Enter to navigate, Escape to dismiss.
- Highlighted token matching on title and snippet fields.
- Visual customizability via data attributes:
  - `data-url`: SearchCraft backend URL
  - `data-index`: Default index name
  - `data-placeholder`: Input placeholder text
  - `data-accent`: Theme accent color (hex or CSS color)
  - `data-shortcut`: Custom keyboard trigger

### 5.6 Embedded React 19 Admin Studio
- Single Page Application compiled and embedded into the Go binary.
- Views:
  - **Search Sandbox**: Test queries, adjust BM25/Vector sliders, review score breakdowns and latency.
  - **Documents**: Manage index records, insert single or batch JSON documents, delete items, view raw JSON.
  - **Indexes**: Create and configure indexes, set field boost weights.
  - **Analytics**: Overview cards (total searches, p50/p95/p99 latency, zero-result rate), popular query list, zero-results table.
  - **Widget Embed Generator**: Visual configuration interface with code generation and live interactive preview.

---

## 6. Implementation Stages & Verification Checklist

1. [x] Phase 0: Repository setup, license, humanized documentation, and portfolio registration.
2. [x] Phase 1: Go module structure, SQLite storage layer, FTS5 virtual tables, WAL mode, and migrations.
3. [x] Phase 2: Embedding engines (Fast deterministic local n-gram vectorizer, ONNX MiniLM, and remote client).
4. [x] Phase 3: BM25 ranking, vector cosine similarity, and Reciprocal Rank Fusion combiner (RRF).
5. [x] Phase 4: Typo-tolerance engine (Damerau-Levenshtein) and dynamic facet aggregator.
6. [x] Phase 5: High-speed ingestion API, REST endpoints, RBAC authentication, and query analytics tracker.
7. [x] Phase 6: Standalone 2KB Cmd+K search widget with auto themes and i18n support.
8. [x] Phase 7: Embedded React 19 Admin Studio with live playground, doc manager, crawler, keys, and synonyms.
9. [x] Phase 8: CLI commands (serve, import, crawl, export, backup, key, benchmark), seed datasets, tests, and integration benchmarks.
10. [x] Phase 9: Advanced enhancements:
    - Role-based API keys (`search` vs `admin`) with SHA-256 storage
    - Synonym expansion mappings (bidirectional and one-way)
    - Merchandising rules and promoted pinned results
    - "Did you mean?" typo correction suggestions and prefix autocomplete
    - Concurrent web crawler with sitemap parsing and semantic text extraction
    - Live hot snapshot backups via SQLite `VACUUM INTO`
    - Widget multi-language localization (EN, RO, ES, FR, DE) and Light/Dark/Auto theme support
