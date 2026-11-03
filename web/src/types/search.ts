export interface IndexMeta {
  name: string;
  created_at: string;
  document_count: number;
  field_boosts: Record<string, number>;
  default_mode: string;
  embedding_provider: string;
}

export interface SearchHit {
  id: string;
  score: number;
  rrf_score: number;
  bm25_score: number;
  vector_score: number;
  document: Record<string, any>;
  highlights?: Record<string, string>;
}

export interface SearchResponse {
  hits: SearchHit[];
  total: number;
  took_ms: number;
  query: string;
  mode: string;
  facets?: Record<string, Record<string, number>>;
}

export interface DocumentItem {
  id: string;
  index?: string;
  data: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface AnalyticsOverview {
  total_searches: number;
  zero_result_searches: number;
  zero_result_rate: number;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
}

export interface PopularQuery {
  query: string;
  count: number;
  avg_hits: number;
}

export interface ZeroResultQuery {
  query: string;
  count: number;
  last_searched: string;
}

export interface SystemHealth {
  status: string;
  version: string;
  uptime_seconds: number;
  memory_bytes: number;
  total_indexes: number;
  total_documents: number;
}
