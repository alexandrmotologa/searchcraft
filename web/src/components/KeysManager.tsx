import React, { useState, useEffect } from 'react';
import { Key, Shield, Plus, Trash2, Copy, Check, HardDrive, Download, AlertCircle } from 'lucide-react';
import { fetchAPIKeys, createAPIKey, deleteAPIKey, backupDatabase } from '../api/client';
import { APIKey } from '../types/search';

export const KeysManager: React.FC = () => {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [keyName, setKeyName] = useState('');
  const [keyRole, setKeyRole] = useState<'admin' | 'search'>('search');
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState(false);
  const [backupStatus, setBackupStatus] = useState<string | null>(null);

  const loadKeys = async () => {
    try {
      const res = await fetchAPIKeys();
      setKeys(res);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadKeys();
  }, []);

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyName.trim()) return;

    try {
      const res = await createAPIKey(keyName.trim(), keyRole);
      setNewlyCreatedKey(res.key);
      setShowCreateModal(false);
      setKeyName('');
      loadKeys();
    } catch (err: any) {
      alert(`Failed to create key: ${err.message}`);
    }
  };

  const handleDeleteKey = async (prefix: string) => {
    if (!window.confirm(`Delete API key with prefix ${prefix}? Any requests using it will be blocked.`)) {
      return;
    }
    try {
      await deleteAPIKey(prefix);
      loadKeys();
    } catch (err: any) {
      alert(`Delete failed: ${err.message}`);
    }
  };

  const handleBackup = async () => {
    try {
      setBackupStatus('Creating snapshot...');
      const res = await backupDatabase();
      setBackupStatus(`Backup snapshot saved successfully: ${res.path}`);
      setTimeout(() => setBackupStatus(null), 5000);
    } catch (err: any) {
      setBackupStatus(`Backup failed: ${err.message}`);
    }
  };

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <h2 className="text-xl font-bold text-slate-100 tracking-tight">API Keys & Access Control</h2>
          <p className="text-xs text-slate-400 mt-1">
            Scoped permissions for public search widgets and private administrative workflows.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={handleBackup}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-border transition-colors"
          >
            <HardDrive className="w-3.5 h-3.5 text-blue-400" />
            Snapshot Database
          </button>

          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm transition-all"
          >
            <Plus className="w-4 h-4" />
            New API Key
          </button>
        </div>
      </div>

      {backupStatus && (
        <div className="mb-6 p-4 rounded-xl bg-slate-900 border border-blue-500/30 text-xs font-mono text-blue-300">
          {backupStatus}
        </div>
      )}

      {/* Newly Created Key Alert Modal */}
      {newlyCreatedKey && (
        <div className="mb-8 p-6 rounded-2xl bg-indigo-950/60 border border-indigo-700/60 shadow-xl">
          <div className="flex items-center gap-2 text-indigo-300 font-bold text-sm mb-2">
            <Shield className="w-5 h-5 text-indigo-400" />
            <span>Copy Your New API Key</span>
          </div>
          <p className="text-xs text-slate-300 mb-4">
            This secret key is only displayed once. Store it securely or add it to your search widget tag.
          </p>

          <div className="flex items-center gap-3">
            <input
              type="text"
              readOnly
              value={newlyCreatedKey}
              className="flex-1 bg-slate-950 border border-indigo-800 rounded-xl px-3.5 py-2 text-xs font-mono text-indigo-200 focus:outline-none"
            />
            <button
              onClick={() => {
                navigator.clipboard.writeText(newlyCreatedKey);
                setCopiedKey(true);
                setTimeout(() => setCopiedKey(false), 2000);
              }}
              className="flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm"
            >
              {copiedKey ? <Check className="w-4 h-4 text-emerald-300" /> : <Copy className="w-4 h-4" />}
              {copiedKey ? 'Copied' : 'Copy'}
            </button>
            <button
              onClick={() => setNewlyCreatedKey(null)}
              className="px-3 py-2 text-xs text-slate-400 hover:text-slate-200"
            >
              Done
            </button>
          </div>
        </div>
      )}

      {/* Keys Table */}
      <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-border text-slate-400 uppercase font-mono tracking-wider">
                <th className="py-3 px-4">Key Name</th>
                <th className="py-3 px-4">Prefix</th>
                <th className="py-3 px-4">Role / Scope</th>
                <th className="py-3 px-4">Created Date</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {keys.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-slate-500">
                    No API keys registered. Running in open developer mode. Click "New API Key" to secure your engine.
                  </td>
                </tr>
              ) : (
                keys.map((k) => (
                  <tr key={k.prefix} className="hover:bg-slate-900/40">
                    <td className="py-3 px-4 font-semibold text-slate-100">{k.name}</td>
                    <td className="py-3 px-4 font-mono text-slate-400">{k.prefix}...</td>
                    <td className="py-3 px-4">
                      {k.role === 'admin' ? (
                        <span className="px-2 py-0.5 rounded-full bg-purple-950/80 text-purple-300 border border-purple-800 text-[10px] font-mono">
                          Admin (Full Access)
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded-full bg-blue-950/80 text-blue-300 border border-blue-800 text-[10px] font-mono">
                          Search-Only (Public)
                        </span>
                      )}
                    </td>
                    <td className="py-3 px-4 text-slate-400">
                      {new Date(k.created_at).toLocaleDateString()}
                    </td>
                    <td className="py-3 px-4 text-right">
                      <button
                        onClick={() => handleDeleteKey(k.prefix)}
                        className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-950/40"
                        title="Revoke Key"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Key Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border border-border rounded-2xl max-w-md w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-slate-100 mb-2">Generate New API Key</h3>
            <p className="text-xs text-slate-400 mb-4">
              Select between a public search-only key (safe for client websites) or an admin key.
            </p>

            <form onSubmit={handleCreateKey} className="space-y-4">
              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">Key Name / Description</label>
                <input
                  type="text"
                  value={keyName}
                  onChange={(e) => setKeyName(e.target.value)}
                  placeholder="e.g. Website Search Widget, CI/CD Pipeline"
                  className="w-full bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-slate-400 block mb-1">Key Role</label>
                <select
                  value={keyRole}
                  onChange={(e) => setKeyRole(e.target.value as any)}
                  className="w-full bg-slate-900 border border-border rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="search">Search-Only (Public - Read only)</option>
                  <option value="admin">Admin Master (Full read and write mutations)</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-400 hover:bg-slate-800"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm"
                >
                  Generate Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
