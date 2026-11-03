import {
  IndexMeta,
  SearchResponse,
  DocumentItem,
  AnalyticsOverview,
  PopularQuery,
  ZeroResultQuery,
  SystemHealth,
  APIKey,
  SynonymRule,
  PromotionRule,
  CrawlResult,
} from '../types/search';

const BASE_URL = '';

export async function fetchHealth(): Promise<SystemHealth> {
  const res = await fetch(`${BASE_URL}/health`);
  if (!res.ok) throw new Error('Failed to fetch health');
  return res.json();
}

export async function fetchIndexes(): Promise<IndexMeta[]> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes`);
  if (!res.ok) throw new Error('Failed to fetch indexes');
  return res.json();
}

export async function createIndex(name: string, field_boosts?: Record<string, number>, default_mode?: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, field_boosts, default_mode }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create index');
  }
  return res.json();
}

export async function deleteIndex(name: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete index');
  return res.json();
}

export async function fetchDocuments(indexName: string, limit = 20, offset = 0): Promise<{ documents: DocumentItem[]; total: number }> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/documents?limit=${limit}&offset=${offset}`);
  if (!res.ok) throw new Error('Failed to fetch documents');
  return res.json();
}

export async function upsertDocuments(indexName: string, docs: any[]): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/documents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(docs),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to upsert documents');
  }
  return res.json();
}

export async function deleteDocument(indexName: string, docId: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/documents/${encodeURIComponent(docId)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete document');
  return res.json();
}

export async function executeSearch(indexName: string, params: {
  query: string;
  mode?: string;
  limit?: number;
  offset?: number;
  hybrid_alpha?: number;
  highlight?: boolean;
  facets?: string[];
  filters?: Record<string, any>;
}): Promise<SearchResponse> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/search`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Search query failed');
  }
  return res.json();
}

export async function fetchAnalyticsOverview(): Promise<AnalyticsOverview> {
  const res = await fetch(`${BASE_URL}/api/v1/analytics/overview`);
  if (!res.ok) throw new Error('Failed to fetch analytics');
  return res.json();
}

export async function fetchPopularQueries(limit = 10): Promise<PopularQuery[]> {
  const res = await fetch(`${BASE_URL}/api/v1/analytics/popular?limit=${limit}`);
  if (!res.ok) throw new Error('Failed to fetch popular queries');
  return res.json();
}

export async function fetchZeroResults(limit = 15): Promise<ZeroResultQuery[]> {
  const res = await fetch(`${BASE_URL}/api/v1/analytics/zero-results?limit=${limit}`);
  if (!res.ok) throw new Error('Failed to fetch zero result queries');
  return res.json();
}

// --- API Keys ---
export async function fetchAPIKeys(): Promise<APIKey[]> {
  const res = await fetch(`${BASE_URL}/api/v1/keys`);
  if (!res.ok) throw new Error('Failed to fetch api keys');
  return res.json();
}

export async function createAPIKey(name: string, role: 'admin' | 'search'): Promise<{ key: string; prefix: string; name: string; role: string }> {
  const res = await fetch(`${BASE_URL}/api/v1/keys`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, role }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create api key');
  }
  return res.json();
}

export async function deleteAPIKey(prefix: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/keys/${encodeURIComponent(prefix)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete api key');
  return res.json();
}

// --- Synonyms ---
export async function fetchSynonyms(indexName: string): Promise<SynonymRule[]> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/synonyms`);
  if (!res.ok) throw new Error('Failed to fetch synonyms');
  return res.json();
}

export async function addSynonym(indexName: string, term: string, synonyms: string[], isTwoWay = true): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/synonyms`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ term, synonyms, is_two_way: isTwoWay }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to add synonym');
  }
  return res.json();
}

export async function deleteSynonym(indexName: string, term: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/synonyms/${encodeURIComponent(term)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete synonym');
  return res.json();
}

// --- Promotions (Merchandising) ---
export async function fetchPromotions(indexName: string): Promise<PromotionRule[]> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/promotions`);
  if (!res.ok) throw new Error('Failed to fetch promotions');
  return res.json();
}

export async function addPromotion(indexName: string, queryPattern: string, pinnedDocIds: string[], position = 1): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/promotions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query_pattern: queryPattern, pinned_doc_ids: pinnedDocIds, position }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to add promotion');
  }
  return res.json();
}

export async function deletePromotion(indexName: string, queryPattern: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/promotions/${encodeURIComponent(queryPattern)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete promotion');
  return res.json();
}

// --- Web Crawler ---
export async function crawlWebsite(indexName: string, url: string, maxPages = 50): Promise<CrawlResult> {
  const res = await fetch(`${BASE_URL}/api/v1/indexes/${encodeURIComponent(indexName)}/crawl`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url, max_pages: maxPages }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Crawl failed');
  }
  return res.json();
}

// --- Backup ---
export async function backupDatabase(path?: string): Promise<{ message: string; path: string }> {
  const res = await fetch(`${BASE_URL}/api/v1/system/backup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Backup failed');
  }
  return res.json();
}

