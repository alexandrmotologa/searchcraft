import {
  IndexMeta,
  SearchResponse,
  DocumentItem,
  AnalyticsOverview,
  PopularQuery,
  ZeroResultQuery,
  SystemHealth,
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
