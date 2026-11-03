# SearchCraft REST API Specification

SearchCraft exposes a JSON REST API for index management, document ingestion, hybrid search queries, synonyms, merchandising promotions, web crawling, and search analytics.

Base URL: `http://localhost:7700`

---

## 1. Authentication & Scoped API Keys

SearchCraft supports role-based API keys:
- **Search Key (`search`)**: Read-only key permitting search, autocomplete suggestions, and health checks. Safe for public web pages and client-side `<script>` widgets.
- **Admin Key (`admin`)**: Privileged key required for document ingestion, deletion, index management, crawling, synonyms, promotions, and backups.

Keys can be passed via the HTTP header:
```
Authorization: Bearer sc_live_search_...
```
or as a query parameter for client-side widgets:
```
?api_key=sc_live_search_...
```

### GET `/api/v1/keys`
Lists registered API keys (admin only).

**Response `200 OK`**:
```json
[
  {
    "prefix": "sc_live_search_27e",
    "name": "Public Documentation Widget",
    "role": "search",
    "created_at": "2026-11-03T15:30:00Z"
  }
]
```

### POST `/api/v1/keys`
Generates a new scoped API key. The raw key is returned only once upon creation.

**Request Body**:
```json
{
  "name": "Website Client Key",
  "role": "search"
}
```

**Response `201 Created`**:
```json
{
  "key": "sc_live_search_27e9958ddfa6ce61f32aca2c7fc642cd2d04c9847f4c5e8e",
  "prefix": "sc_live_search_27e",
  "name": "Website Client Key",
  "role": "search",
  "created_at": "2026-11-03T15:30:00Z"
}
```

### DELETE `/api/v1/keys/{prefix}`
Revokes an API key by prefix.

---

## 2. System Health

### GET `/health`
Returns system status, active version, document statistics, and memory usage.

**Response `200 OK`**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime_seconds": 1284,
  "memory_bytes": 545744,
  "total_indexes": 3,
  "total_documents": 14250
}
```

---

## 3. Index Management

### GET `/api/v1/indexes`
Lists all available search indexes along with document counts and configuration.

### POST `/api/v1/indexes`
Creates a new search index.

**Request Body**:
```json
{
  "name": "products",
  "field_boosts": {
    "title": 3.0,
    "description": 1.0,
    "tags": 2.0
  },
  "default_mode": "hybrid"
}
```

### DELETE `/api/v1/indexes/{name}`
Deletes an index and all associated data.

---

## 4. Document Ingestion & Export

### POST `/api/v1/indexes/{name}/documents`
Batch inserts or updates documents in the specified index.

### GET `/api/v1/indexes/{name}/documents`
Lists documents with pagination support (`limit`, `offset`).

### GET `/api/v1/indexes/{name}/export`
Streams all indexed documents as a downloadable JSON file.

### DELETE `/api/v1/indexes/{name}/documents/{id}`
Deletes a specific document.

---

## 5. Search & Autocomplete

### POST `/api/v1/indexes/{name}/search`
Executes an instant search query using BM25 keyword matching, vector similarity, or hybrid Reciprocal Rank Fusion.

**Request Body**:
```json
{
  "query": "mechanical wireless",
  "mode": "hybrid",
  "limit": 10,
  "offset": 0,
  "hybrid_alpha": 0.5,
  "highlight": true,
  "facets": ["category", "brand"]
}
```

**Response `200 OK`**:
```json
{
  "hits": [
    {
      "id": "prod_1",
      "score": 0.942,
      "rrf_score": 0.0322,
      "bm25_score": 0.884,
      "vector_score": 0.912,
      "document": {
        "id": "prod_1",
        "title": "Wireless Mechanical Keyboard",
        "category": "Keyboards",
        "price": 199.00
      },
      "highlights": {
        "title": "<mark>Wireless</mark> <mark>Mechanical</mark> Keyboard"
      }
    }
  ],
  "total": 1,
  "took_ms": 0.92,
  "query": "mechanical wireless",
  "mode": "hybrid",
  "facets": {
    "category": { "Keyboards": 1 }
  },
  "suggestions": ["mechanical switches", "wireless mouse"]
}
```

### GET `/api/v1/indexes/{name}/suggest?q={prefix}&limit=5`
Returns autocomplete suggestions matching a query prefix.

---

## 6. Synonyms Engine

### GET `/api/v1/indexes/{name}/synonyms`
Lists all synonym rules registered for an index.

### POST `/api/v1/indexes/{name}/synonyms`
Adds a synonym mapping.

**Request Body**:
```json
{
  "term": "laptop",
  "synonyms": ["notebook", "macbook"],
  "is_two_way": true
}
```

### DELETE `/api/v1/indexes/{name}/synonyms/{term}`
Deletes a synonym rule.

---

## 7. Merchandising & Pinned Results

### GET `/api/v1/indexes/{name}/promotions`
Lists active promotion rules.

### POST `/api/v1/indexes/{name}/promotions`
Pins specific documents to top ranks when a query matches the designated pattern.

**Request Body**:
```json
{
  "query_pattern": "pricing",
  "pinned_doc_ids": ["doc_pricing_plans"],
  "position": 1
}
```

### DELETE `/api/v1/indexes/{name}/promotions/{query}`
Deletes a promotion rule.

---

## 8. Integrated Web Crawler

### POST `/api/v1/indexes/{name}/crawl`
Crawls a website or XML sitemap, extracting pages and body text into the index.

**Request Body**:
```json
{
  "url": "https://docs.example.com/sitemap.xml",
  "max_pages": 50
}
```

**Response `200 OK`**:
```json
{
  "url": "https://docs.example.com/sitemap.xml",
  "pages_crawled": 34,
  "docs_indexed": 34,
  "took_ms": 1420.5
}
```

---

## 9. Backups & Snapshots

### POST `/api/v1/system/backup`
Performs an atomic hot snapshot using SQLite `VACUUM INTO`.

**Request Body**:
```json
{
  "path": "./backup_prod.db"
}
```
