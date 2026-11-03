import React, { useState } from 'react';
import { Code2, Copy, Check, Sparkles, Download, Terminal } from 'lucide-react';
import { IndexMeta } from '../types/search';

interface WidgetEmbedProps {
  indexes: IndexMeta[];
  selectedIndex: string;
}

export const WidgetEmbed: React.FC<WidgetEmbedProps> = ({ indexes, selectedIndex }) => {
  const [targetIndex, setTargetIndex] = useState(selectedIndex || 'products');
  const [placeholder, setPlaceholder] = useState('Search docs or products (Cmd+K)...');
  const [accentColor, setAccentColor] = useState('#3b82f6');
  const [shortcut, setShortcut] = useState('k');
  const [theme, setTheme] = useState('auto');
  const [lang, setLang] = useState('en');
  const [apiKey, setApiKey] = useState('');
  const [limit, setLimit] = useState(8);
  const [copied, setCopied] = useState(false);

  const baseUrl = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:7700';

  const embedCode = `<script
  src="${baseUrl}/widget.js"
  data-url="${baseUrl}"
  data-index="${targetIndex}"${apiKey ? `\n  data-api-key="${apiKey}"` : ''}
  data-theme="${theme}"
  data-lang="${lang}"
  data-placeholder="${placeholder}"
  data-accent="${accentColor}"
  data-shortcut="${shortcut}"
  data-limit="${limit}"
></script>`;

  const copyToClipboard = () => {
    navigator.clipboard.writeText(embedCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const openLiveWidget = () => {
    if ((window as any).SearchCraft) {
      (window as any).SearchCraft.open();
    } else {
      const script = document.createElement('script');
      script.src = `${baseUrl}/widget.js`;
      script.setAttribute('data-url', baseUrl);
      script.setAttribute('data-index', targetIndex);
      script.setAttribute('data-theme', theme);
      script.setAttribute('data-lang', lang);
      if (apiKey) script.setAttribute('data-api-key', apiKey);
      script.setAttribute('data-placeholder', placeholder);
      script.setAttribute('data-accent', accentColor);
      script.onload = () => {
        if ((window as any).SearchCraft) {
          (window as any).SearchCraft.open();
        }
      };
      document.body.appendChild(script);
    }
  };

  const handleExportJSON = () => {
    window.open(`${baseUrl}/api/v1/indexes/${encodeURIComponent(targetIndex)}/export`, '_blank');
  };

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <h2 className="text-xl font-bold text-slate-100 tracking-tight">Drop-In Search Widget Generator</h2>
          <p className="text-xs text-slate-400 mt-1">
            Embed an instant Cmd+K search palette into any website with a single 2KB script tag. Zero dependencies.
          </p>
        </div>

        <button
          onClick={handleExportJSON}
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
        >
          <Download className="w-3.5 h-3.5 text-blue-400" />
          Export Index JSON
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Customization Options */}
        <div className="lg:col-span-1 bg-card border border-border rounded-2xl p-6 shadow-sm space-y-4">
          <h3 className="text-sm font-bold text-slate-200 pb-2 border-b border-border">Configuration</h3>

          <div>
            <label className="text-xs font-semibold text-slate-400 block mb-1.5">Target Index</label>
            <select
              value={targetIndex}
              onChange={(e) => setTargetIndex(e.target.value)}
              className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
            >
              {indexes.map((idx) => (
                <option key={idx.name} value={idx.name}>
                  {idx.name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">Theme</label>
              <select
                value={theme}
                onChange={(e) => setTheme(e.target.value)}
                className="w-full bg-slate-900 border border-border rounded-xl px-2.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
              >
                <option value="auto">Auto (System)</option>
                <option value="dark">Dark</option>
                <option value="light">Light</option>
              </select>
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">Language</label>
              <select
                value={lang}
                onChange={(e) => setLang(e.target.value)}
                className="w-full bg-slate-900 border border-border rounded-xl px-2.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
              >
                <option value="en">English</option>
                <option value="ro">Română</option>
                <option value="es">Español</option>
                <option value="fr">Français</option>
                <option value="de">Deutsch</option>
              </select>
            </div>
          </div>

          <div>
            <label className="text-xs font-semibold text-slate-400 block mb-1.5">Public API Key (Optional)</label>
            <input
              type="text"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="sc_live_search_..."
              className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs font-mono text-slate-100 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="text-xs font-semibold text-slate-400 block mb-1.5">Input Placeholder</label>
            <input
              type="text"
              value={placeholder}
              onChange={(e) => setPlaceholder(e.target.value)}
              className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="text-xs font-semibold text-slate-400 block mb-1.5">Accent Color</label>
            <div className="flex items-center gap-2">
              <input
                type="color"
                value={accentColor}
                onChange={(e) => setAccentColor(e.target.value)}
                className="w-8 h-8 rounded-lg border border-border bg-slate-900 cursor-pointer"
              />
              <input
                type="text"
                value={accentColor}
                onChange={(e) => setAccentColor(e.target.value)}
                className="flex-1 bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs font-mono text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">Shortcut Key</label>
              <input
                type="text"
                maxLength={1}
                value={shortcut}
                onChange={(e) => setShortcut(e.target.value)}
                className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs font-mono text-center text-slate-100 focus:outline-none focus:border-blue-500 uppercase"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-slate-400 block mb-1.5">Max Results</label>
              <input
                type="number"
                min={1}
                max={25}
                value={limit}
                onChange={(e) => setLimit(parseInt(e.target.value, 10) || 8)}
                className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs font-mono text-center text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div className="pt-2">
            <button
              onClick={openLiveWidget}
              className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-lg shadow-blue-500/20 transition-all"
            >
              <Sparkles className="w-4 h-4" />
              Test Live Widget (Cmd+{shortcut.toUpperCase()})
            </button>
          </div>
        </div>

        {/* Code Snippet Output */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center justify-between mb-3 pb-3 border-b border-border">
              <div className="flex items-center gap-2 text-sm font-bold text-slate-200">
                <Code2 className="w-4 h-4 text-blue-400" />
                <span>HTML Embed Code</span>
              </div>
              <button
                onClick={copyToClipboard}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                {copied ? 'Copied!' : 'Copy Code'}
              </button>
            </div>

            <pre className="bg-slate-950 border border-border rounded-xl p-4 text-xs font-mono text-blue-300 overflow-x-auto leading-relaxed">
              {embedCode}
            </pre>
          </div>

          {/* Trigger Button Snippet */}
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-3 pb-3 border-b border-border">
              <Terminal className="w-4 h-4 text-emerald-400" />
              <span className="text-sm font-bold text-slate-200">Custom Trigger Button (Optional)</span>
            </div>

            <p className="text-xs text-slate-400 mb-3">
              Add <code>data-searchcraft-trigger</code> to any button or link on your site to open the search modal on click:
            </p>

            <pre className="bg-slate-950 border border-border rounded-xl p-4 text-xs font-mono text-emerald-300 overflow-x-auto">
{`<button data-searchcraft-trigger class="search-btn">
  <span>Search documentation...</span>
  <kbd>Cmd+${shortcut.toUpperCase()}</kbd>
</button>`}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
};
