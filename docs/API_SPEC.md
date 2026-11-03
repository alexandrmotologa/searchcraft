# SearchCraft REST API Specification

SearchCraft exposes a JSON REST API for index management, document ingestion, hybrid search queries, and search analytics.

Base URL: `http://localhost:7700`

---

## 1. System Health

### GET `/health`
Returns system status, active version, document statistics, and memory usage.

**Response `200 OK`**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime_seconds": 1284,
  "memory_bytes": 28410200,
  "total_indexes": 3,
  "total_documents": 14250
}
```

---

## 2. Index Management

### GET `/api/v1/indexes`
Lists all available search indexes along with document counts and configuration.

**Response `200 OK`**:
```json
[
  {
    "name": "products",
    "created_at": "2026-11-03T14:00:00Z",
    "document_count": 1240,
    "field_boosts": {
      "title": 3.0,
      "description": 1.0,
      "brand": 1.5
    },
    "default_mode": "hybrid",
    "embedding_provider": "fast_local"
  }
]
```

### POST `/api/v1/indexes`
Creates a new search index.

**Request Body**:
```json
{
  "name": "docs",
  "field_boosts": {
    "title": 3.0,
    "content": 1.0,
    "tags": 2.0
  },
  "default_mode": "hybrid",
  "embedding_provider": "fast_local"
}
```

**Response `201 Created`**:
```json
{
  "message": "Index created successfully",
  "name": "docs"
}
```

### GET `/api/v1/indexes/{name}`
Returns details and statistics for a specific index.

**Response `200 OK`**:
```json
{
  "name": "docs",
  "created_at": "2026-11-03T14:10:00Z",
  "document_count": 450,
  "field_boosts": {
    "title": 3.0,
    "content": 1.0
  },
  "default_mode": "hybrid",
  "embedding_provider": "fast_local"
}
```

### DELETE `/api/v1/indexes/{name}`
Deletes an index, including its FTS5 virtual table, document records, and stored vectors.

**Response `200 OK`**:
```json
{
  "message": "Index deleted successfully"
}
```

---

## 3. Document Ingestion

### POST `/api/v1/indexes/{name}/documents`
Batch inserts or updates documents in the specified index. Each document must contain an `id` field (string or integer). If omitted, an ID is automatically generated.

**Request Body**:
```json
[
  {
    "id": "doc_101",
    "title": "Ultra-lightweight Mechanical Keyboard",
    "description": "Compact 65% wireless mechanical keyboard with hot-swappable switches",
    "category": "Keyboards",
    "brand": "Keychron",
    "price": 89.99,
    "rating": 4.8,
    "in_stock": true,
    "tags": ["wireless", "mechanical", "compact"]
  },
  {
    "id": "doc_102",
    "title": "Ergonomic Vertical Mouse",
    "description": "Ergonomic wireless mouse designed to reduce wrist strain during long sessions",
    "category": "Mice",
    "brand": "Logitech",
    "price": 59.99,
    "rating": 4.6,
    "in_stock": true,
    "tags": ["wireless", "ergonomic", "office"]
  }
]
```

**Response `200 OK`**:
```json
{
  "indexed": 2,
  "failed": 0,
  "took_ms": 1.8
}
```

### GET `/api/v1/indexes/{name}/documents`
Lists documents with pagination support.

**Query Parameters**:
- `limit`: Number of documents to return (default: 20, max: 100)
- `offset`: Pagination offset (default: 0)

**Response `200 OK`**:
```json
{
  "documents": [
    {
      "id": "doc_101",
      "data": {
        "title": "Ultra-lightweight Mechanical Keyboard",
        "description": "Compact 65% wireless mechanical keyboard with hot-swappable switches",
        "category": "Keyboards",
        "price": 89.99
      },
      "created_at": "2026-11-03T14:15:00Z"
    }
  ],
  "total": 1240,
  "limit": 20,
  "offset": 0
}
```

### GET `/api/v1/indexes/{name}/documents/{id}`
Retrieves a single document by its identifier.

**Response `200 OK`**:
```json
{
  "id": "doc_101",
  "data": {
    "title": "Ultra-lightweight Mechanical Keyboard",
    "description": "Compact 65% wireless mechanical keyboard with hot-swappable switches",
    "category": "Keyboards",
    "price": 89.99
  },
  "created_at": "2026-11-03T14:15:00Z"
}
```

### DELETE `/api/v1/indexes/{name}/documents/{id}`
Deletes a specific document from the index, its FTS5 index, and its vector store.

**Response `200 OK`**:
```json
{
  "message": "Document deleted successfully"
}
```

---

## 4. Search Execution

### POST `/api/v1/indexes/{name}/search`
Executes an instant search query using BM25 keyword matching, vector similarity, or hybrid Reciprocal Rank Fusion.

**Request Body**:
```json
{
  "query": "mechanical keyboard wireless",
  "mode": "hybrid",
  "limit": 10,
  "offset": 0,
  "hybrid_alpha": 0.5,
  "highlight": true,
  "facets": ["category", "brand"],
  "filters": {
    "category": "Keyboards",
    "price": {
      "$gte": 50.0,
      "$lte": 150.0
    },
    "in_stock": true
  }
}
```

**Field Descriptions**:
- `query`: Text string to search.
- `mode`: Search strategy (`hybrid`, `bm25`, `vector`). Default is `hybrid`.
- `limit`: Maximum number of matching hits (default: 20).
- `offset`: Pagination offset (default: 0).
- `hybrid_alpha`: Balance between BM25 (1.0) and Vector (0.0). Default is 0.5.
- `highlight`: When true, wraps matching text tokens inside `<mark>` tags.
- `facets`: List of fields for which category counts should be calculated.
- `filters`: Structured filter criteria supporting equality, `$gte`, `$lte`, `$gt`, `$lt`, and `$in`.

**Response `200 OK`**:
```json
{
  "hits": [
    {
      "id": "doc_101",
      "score": 0.942,
      "rrf_score": 0.0322,
      "bm25_score": 0.884,
      "vector_score": 0.912,
      "document": {
        "id": "doc_101",
        "title": "Ultra-lightweight Mechanical Keyboard",
        "description": "Compact 65% wireless mechanical keyboard with hot-swappable switches",
        "category": "Keyboards",
        "brand": "Keychron",
        "price": 89.99,
        "rating": 4.8,
        "in_stock": true
      },
      "highlights": {
        "title": "Ultra-lightweight <mark>Mechanical</mark> <mark>Keyboard</mark>",
        "description": "Compact 65% <mark>wireless</mark> <mark>mechanical</mark> <mark>keyboard</mark> with hot-swappable switches"
      }
    }
  ],
  "total": 1,
  "took_ms": 2.4,
  "query": "mechanical keyboard wireless",
  "mode": "hybrid",
  "facets": {
    "category": {
      "Keyboards": 1
    },
    "brand": {
      "Keychron": 1
    }
  }
}
```

---

## 5. Analytics & Observability

### GET `/api/v1/analytics/overview`
Returns summary metrics: total search queries, zero-result search count, average query latency, and latency percentiles.

**Response `200 OK`**:
```json
{
  "total_searches": 8420,
  "zero_result_searches": 142,
  "zero_result_rate": 0.0168,
  "avg_latency_ms": 2.7,
  "p50_latency_ms": 2.1,
  "p95_latency_ms": 4.8,
  "p99_latency_ms": 8.2
}
```

### GET `/api/v1/analytics/popular`
Returns the most frequently executed search queries.

**Response `200 OK`**:
```json
[
  { "query": "mechanical keyboard", "count": 1420, "avg_hits": 18 },
  { "query": "wireless mouse", "count": 940, "avg_hits": 12 },
  { "query": "usb-c dock", "count": 512, "avg_hits": 8 }
]
```

### GET `/api/v1/analytics/zero-results`
Lists queries that produced zero matching documents, ordered by search frequency.

**Response `200 OK`**:
```json
[
  { "query": "trackball wireless red", "count": 34, "last_searched": "2026-11-03T14:22:00Z" },
  { "query": "curved oled monitor 49", "count": 21, "last_searched": "2026-11-03T13:45:00Z" }
]
```
