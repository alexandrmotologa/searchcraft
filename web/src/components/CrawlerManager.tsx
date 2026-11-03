import React, { useState } from 'react';
import { Globe, Play, CheckCircle2, AlertCircle, RefreshCw, Layers } from 'lucide-react';
import { crawlWebsite } from '../api/client';
import { IndexMeta, CrawlResult } from '../types/search';

interface CrawlerManagerProps {
  indexes: IndexMeta[];
  selectedIndex: string;
  onSelectIndex: (name: string) => void;
  onRefreshIndexes: () => void;
}

export const CrawlerManager: React.FC<CrawlerManagerProps> = ({
  indexes,
  selectedIndex,
  onSelectIndex,
  onRefreshIndexes,
}) => {
  const [url, setUrl] = useState('');
  const [maxPages, setMaxPages] = useState(50);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CrawlResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleCrawl = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) return;

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const res = await crawlWebsite(selectedIndex, url.trim(), maxPages);
      setResult(res);
      onRefreshIndexes();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 py-8">
      <div className="mb-8">
        <h2 className="text-xl font-bold text-slate-100 tracking-tight">Integrated Web Crawler</h2>
        <p className="text-xs text-slate-400 mt-1">
          Automatically extract and index documentation pages, blog posts, and articles from any website or sitemap.xml.
        </p>
      </div>

      <div className="bg-card border border-border rounded-2xl p-6 shadow-sm mb-8">
        <form onSubmit={handleCrawl} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">
                Target Website URL or Sitemap (.xml)
              </label>
              <input
                type="text"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://example.com/sitemap.xml or https://example.com/docs"
                className="w-full bg-slate-900 border border-border rounded-xl px-3.5 py-2.5 text-xs text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">Target Index</label>
              <select
                value={selectedIndex}
                onChange={(e) => onSelectIndex(e.target.value)}
                className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2.5 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
              >
                {indexes.map((idx) => (
                  <option key={idx.name} value={idx.name}>
                    {idx.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex items-center justify-between pt-2">
            <div className="flex items-center gap-3">
              <label className="text-xs text-slate-400">Max Pages:</label>
              <input
                type="number"
                min={1}
                max={200}
                value={maxPages}
                onChange={(e) => setMaxPages(parseInt(e.target.value, 10) || 50)}
                className="w-20 bg-slate-900 border border-border rounded-xl px-2.5 py-1.5 text-xs text-center text-slate-100"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-semibold text-white shadow-lg shadow-blue-500/20 transition-all ${
                loading
                  ? 'bg-blue-800 cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-500'
              }`}
            >
              {loading ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  <span>Crawling Website...</span>
                </>
              ) : (
                <>
                  <Globe className="w-4 h-4" />
                  <span>Start Web Crawl</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Crawl Result Summary */}
      {result && (
        <div className="bg-emerald-950/40 border border-emerald-800 rounded-2xl p-6 shadow-sm mb-6">
          <div className="flex items-center gap-2 text-emerald-400 font-bold text-sm mb-3">
            <CheckCircle2 className="w-5 h-5" />
            <span>Crawl Completed Successfully!</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs font-mono">
            <div className="bg-slate-900/80 p-3 rounded-xl border border-emerald-900/60">
              <span className="text-slate-400 block text-[10px] uppercase">Pages Crawled</span>
              <strong className="text-base text-slate-100">{result.pages_crawled}</strong>
            </div>
            <div className="bg-slate-900/80 p-3 rounded-xl border border-emerald-900/60">
              <span className="text-slate-400 block text-[10px] uppercase">Documents Indexed</span>
              <strong className="text-base text-emerald-300">{result.docs_indexed}</strong>
            </div>
            <div className="bg-slate-900/80 p-3 rounded-xl border border-emerald-900/60">
              <span className="text-slate-400 block text-[10px] uppercase">Execution Duration</span>
              <strong className="text-base text-blue-300">{result.took_ms.toFixed(1)} ms</strong>
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="bg-rose-950/40 border border-rose-800 rounded-2xl p-6 shadow-sm mb-6 flex items-start gap-3 text-xs text-rose-200">
          <AlertCircle className="w-5 h-5 text-rose-400 flex-shrink-0 mt-0.5" />
          <div>
            <strong className="block font-semibold mb-1">Crawl Operation Failed</strong>
            <p>{error}</p>
          </div>
        </div>
      )}
    </div>
  );
};
