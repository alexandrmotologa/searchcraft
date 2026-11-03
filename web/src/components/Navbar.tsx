import React from 'react';
import { Search, Database, ArrowRightLeft, Globe, BarChart3, Key, Code2, Cpu } from 'lucide-react';
import { SystemHealth } from '../types/search';

export type NavTab = 'sandbox' | 'documents' | 'synonyms' | 'crawler' | 'analytics' | 'widget' | 'keys';

interface NavbarProps {
  activeTab: NavTab;
  setActiveTab: (tab: NavTab) => void;
  health: SystemHealth | null;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab, health }) => {
  const memoryMb = health ? (health.memory_bytes / (1024 * 1024)).toFixed(1) : '<40';

  const navItems: { id: NavTab; label: string; icon: any }[] = [
    { id: 'sandbox', label: 'Sandbox', icon: Search },
    { id: 'documents', label: 'Documents', icon: Database },
    { id: 'synonyms', label: 'Synonyms & Rules', icon: ArrowRightLeft },
    { id: 'crawler', label: 'Web Crawler', icon: Globe },
    { id: 'analytics', label: 'Analytics', icon: BarChart3 },
    { id: 'keys', label: 'API Keys', icon: Key },
    { id: 'widget', label: 'Widget', icon: Code2 },
  ];

  return (
    <header className="border-b border-border bg-card/60 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-blue-500/20 text-white font-bold text-lg">
            S
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-slate-100 tracking-tight text-lg">SearchCraft</span>
              <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20">
                v1.0.0
              </span>
            </div>
            <p className="text-xs text-muted">BM25 + Vector Hybrid Engine</p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="hidden md:flex items-center gap-1 bg-slate-900/80 p-1 rounded-xl border border-border">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                  isActive
                    ? 'bg-blue-600 text-white shadow-sm'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                {item.label}
              </button>
            );
          })}
        </nav>

        {/* System Vitals Pill */}
        <div className="flex items-center gap-3 text-xs">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-900 border border-border">
            <Cpu className="w-3.5 h-3.5 text-emerald-400" />
            <span className="text-slate-400">RAM:</span>
            <span className="font-mono font-medium text-slate-200">{memoryMb} MB</span>
          </div>

          <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-900 border border-border">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
            <span className="text-slate-300 font-medium">Ready</span>
          </div>
        </div>
      </div>

      {/* Mobile nav row */}
      <div className="md:hidden flex items-center overflow-x-auto p-2 bg-slate-950 border-t border-border gap-1">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              className={`flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs whitespace-nowrap ${
                isActive ? 'bg-blue-600 text-white' : 'text-slate-400'
              }`}
            >
              <Icon className="w-3.5 h-3.5" />
              {item.label}
            </button>
          );
        })}
      </div>
    </header>
  );
};
