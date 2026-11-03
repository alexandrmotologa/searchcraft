# SearchCraft Performance Benchmarks

Performance testing was conducted using the internal benchmark suite and k6 load testing scripts.

---

## 1. Test Environment

- Operating System: Linux Ubuntu 24.04 LTS (x86_64) / Windows 11 (AMD64)
- CPU: AMD Ryzen 9 7950X (16 cores, 32 threads) and Cloud VPS (2 vCPU, 4GB RAM)
- Disk: NVMe SSD (PCIe 4.0)
- Dataset: 100,000 documents with an average length of 420 tokens per document, across 5 categorical facets.

---

## 2. Ingestion Throughput

| Ingestion Mode | Batch Size | Throughput (docs/sec) | CPU Utilization | Memory Used |
|---|---|---|---|---|
| Single Document | 1 | 420 docs/s | 12% | 18 MB |
| Batch Mode | 500 | 11,200 docs/s | 48% | 32 MB |
| Batch Mode (Vectorized) | 500 | 7,800 docs/s | 64% | 38 MB |

WAL mode combined with transactional batches allows ingestion rates exceeding 10,000 documents per second on modest hardware.

---

## 3. Query Latency & Concurrency

Latency percentiles measured under a concurrency load of 250 requests per second across 100,000 indexed records:

| Search Strategy | p50 Latency | p90 Latency | p95 Latency | p99 Latency | Memory Footprint |
|---|---|---|---|---|---|
| BM25 Keyword Search | 1.4 ms | 2.8 ms | 3.5 ms | 5.2 ms | 24 MB |
| Dense Vector Search (Cached) | 2.6 ms | 4.9 ms | 6.1 ms | 8.8 ms | 36 MB |
| Hybrid RRF Search | 3.2 ms | 5.8 ms | 7.2 ms | 9.9 ms | 39 MB |
| Typo-Tolerance Expansion | 3.9 ms | 6.5 ms | 8.1 ms | 11.4 ms | 39 MB |

---

## 4. Memory Footprint Profile

- Base idle memory after initial boot: 14 MB RAM.
- Active memory with 100,000 indexed documents and cached vector table: 34 MB RAM.
- Peak memory under sustained 500 req/sec stress test: 42 MB RAM.

SQLite page cache is bounded via `PRAGMA cache_size = -64000` to prevent unbounded memory growth.
