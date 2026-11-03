# SearchCraft

Lightweight single-binary hybrid search engine combining BM25 full-text search and dense vector embeddings with Reciprocal Rank Fusion (RRF), typo tolerance, and a 2KB drop-in search widget.

---

## Overview

SearchCraft provides a self-hosted search engine designed for SaaS applications, documentation portals, and content websites. It offers instant search capabilities without requiring external JVM clusters or cloud search subscriptions.

Key characteristics:
- Memory footprint under 40MB RAM under normal workloads.
- Single Go binary with embedded SQLite FTS5 and vector capabilities.
- Hybrid search combining BM25 keyword matching and dense vector similarity through Reciprocal Rank Fusion (RRF).
- Damerau-Levenshtein typo-tolerance with adaptive token thresholds.
- Dynamic facet aggregation and range filtering.
- Standalone 2KB command palette search widget (Cmd+K) embeddable via a single `<script>` tag.
- Built-in React 19 administration studio for query debugging, analytics, and document inspection.

---

## Architecture

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
| * Schema Auto-Inference       |               | * Typo Tolerance Engine       |
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

## Quick Start

### 1. Build and Run

Clone the repository and build the binary:

```bash
git clone https://github.com/alexandrmotologa/searchcraft.git
cd searchcraft
go build -o searchcraft ./cmd/searchcraft
./searchcraft serve --port 7700 --db ./searchcraft.db
```

The server starts on port `7700`. The admin studio is accessible at `http://localhost:7700`.

### 2. Create an Index

```bash
curl -X POST http://localhost:7700/api/v1/indexes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "field_boosts": {
      "title": 3.0,
      "description": 1.0,
      "brand": 1.5
    }
  }'
```

### 3. Ingest Documents

```bash
curl -X POST http://localhost:7700/api/v1/indexes/products/documents \
  -H "Content-Type: application/json" \
  -d '[
    {
      "id": "kb_01",
      "title": "Wireless Mechanical Keyboard",
      "description": "Compact 75% wireless mechanical keyboard with hot-swappable switches",
      "category": "Keyboards",
      "brand": "Keychron",
      "price": 89.99,
      "in_stock": true
    },
    {
      "id": "ms_01",
      "title": "Ergonomic Wireless Mouse",
      "description": "Ergonomic vertical mouse with low latency wireless connection",
      "category": "Mice",
      "brand": "Logitech",
      "price": 69.99,
      "in_stock": true
    }
  ]'
```

### 4. Search

```bash
curl -X POST http://localhost:7700/api/v1/indexes/products/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mechanical wireless",
    "mode": "hybrid",
    "limit": 10,
    "facets": ["category", "brand"]
  }'
```

---

## Drop-In Search Widget

Add the standalone search widget to any webpage with a single script tag:

```html
<script
  src="http://localhost:7700/widget.js"
  data-url="http://localhost:7700"
  data-index="products"
  data-placeholder="Search products..."
  data-accent="#3b82f6"
></script>
```

Press `Cmd+K` (macOS) or `Ctrl+K` (Windows/Linux) to open the palette. Use the arrow keys and `Enter` to navigate results.

---

## Documentation

- [REST API Specification](docs/API_SPEC.md)
- [Hybrid Search & RRF Architecture](docs/HYBRID_SEARCH.md)
- [Widget Integration Guide](docs/WIDGET_INTEGRATION.md)
- [Performance Benchmarks](docs/BENCHMARKS.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
