import React, { useState, useEffect, useCallback } from 'react';
import {
  Search,
  Sliders,
  Sparkles,
  Zap,
  Filter,
  Check,
  ChevronRight,
  ChevronDown,
  Code,
  Layers,
  ArrowRight
} from 'lucide-react';
import { executeSearch } from '../api/client';
import { IndexMeta, SearchHit } from '../types/search';

interface SearchSandboxProps {
  indexes: IndexMeta[];
  selectedIndex: string;
  onSelectIndex: (name: string) => void;
}

export const SearchSandbox: React.FC<SearchSandboxProps> = ({
  indexes,
  selectedIndex,
  onSelectIndex,
}) => {
  const [query, setQuery] = useState('');
  const [mode, setMode] = useState<'hybrid' | 'bm25' | 'vector'>('hybrid');
  const [hybridAlpha, setHybridAlpha] = useState<number>(0.5);
  const [hits, setHits] = useState<SearchHit[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [tookMs, setTookMs] = useState<number>(0);
  const [facets, setFacets] = useState<Record<string, Record<string, number>>>({});
  const [selectedFilters, setSelectedFilters] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [expandedJson, setExpandedJson] = useState<Record<string, boolean>>({});
  const [mobileFiltersOpen, setMobileFiltersOpen] = useState(false);

  const runSearch = useCallback(async () => {
    if (!selectedIndex) return;
    setLoading(true);

    try {
      const activeFilters: Record<string, any> = {};
      for (const [key, val] of Object.entries(selectedFilters)) {
        if (val) activeFilters[key] = val;
      }

      const res = await executeSearch(selectedIndex, {
        query: query.trim(),
        mode: mode,
        limit: 20,
        hybrid_alpha: hybridAlpha,
        highlight: true,
        facets: ['category', 'brand', 'tags', 'type'],
        filters: Object.keys(activeFilters).length > 0 ? activeFilters : undefined,
      });

      setHits(res.hits || []);
      setTotal(res.total || 0);
      setTookMs(res.took_ms || 0);
      setFacets(res.facets || {});
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [selectedIndex, query, mode, hybridAlpha, selectedFilters]);

  useEffect(() => {
    const timer = setTimeout(() => {
      runSearch();
    }, 120);
    return () => clearTimeout(timer);
  }, [runSearch]);

  const toggleFilter = (field: string, value: string) => {
    setSelectedFilters((prev) => {
      const next = { ...prev };
      if (next[field] === value) {
        delete next[field];
      } else {
        next[field] = value;
      }
      return next;
    });
  };

  const toggleJson = (id: string) => {
    setExpandedJson((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      {/* Search Header Controls */}
      <div className="bg-card border border-border rounded-2xl p-6 shadow-xl mb-6">
        <div className="flex flex-col md:flex-row gap-4 items-center justify-between mb-6">
          {/* Index Selector */}
          <div className="flex items-center gap-3 w-full md:w-auto">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Index:</span>
            <select
              value={selectedIndex}
              onChange={(e) => {
                onSelectIndex(e.target.value);
                setSelectedFilters({});
              }}
              className="bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-sm font-medium text-slate-100 focus:outline-none focus:border-blue-500"
            >
              {indexes.map((idx) => (
                <option key={idx.name} value={idx.name}>
                  {idx.name} ({idx.document_count} docs)
                </option>
              ))}
            </select>
          </div>

          {/* Mode Switcher */}
          <div className="flex items-center gap-1 bg-slate-900 p-1 rounded-xl border border-border w-full md:w-auto justify-center">
            <button
              onClick={() => setMode('hybrid')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                mode === 'hybrid'
                  ? 'bg-blue-600 text-white shadow'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Sparkles className="w-3.5 h-3.5" />
              Hybrid (RRF)
            </button>
            <button
              onClick={() => setMode('bm25')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                mode === 'bm25'
                  ? 'bg-blue-600 text-white shadow'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Zap className="w-3.5 h-3.5" />
              BM25 Keyword
            </button>
            <button
              onClick={() => setMode('vector')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                mode === 'vector'
                  ? 'bg-blue-600 text-white shadow'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Layers className="w-3.5 h-3.5" />
              Dense Vector
            </button>
          </div>
        </div>

        {/* Query Input */}
        <div className="relative mb-4">
          <Search className="w-5 h-5 text-slate-400 absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={`Search across ${selectedIndex || 'documents'} with instant hybrid BM25 + vector ranking...`}
            className="w-full bg-slate-900/90 border border-border rounded-xl pl-12 pr-12 py-3.5 text-base text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all font-sans"
          />
          {query && (
            <button
              onClick={() => setQuery('')}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 px-2 py-1 rounded"
            >
              Clear
            </button>
          )}
        </div>

        {/* Hybrid Alpha Balance Slider */}
        {mode === 'hybrid' && (
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-3 border-t border-border/60 text-xs">
            <div className="flex items-center gap-2 text-slate-400 font-medium">
              <Sliders className="w-4 h-4 text-blue-400" />
              <span>Hybrid Score Balance (&alpha; = {hybridAlpha.toFixed(2)}):</span>
            </div>

            <div className="flex items-center gap-3 w-full sm:w-80">
              <span className={`text-[11px] font-mono ${hybridAlpha < 0.5 ? 'text-indigo-400 font-bold' : 'text-slate-500'}`}>
                Semantic (0.0)
              </span>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={hybridAlpha}
                onChange={(e) => setHybridAlpha(parseFloat(e.target.value))}
                className="flex-1 accent-blue-500 cursor-pointer"
              />
              <span className={`text-[11px] font-mono ${hybridAlpha > 0.5 ? 'text-blue-400 font-bold' : 'text-slate-500'}`}>
                Keyword (1.0)
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Main Content: Facets on Left + Results on Right */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Facet Filters Sidebar */}
        <div className="lg:col-span-1 space-y-4">
          <div className="bg-card border border-border rounded-2xl p-4 shadow-sm">
            <div className="flex items-center justify-between mb-3 pb-2 border-b border-border">
              <button
                type="button"
                onClick={() => setMobileFiltersOpen(!mobileFiltersOpen)}
                className="flex items-center gap-2 text-sm font-semibold text-slate-200 cursor-pointer lg:cursor-default"
              >
                <Filter className="w-4 h-4 text-blue-400" />
                <span>Facet Filters</span>
                {Object.keys(selectedFilters).length > 0 && (
                  <span className="px-1.5 py-0.2 rounded-full bg-blue-600/30 text-blue-300 text-[10px] font-mono border border-blue-500/30">
                    {Object.keys(selectedFilters).length}
                  </span>
                )}
                <ChevronDown className={`w-3.5 h-3.5 text-slate-400 lg:hidden transition-transform duration-200 ${mobileFiltersOpen ? 'rotate-180' : ''}`} />
              </button>
              {Object.keys(selectedFilters).length > 0 && (
                <button
                  onClick={() => setSelectedFilters({})}
                  className="text-[11px] text-blue-400 hover:text-blue-300 font-medium"
                >
                  Reset
                </button>
              )}
            </div>

            <div className={`${mobileFiltersOpen ? 'block' : 'hidden'} lg:block`}>
              {Object.keys(facets).length === 0 ? (
                <p className="text-xs text-slate-500 py-3 text-center">No facets for current search.</p>
              ) : (
                Object.entries(facets).map(([facetName, values]) => {
                  const entries = Object.entries(values);
                  if (entries.length === 0) return null;

                  return (
                    <div key={facetName} className="mb-4 last:mb-0">
                      <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                        {facetName}
                      </span>
                      <div className="space-y-1">
                        {entries.map(([val, count]) => {
                          const isSelected = selectedFilters[facetName] === val;
                          return (
                            <button
                              key={val}
                              onClick={() => toggleFilter(facetName, val)}
                              className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-colors ${
                                isSelected
                                  ? 'bg-blue-600/20 text-blue-300 border border-blue-500/30'
                                  : 'text-slate-300 hover:bg-slate-800/60'
                              }`}
                            >
                              <span className="truncate pr-2">{val}</span>
                              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-full bg-slate-800 text-slate-400">
                                {count}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>

        {/* Results Stream */}
        <div className="lg:col-span-3 space-y-4">
          {/* Status Bar */}
          <div className="flex items-center justify-between text-xs px-2 text-slate-400">
            <div>
              <span>Found </span>
              <strong className="text-slate-200 font-semibold">{total}</strong>
              <span> results in </span>
              <strong className="text-emerald-400 font-mono font-semibold">{tookMs.toFixed(2)} ms</strong>
            </div>

            <div className="flex items-center gap-2">
              <span className="font-mono text-slate-500 uppercase">{mode}</span>
            </div>
          </div>

          {/* Results List */}
          {hits.length === 0 && !loading ? (
            <div className="bg-card border border-border rounded-2xl p-12 text-center">
              <Search className="w-10 h-10 text-slate-600 mx-auto mb-3" />
              <h3 className="text-base font-semibold text-slate-200 mb-1">No matching documents found</h3>
              <p className="text-xs text-slate-500 max-w-md mx-auto">
                Try typing broader keywords, adjusting the hybrid balance slider, or clearing active facet filters.
              </p>
            </div>
          ) : (
            hits.map((hit) => {
              const doc = hit.document || {};
              const title = hit.highlights?.title || doc.title || doc.name || `Document #${hit.id}`;
              const desc = hit.highlights?.description || hit.highlights?.content || doc.description || doc.content || doc.summary || '';
              const category = doc.category || doc.brand || '';
              const isExpanded = !!expandedJson[hit.id];

              return (
                <div
                  key={hit.id}
                  className="bg-card border border-border hover:border-slate-700/80 rounded-2xl p-5 shadow-sm transition-all"
                >
                  <div className="flex items-start justify-between gap-4 mb-2">
                    <div>
                      <h4
                        className="text-base font-semibold text-slate-100 leading-snug"
                        dangerouslySetInnerHTML={{ __html: title }}
                      />
                      {category && (
                        <span className="inline-block mt-1 text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-slate-800 text-blue-400 border border-slate-700">
                          {category}
                        </span>
                      )}
                    </div>

                    {/* Score Badges */}
                    <div className="flex items-center gap-1.5 flex-shrink-0">
                      <div className="text-right">
                        <div className="text-sm font-mono font-bold text-blue-400">
                          {hit.score.toFixed(3)}
                        </div>
                        <div className="text-[10px] font-mono text-slate-500">score</div>
                      </div>
                    </div>
                  </div>

                  {desc && (
                    <p
                      className="text-sm text-slate-300 leading-relaxed line-clamp-3 mb-3"
                      dangerouslySetInnerHTML={{ __html: desc }}
                    />
                  )}

                  {/* Score Breakdown Footer */}
                  <div className="pt-3 border-t border-border flex flex-wrap items-center justify-between gap-2 text-[11px] text-slate-400">
                    <div className="flex items-center gap-3 font-mono">
                      <span>BM25: <strong className="text-slate-200">{hit.bm25_score.toFixed(3)}</strong></span>
                      <span>Vector: <strong className="text-slate-200">{hit.vector_score.toFixed(3)}</strong></span>
                      <span>RRF: <strong className="text-slate-200">{hit.rrf_score.toFixed(4)}</strong></span>
                    </div>

                    <button
                      onClick={() => toggleJson(hit.id)}
                      className="flex items-center gap-1 text-slate-400 hover:text-slate-200"
                    >
                      <Code className="w-3.5 h-3.5" />
                      <span>{isExpanded ? 'Hide JSON' : 'Inspect JSON'}</span>
                    </button>
                  </div>

                  {isExpanded && (
                    <pre className="mt-3 p-3 bg-slate-950 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto border border-border">
                      {JSON.stringify(doc, null, 2)}
                    </pre>
                  )}
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};
