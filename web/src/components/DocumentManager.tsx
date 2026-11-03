import React, { useState, useEffect } from 'react';
import {
  Database,
  Plus,
  Trash2,
  Upload,
  RefreshCw,
  Search,
  BookOpen,
  ShoppingBag,
  HelpCircle,
  CheckCircle2,
  AlertCircle
} from 'lucide-react';
import {
  fetchDocuments,
  upsertDocuments,
  deleteDocument,
  createIndex,
  deleteIndex,
} from '../api/client';
import { IndexMeta, DocumentItem } from '../types/search';

interface DocumentManagerProps {
  indexes: IndexMeta[];
  selectedIndex: string;
  onSelectIndex: (name: string) => void;
  onRefreshIndexes: () => void;
}

export const DocumentManager: React.FC<DocumentManagerProps> = ({
  indexes,
  selectedIndex,
  onSelectIndex,
  onRefreshIndexes,
}) => {
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [searchFilter, setSearchFilter] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showCreateIndexModal, setShowCreateIndexModal] = useState(false);
  const [newIndexName, setNewIndexName] = useState('');
  const [rawJsonInput, setRawJsonInput] = useState('');
  const [statusMessage, setStatusMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadDocuments = async () => {
    if (!selectedIndex) return;
    setLoading(true);
    try {
      const res = await fetchDocuments(selectedIndex, 50, 0);
      setDocuments(res.documents || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDocuments();
  }, [selectedIndex]);

  const handleCreateNewIndex = async () => {
    if (!newIndexName.trim()) return;
    try {
      await createIndex(newIndexName.trim());
      onRefreshIndexes();
      onSelectIndex(newIndexName.trim());
      setShowCreateIndexModal(false);
      setNewIndexName('');
      setStatusMessage({ type: 'success', text: `Index "${newIndexName}" created successfully.` });
    } catch (err: any) {
      setStatusMessage({ type: 'error', text: err.message });
    }
  };

  const handleDeleteCurrentIndex = async () => {
    if (!selectedIndex) return;
    if (!window.confirm(`Are you sure you want to delete index "${selectedIndex}" and all its documents?`)) {
      return;
    }
    try {
      await deleteIndex(selectedIndex);
      onRefreshIndexes();
      setStatusMessage({ type: 'success', text: `Index "${selectedIndex}" deleted.` });
    } catch (err: any) {
      setStatusMessage({ type: 'error', text: err.message });
    }
  };

  const handleAddDocuments = async () => {
    try {
      const parsed = JSON.parse(rawJsonInput);
      const docsArray = Array.isArray(parsed) ? parsed : [parsed];
      await upsertDocuments(selectedIndex, docsArray);
      setShowAddModal(false);
      setRawJsonInput('');
      loadDocuments();
      onRefreshIndexes();
      setStatusMessage({ type: 'success', text: `Successfully indexed ${docsArray.length} documents.` });
    } catch (err: any) {
      setStatusMessage({ type: 'error', text: `Failed to insert JSON: ${err.message}` });
    }
  };

  const handleDeleteDoc = async (id: string) => {
    try {
      await deleteDocument(selectedIndex, id);
      setDocuments((prev) => prev.filter((d) => d.id !== id));
      setTotal((prev) => prev - 1);
      onRefreshIndexes();
    } catch (err: any) {
      alert(`Delete failed: ${err.message}`);
    }
  };

  const loadPresetDataset = async (preset: 'docs' | 'ecommerce' | 'helpdesk') => {
    let presetDocs: any[] = [];

    if (preset === 'docs') {
      presetDocs = [
        {
          id: 'doc_1',
          title: 'Getting Started with SQLite FTS5',
          description: 'A comprehensive guide on tokenization, porter stemmers, and BM25 ranking inside SQLite.',
          category: 'Database',
          tags: ['sqlite', 'fts5', 'fulltext'],
        },
        {
          id: 'doc_2',
          title: 'Dense Vectors and Cosine Similarity',
          description: 'How high-dimensional embeddings enable semantic similarity search and clustering.',
          category: 'AI & Vectors',
          tags: ['vector', 'cosine', 'embeddings'],
        },
        {
          id: 'doc_3',
          title: 'Reciprocal Rank Fusion (RRF) Deep Dive',
          description: 'Merging multiple ranking lists with smoothing constant k=60 to eliminate score scale disparities.',
          category: 'Search Engine',
          tags: ['rrf', 'hybrid', 'ranking'],
        },
        {
          id: 'doc_4',
          title: 'Building Real-time Command Palettes in React 19',
          description: 'Accessible modal dialogs, keyboard navigation, focus trapping, and debounced queries.',
          category: 'Frontend',
          tags: ['react', 'ui', 'accessibility'],
        },
        {
          id: 'doc_5',
          title: 'Damerau-Levenshtein Automaton for Typo Tolerance',
          description: 'Calculating minimum edit distance including character transpositions in sub-millisecond execution.',
          category: 'Algorithms',
          tags: ['typo', 'levenshtein', 'string'],
        },
      ];
    } else if (preset === 'ecommerce') {
      presetDocs = [
        {
          id: 'prod_1',
          title: 'Keychron Q1 Pro Wireless Custom Mechanical Keyboard',
          description: '75% layout, CNC aluminum body, hot-swappable switches, wireless Bluetooth 5.1 and Type-C.',
          category: 'Keyboards',
          brand: 'Keychron',
          price: 199.00,
          in_stock: true,
          tags: ['wireless', 'aluminum', 'mechanical'],
        },
        {
          id: 'prod_2',
          title: 'Logitech MX Master 3S Wireless Performance Mouse',
          description: 'Quiet clicks, 8K DPI any-surface tracking, MagSpeed electromagnetic scrolling wheel.',
          category: 'Mice',
          brand: 'Logitech',
          price: 99.99,
          in_stock: true,
          tags: ['wireless', 'ergonomic', 'productivity'],
        },
        {
          id: 'prod_3',
          title: 'GMMK Pro 75% Gasket-Mounted Mechanical Keyboard',
          description: 'Rotary encoder knob, fully modular switch sockets, RGB backlighting, anodized case.',
          category: 'Keyboards',
          brand: 'Glorious',
          price: 169.99,
          in_stock: true,
          tags: ['mechanical', 'rgb', 'custom'],
        },
        {
          id: 'prod_4',
          title: 'Razer DeathAdder V3 Pro Ultra-Lightweight Wireless Mouse',
          description: '63g ergonomic esports mouse, Focus Pro 30K optical sensor, optical switches Gen-3.',
          category: 'Mice',
          brand: 'Razer',
          price: 149.99,
          in_stock: true,
          tags: ['gaming', 'wireless', 'ultralight'],
        },
      ];
    } else {
      presetDocs = [
        {
          id: 'kb_1',
          title: 'How do I reset my account password?',
          description: 'Click on the forgot password link on the login screen and follow the email verification instructions.',
          category: 'Account',
          tags: ['password', 'login', 'security'],
        },
        {
          id: 'kb_2',
          title: 'Configuring custom domains and SSL certificates',
          description: 'Point your CNAME record to our edge cluster and automatic Let\'s Encrypt SSL will provision within 5 minutes.',
          category: 'Infrastructure',
          tags: ['domains', 'ssl', 'dns'],
        },
        {
          id: 'kb_3',
          title: 'API rate limits and quota management',
          description: 'Standard plans allow 1,000 requests per minute. Contact enterprise support for dedicated quotas.',
          category: 'Billing & Quotas',
          tags: ['api', 'rate-limit', 'billing'],
        },
      ];
    }

    try {
      await upsertDocuments(selectedIndex, presetDocs);
      loadDocuments();
      onRefreshIndexes();
      setStatusMessage({ type: 'success', text: `Loaded ${presetDocs.length} preset documents into "${selectedIndex}".` });
    } catch (err: any) {
      setStatusMessage({ type: 'error', text: err.message });
    }
  };

  const filteredDocs = documents.filter((doc) => {
    if (!searchFilter.trim()) return true;
    const jsonStr = JSON.stringify(doc.data).toLowerCase();
    return jsonStr.includes(searchFilter.toLowerCase());
  });

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      {statusMessage && (
        <div
          className={`mb-6 p-4 rounded-xl flex items-center justify-between text-sm ${
            statusMessage.type === 'success'
              ? 'bg-emerald-950/60 border border-emerald-800 text-emerald-200'
              : 'bg-rose-950/60 border border-rose-800 text-rose-200'
          }`}
        >
          <div className="flex items-center gap-2">
            {statusMessage.type === 'success' ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-400" />
            ) : (
              <AlertCircle className="w-5 h-5 text-rose-400" />
            )}
            <span>{statusMessage.text}</span>
          </div>
          <button
            onClick={() => setStatusMessage(null)}
            className="text-xs hover:underline"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Index Selector Bar */}
      <div className="bg-card border border-border rounded-2xl p-6 shadow-sm mb-6 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-3 w-full md:w-auto">
          <Database className="w-5 h-5 text-blue-400" />
          <span className="text-sm font-semibold text-slate-200">Active Index:</span>
          <select
            value={selectedIndex}
            onChange={(e) => onSelectIndex(e.target.value)}
            className="bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-sm font-medium text-slate-100 focus:outline-none focus:border-blue-500"
          >
            {indexes.map((idx) => (
              <option key={idx.name} value={idx.name}>
                {idx.name} ({idx.document_count} records)
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          <button
            onClick={() => setShowCreateIndexModal(true)}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
          >
            <Plus className="w-4 h-4 text-blue-400" />
            New Index
          </button>

          {indexes.length > 1 && (
            <button
              onClick={handleDeleteCurrentIndex}
              className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-rose-950/40 hover:bg-rose-900/60 text-rose-300 border border-rose-900/50 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
              Delete Index
            </button>
          )}
        </div>
      </div>

      {/* Quick Preset Dataset Bar */}
      <div className="bg-slate-900/60 border border-border rounded-2xl p-4 mb-6 flex flex-wrap items-center justify-between gap-3">
        <div className="text-xs text-slate-400 font-medium">
          Quickly populate <strong>{selectedIndex}</strong> with realistic benchmark data:
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <button
            onClick={() => loadPresetDataset('docs')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
          >
            <BookOpen className="w-3.5 h-3.5 text-blue-400" />
            + Dev Docs
          </button>
          <button
            onClick={() => loadPresetDataset('ecommerce')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
          >
            <ShoppingBag className="w-3.5 h-3.5 text-indigo-400" />
            + E-Commerce
          </button>
          <button
            onClick={() => loadPresetDataset('helpdesk')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
          >
            <HelpCircle className="w-3.5 h-3.5 text-emerald-400" />
            + Helpdesk
          </button>
        </div>
      </div>

      {/* Documents Explorer Header */}
      <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 mb-6">
          <div className="relative w-full sm:w-80">
            <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
              placeholder="Filter indexed records..."
              className="w-full bg-slate-900 border border-border rounded-xl pl-9 pr-4 py-2 text-xs text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
            <button
              onClick={loadDocuments}
              className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 border border-border"
              title="Refresh"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={() => setShowAddModal(true)}
              className="flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm transition-all"
            >
              <Plus className="w-4 h-4" />
              Add Documents
            </button>
          </div>
        </div>

        {/* Documents Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-border text-slate-400 uppercase font-mono tracking-wider">
                <th className="py-3 px-4">Document ID</th>
                <th className="py-3 px-4">Title / Name</th>
                <th className="py-3 px-4">Category</th>
                <th className="py-3 px-4">Attributes</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {filteredDocs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-slate-500">
                    No documents found in this index. Click "Add Documents" or load a preset above.
                  </td>
                </tr>
              ) : (
                filteredDocs.map((doc) => {
                  const title = doc.data.title || doc.data.name || 'Untitled';
                  const category = doc.data.category || doc.data.brand || '-';

                  return (
                    <tr key={doc.id} className="hover:bg-slate-900/40 transition-colors">
                      <td className="py-3 px-4 font-mono font-medium text-blue-400">{doc.id}</td>
                      <td className="py-3 px-4 font-medium text-slate-100">{title}</td>
                      <td className="py-3 px-4 text-slate-300">
                        <span className="px-2 py-0.5 rounded-full bg-slate-800 text-[10px] font-mono border border-slate-700">
                          {category}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono text-[11px] text-slate-400 max-w-xs truncate">
                        {JSON.stringify(doc.data)}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <button
                          onClick={() => handleDeleteDoc(doc.id)}
                          className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-950/40 transition-colors"
                          title="Delete Document"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Document Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border border-border rounded-2xl max-w-xl w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-slate-100 mb-2">Add Documents (JSON)</h3>
            <p className="text-xs text-slate-400 mb-4">
              Paste a single JSON document or an array of documents. Each document should have an <code>id</code>, <code>title</code>, and other facetable attributes.
            </p>

            <textarea
              rows={10}
              value={rawJsonInput}
              onChange={(e) => setRawJsonInput(e.target.value)}
              placeholder={`[\n  {\n    "id": "item_1",\n    "title": "Wireless Ergonomic Mouse",\n    "description": "Optical sensor with multi-device pairing",\n    "category": "Mice",\n    "price": 49.99\n  }\n]`}
              className="w-full bg-slate-950 border border-border rounded-xl p-3 text-xs font-mono text-slate-200 placeholder:text-slate-600 focus:outline-none focus:border-blue-500 mb-4"
            />

            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => setShowAddModal(false)}
                className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-400 hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                onClick={handleAddDocuments}
                className="px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm"
              >
                Index Documents
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Index Modal */}
      {showCreateIndexModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border border-border rounded-2xl max-w-md w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-slate-100 mb-2">Create New Index</h3>
            <p className="text-xs text-slate-400 mb-4">
              Enter a unique identifier for the new search index (alphanumeric and dashes/underscores).
            </p>

            <input
              type="text"
              value={newIndexName}
              onChange={(e) => setNewIndexName(e.target.value)}
              placeholder="e.g. blog_posts, products, help_articles"
              className="w-full bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-blue-500 mb-4"
            />

            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => setShowCreateIndexModal(false)}
                className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-400 hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateNewIndex}
                className="px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm"
              >
                Create Index
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
