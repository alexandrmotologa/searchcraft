import React, { useState, useEffect } from 'react';
import { Sparkles, Pin, Plus, Trash2, ArrowRightLeft, ArrowRight, CheckCircle2, AlertCircle } from 'lucide-react';
import {
  fetchSynonyms,
  addSynonym,
  deleteSynonym,
  fetchPromotions,
  addPromotion,
  deletePromotion,
} from '../api/client';
import { IndexMeta, SynonymRule, PromotionRule } from '../types/search';

interface SynonymsManagerProps {
  indexes: IndexMeta[];
  selectedIndex: string;
  onSelectIndex: (name: string) => void;
}

export const SynonymsManager: React.FC<SynonymsManagerProps> = ({
  indexes,
  selectedIndex,
  onSelectIndex,
}) => {
  const [synonyms, setSynonyms] = useState<SynonymRule[]>([]);
  const [promotions, setPromotions] = useState<PromotionRule[]>([]);
  const [newTerm, setNewTerm] = useState('');
  const [newSynonyms, setNewSynonyms] = useState('');
  const [isTwoWay, setIsTwoWay] = useState(true);

  const [promoPattern, setPromoPattern] = useState('');
  const [promoDocIDs, setPromoDocIDs] = useState('');
  const [promoPos, setPromoPos] = useState(1);

  const [statusMsg, setStatusMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadData = async () => {
    if (!selectedIndex) return;
    try {
      const [syns, promos] = await Promise.all([
        fetchSynonyms(selectedIndex),
        fetchPromotions(selectedIndex),
      ]);
      setSynonyms(syns);
      setPromotions(promos);
    } catch (err: any) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedIndex]);

  const handleAddSynonym = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTerm.trim() || !newSynonyms.trim()) return;

    const synList = newSynonyms.split(',').map((s) => s.trim()).filter(Boolean);
    try {
      await addSynonym(selectedIndex, newTerm.trim(), synList, isTwoWay);
      setNewTerm('');
      setNewSynonyms('');
      loadData();
      setStatusMsg({ type: 'success', text: `Synonym for "${newTerm}" added.` });
    } catch (err: any) {
      setStatusMsg({ type: 'error', text: err.message });
    }
  };

  const handleDeleteSynonym = async (term: string) => {
    try {
      await deleteSynonym(selectedIndex, term);
      loadData();
    } catch (err: any) {
      setStatusMsg({ type: 'error', text: err.message });
    }
  };

  const handleAddPromotion = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!promoPattern.trim() || !promoDocIDs.trim()) return;

    const idList = promoDocIDs.split(',').map((s) => s.trim()).filter(Boolean);
    try {
      await addPromotion(selectedIndex, promoPattern.trim(), idList, promoPos);
      setPromoPattern('');
      setPromoDocIDs('');
      loadData();
      setStatusMsg({ type: 'success', text: `Promotion for "${promoPattern}" added.` });
    } catch (err: any) {
      setStatusMsg({ type: 'error', text: err.message });
    }
  };

  const handleDeletePromotion = async (pattern: string) => {
    try {
      await deletePromotion(selectedIndex, pattern);
      loadData();
    } catch (err: any) {
      setStatusMsg({ type: 'error', text: err.message });
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      {statusMsg && (
        <div
          className={`mb-6 p-4 rounded-xl flex items-center justify-between text-sm ${
            statusMsg.type === 'success'
              ? 'bg-emerald-950/60 border border-emerald-800 text-emerald-200'
              : 'bg-rose-950/60 border border-rose-800 text-rose-200'
          }`}
        >
          <div className="flex items-center gap-2">
            {statusMsg.type === 'success' ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-400" />
            ) : (
              <AlertCircle className="w-5 h-5 text-rose-400" />
            )}
            <span>{statusMsg.text}</span>
          </div>
          <button onClick={() => setStatusMsg(null)} className="text-xs hover:underline">
            Dismiss
          </button>
        </div>
      )}

      {/* Index Selector */}
      <div className="bg-card border border-border rounded-2xl p-6 shadow-sm mb-8 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Sparkles className="w-5 h-5 text-blue-400" />
          <span className="text-sm font-semibold text-slate-200">Active Index:</span>
          <select
            value={selectedIndex}
            onChange={(e) => onSelectIndex(e.target.value)}
            className="bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-sm font-medium text-slate-100 focus:outline-none focus:border-blue-500"
          >
            {indexes.map((idx) => (
              <option key={idx.name} value={idx.name}>
                {idx.name}
              </option>
            ))}
          </select>
        </div>
        <div className="text-xs text-slate-400">
          Synonyms expand BM25 search terms &bull; Promotions pin documents at top ranks
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Synonyms Section */}
        <div className="space-y-6">
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-4 pb-3 border-b border-border">
              <ArrowRightLeft className="w-4 h-4 text-blue-400" />
              <h3 className="text-sm font-bold text-slate-100">Synonyms Engine</h3>
            </div>

            <form onSubmit={handleAddSynonym} className="space-y-3 mb-6">
              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">Root Term</label>
                <input
                  type="text"
                  value={newTerm}
                  onChange={(e) => setNewTerm(e.target.value)}
                  placeholder="e.g. laptop"
                  className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">
                  Synonyms (comma-separated)
                </label>
                <input
                  type="text"
                  value={newSynonyms}
                  onChange={(e) => setNewSynonyms(e.target.value)}
                  placeholder="e.g. notebook, macbook, ultrabook"
                  className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex items-center justify-between pt-1">
                <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={isTwoWay}
                    onChange={(e) => setIsTwoWay(e.target.checked)}
                    className="accent-blue-500 rounded"
                  />
                  <span>Two-Way (Bidirectional mapping)</span>
                </label>

                <button
                  type="submit"
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm"
                >
                  <Plus className="w-3.5 h-3.5" />
                  Add Synonym
                </button>
              </div>
            </form>

            {/* Synonyms List */}
            <div className="space-y-2">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                Active Synonyms ({synonyms.length})
              </span>

              {synonyms.length === 0 ? (
                <p className="text-xs text-slate-500 py-4 text-center">No synonyms registered for this index.</p>
              ) : (
                synonyms.map((rule) => (
                  <div
                    key={rule.term}
                    className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-border text-xs"
                  >
                    <div className="flex items-center gap-2">
                      <strong className="text-blue-400 font-semibold">{rule.term}</strong>
                      {rule.is_two_way ? (
                        <ArrowRightLeft className="w-3.5 h-3.5 text-slate-500" />
                      ) : (
                        <ArrowRight className="w-3.5 h-3.5 text-slate-500" />
                      )}
                      <span className="text-slate-300">{rule.synonyms.join(', ')}</span>
                    </div>

                    <button
                      onClick={() => handleDeleteSynonym(rule.term)}
                      className="text-slate-400 hover:text-rose-400 p-1"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Promotions (Merchandising) Section */}
        <div className="space-y-6">
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-4 pb-3 border-b border-border">
              <Pin className="w-4 h-4 text-indigo-400" />
              <h3 className="text-sm font-bold text-slate-100">Merchandising & Pinned Results</h3>
            </div>

            <form onSubmit={handleAddPromotion} className="space-y-3 mb-6">
              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">
                  Query Pattern (Substring or Keyword)
                </label>
                <input
                  type="text"
                  value={promoPattern}
                  onChange={(e) => setPromoPattern(e.target.value)}
                  placeholder="e.g. pricing, sale, mechanical"
                  className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">
                  Pinned Document IDs (comma-separated)
                </label>
                <input
                  type="text"
                  value={promoDocIDs}
                  onChange={(e) => setPromoDocIDs(e.target.value)}
                  placeholder="e.g. prod_1, doc_pricing_plan"
                  className="w-full bg-slate-900 border border-border rounded-xl px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex items-center justify-between pt-1">
                <div className="flex items-center gap-2">
                  <label className="text-xs text-slate-400">Position:</label>
                  <input
                    type="number"
                    min={1}
                    max={10}
                    value={promoPos}
                    onChange={(e) => setPromoPos(parseInt(e.target.value, 10) || 1)}
                    className="w-16 bg-slate-900 border border-border rounded-lg px-2 py-1 text-xs text-center text-slate-100"
                  />
                </div>

                <button
                  type="submit"
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm"
                >
                  <Plus className="w-3.5 h-3.5" />
                  Pin Results
                </button>
              </div>
            </form>

            {/* Promotions List */}
            <div className="space-y-2">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                Active Promotions ({promotions.length})
              </span>

              {promotions.length === 0 ? (
                <p className="text-xs text-slate-500 py-4 text-center">No promotions registered for this index.</p>
              ) : (
                promotions.map((promo) => (
                  <div
                    key={promo.query_pattern}
                    className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-border text-xs"
                  >
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="px-2 py-0.5 rounded-full bg-indigo-950/80 text-indigo-300 border border-indigo-800 text-[10px] font-mono">
                          Rank #{promo.position}
                        </span>
                        <strong className="text-slate-100">"{promo.query_pattern}"</strong>
                      </div>
                      <div className="text-slate-400 font-mono text-[11px]">
                        Pinned IDs: {promo.pinned_doc_ids.join(', ')}
                      </div>
                    </div>

                    <button
                      onClick={() => handleDeletePromotion(promo.query_pattern)}
                      className="text-slate-400 hover:text-rose-400 p-1"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
