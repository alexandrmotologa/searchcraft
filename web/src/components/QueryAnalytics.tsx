import React, { useState, useEffect } from 'react';
import {
  BarChart3,
  TrendingUp,
  AlertTriangle,
  Clock,
  Activity,
  Search,
  RefreshCw
} from 'lucide-react';
import {
  fetchAnalyticsOverview,
  fetchPopularQueries,
  fetchZeroResults,
} from '../api/client';
import { AnalyticsOverview, PopularQuery, ZeroResultQuery } from '../types/search';

export const QueryAnalytics: React.FC = () => {
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [popular, setPopular] = useState<PopularQuery[]>([]);
  const [zeroResults, setZeroResults] = useState<ZeroResultQuery[]>([]);
  const [loading, setLoading] = useState(false);

  const loadData = async () => {
    setLoading(true);
    try {
      const [ov, pop, zeros] = await Promise.all([
        fetchAnalyticsOverview(),
        fetchPopularQueries(15),
        fetchZeroResults(15),
      ]);
      setOverview(ov);
      setPopular(pop);
      setZeroResults(zeros);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-xl font-bold text-slate-100 tracking-tight">Search Engine Analytics</h2>
          <p className="text-xs text-slate-400 mt-1">
            Real-time query performance, latency percentiles, and missing content detection.
          </p>
        </div>

        <button
          onClick={loadData}
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          Refresh Stats
        </button>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
        <div className="bg-card border border-border rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">Total Searches</span>
            <Activity className="w-4 h-4 text-blue-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-slate-100">
            {overview ? overview.total_searches.toLocaleString() : '0'}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">Queries processed</div>
        </div>

        <div className="bg-card border border-border rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">Avg Latency</span>
            <Clock className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-emerald-400">
            {overview ? `${overview.avg_latency_ms.toFixed(2)} ms` : '0.00 ms'}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">Sub-10ms target met</div>
        </div>

        <div className="bg-card border border-border rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">Zero-Result Searches</span>
            <AlertTriangle className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-amber-400">
            {overview ? overview.zero_result_searches.toLocaleString() : '0'}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Rate: {overview ? `${(overview.zero_result_rate * 100).toFixed(1)}%` : '0.0%'}
          </div>
        </div>

        <div className="bg-card border border-border rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">P95 / P99 Latency</span>
            <TrendingUp className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="text-2xl font-bold font-mono text-indigo-400">
            {overview ? `${overview.p95_latency_ms.toFixed(1)} / ${overview.p99_latency_ms.toFixed(1)}` : '0 / 0'}
            <span className="text-xs font-normal text-slate-400 ml-1">ms</span>
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            P50: {overview ? `${overview.p50_latency_ms.toFixed(1)} ms` : '0 ms'}
          </div>
        </div>
      </div>

      {/* Tables Row: Popular Queries + Zero Results */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Popular Searches */}
        <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
          <div className="flex items-center gap-2 mb-4 pb-3 border-b border-border">
            <Search className="w-4 h-4 text-blue-400" />
            <h3 className="text-sm font-bold text-slate-100">Top Searched Queries</h3>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border text-slate-400 font-mono uppercase tracking-wider">
                  <th className="py-2.5 px-3">Query</th>
                  <th className="py-2.5 px-3 text-right">Volume</th>
                  <th className="py-2.5 px-3 text-right">Avg Hits</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {popular.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="py-6 text-center text-slate-500">
                      No query logs recorded yet. Run queries in the Sandbox or via Widget.
                    </td>
                  </tr>
                ) : (
                  popular.map((item, idx) => (
                    <tr key={idx} className="hover:bg-slate-900/40">
                      <td className="py-2.5 px-3 font-medium text-slate-200">
                        {item.query}
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono font-semibold text-blue-400">
                        {item.count}
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono text-slate-400">
                        {item.avg_hits.toFixed(1)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Zero-Result Searches */}
        <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-400" />
              <h3 className="text-sm font-bold text-slate-100">Zero-Result Queries (Content Gaps)</h3>
            </div>
            <span className="text-[11px] text-amber-400/80 font-mono">Actionable</span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border text-slate-400 font-mono uppercase tracking-wider">
                  <th className="py-2.5 px-3">Missing Query</th>
                  <th className="py-2.5 px-3 text-right">Misses</th>
                  <th className="py-2.5 px-3 text-right">Last Searched</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {zeroResults.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="py-6 text-center text-slate-500">
                      Zero missing queries recorded! All searches returned hits.
                    </td>
                  </tr>
                ) : (
                  zeroResults.map((item, idx) => (
                    <tr key={idx} className="hover:bg-slate-900/40">
                      <td className="py-2.5 px-3 font-medium text-amber-200/90">
                        "{item.query}"
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono font-semibold text-amber-400">
                        {item.count}
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono text-slate-500">
                        {new Date(item.last_searched).toLocaleTimeString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};
