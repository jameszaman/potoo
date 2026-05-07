"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { Plus, Trash2, Copy, Check, Terminal } from "lucide-react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface ApiKey {
  id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  created_at: string;
  last_used_at?: string;
  revoked_at?: string;
}

export default function ApiKeysPage() {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [copiedSnippet, setCopiedSnippet] = useState<string | null>(null);

  const load = async () => {
    try {
      const res = await apiFetch<{ data: ApiKey[] }>("/v1/api-keys");
      setKeys(res.data.filter((k) => !k.revoked_at));
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      const res = await apiFetch<ApiKey & { key: string }>("/v1/api-keys", {
        method: "POST",
        body: JSON.stringify({ name }),
      });
      setNewKey(res.key);
      setName("");
      setShowForm(false);
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm("Revoke this API key? It will stop working immediately.")) return;
    try {
      await apiFetch(`/v1/api-keys/${id}`, { method: "DELETE" });
      await load();
    } catch (e) {
      setError(String(e));
    }
  };

  const handleCopy = () => {
    if (!newKey) return;
    navigator.clipboard.writeText(newKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCopySnippet = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSnippet(id);
    setTimeout(() => setCopiedSnippet(null), 2000);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold mb-0.5">API keys</h1>
          <p className="text-sm text-gray-500">Keys used to authenticate API requests for this organization.</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors"
        >
          <Plus size={14} /> New key
        </button>
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {newKey && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6">
          <p className="text-sm font-medium text-amber-900 mb-2">Copy your key now — it won&apos;t be shown again.</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 text-xs bg-white border border-amber-200 rounded px-3 py-2 font-mono truncate">{newKey}</code>
            <button onClick={handleCopy} className="flex items-center gap-1.5 text-xs font-medium text-amber-800 hover:text-amber-600 transition-colors">
              {copied ? <Check size={13} /> : <Copy size={13} />}
              {copied ? "Copied" : "Copy"}
            </button>
          </div>
          <button onClick={() => setNewKey(null)} className="text-xs text-amber-700 mt-2 hover:underline">Dismiss</button>
        </div>
      )}

      {showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold mb-3">New API key</h2>
          <form onSubmit={handleCreate} className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 mb-1">Key name</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Production backend"
                required
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <button
              type="submit"
              disabled={saving || !name}
              className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
            >
              {saving ? "Creating…" : "Create"}
            </button>
            <button type="button" onClick={() => setShowForm(false)} className="text-sm text-gray-500 px-3 py-2 hover:bg-gray-100 rounded-md transition-colors">
              Cancel
            </button>
          </form>
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-400">Loading…</p>
      ) : keys.length === 0 ? (
        <p className="text-sm text-gray-500">No active API keys. Create one to start making API requests.</p>
      ) : (
        <div className="space-y-2">
          {keys.map((k) => (
            <div key={k.id} className="bg-white border border-gray-200 rounded-lg px-5 py-4 flex items-center justify-between shadow-sm">
              <div>
                <p className="text-sm font-medium">{k.name}</p>
                <p className="text-xs text-gray-400 mt-0.5 font-mono">{k.key_prefix}…</p>
                <p className="text-xs text-gray-400 mt-0.5">
                  Created {new Date(k.created_at).toLocaleDateString()}
                  {k.last_used_at && <> · Last used {new Date(k.last_used_at).toLocaleDateString()}</>}
                </p>
              </div>
              <button onClick={() => handleRevoke(k.id)} className="text-gray-400 hover:text-red-500 transition-colors">
                <Trash2 size={15} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Usage documentation */}
      <div className="mt-10">
        <div className="flex items-center gap-2 mb-4">
          <Terminal size={16} className="text-gray-500" />
          <h2 className="text-sm font-semibold text-gray-800">Using your API key</h2>
        </div>

        <div className="space-y-5">
          {/* Base URL + auth */}
          <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-3">Authentication</p>
            <p className="text-sm text-gray-700 mb-3">
              Pass your API key as a <code className="bg-gray-100 px-1 py-0.5 rounded text-xs font-mono">Bearer</code> token in the{" "}
              <code className="bg-gray-100 px-1 py-0.5 rounded text-xs font-mono">Authorization</code> header on every request.
            </p>
            <div className="relative group">
              <pre className="bg-gray-950 text-gray-100 text-xs font-mono rounded-md px-4 py-3 overflow-x-auto">{`Authorization: Bearer <your-api-key>`}</pre>
              <button
                onClick={() => handleCopySnippet("auth-header", "Authorization: Bearer <your-api-key>")}
                className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-white bg-gray-800 rounded px-2 py-1 text-xs flex items-center gap-1"
              >
                {copiedSnippet === "auth-header" ? <Check size={11} /> : <Copy size={11} />}
                {copiedSnippet === "auth-header" ? "Copied" : "Copy"}
              </button>
            </div>
            <p className="text-xs text-gray-400 mt-2">Base URL: <span className="font-mono">{API_BASE}</span></p>
          </div>

          {/* List templates */}
          <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-3">List templates</p>
            <p className="text-sm text-gray-700 mb-3">
              Fetch all templates to find the right <code className="bg-gray-100 px-1 py-0.5 rounded text-xs font-mono">key</code> to use when sending a notification.
            </p>
            {(() => {
              const snippet = `curl ${API_BASE}/v1/templates \\
  -H "Authorization: Bearer <your-api-key>"`;
              return (
                <div className="relative group">
                  <pre className="bg-gray-950 text-gray-100 text-xs font-mono rounded-md px-4 py-3 overflow-x-auto whitespace-pre">{snippet}</pre>
                  <button
                    onClick={() => handleCopySnippet("list-templates", snippet)}
                    className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-white bg-gray-800 rounded px-2 py-1 text-xs flex items-center gap-1"
                  >
                    {copiedSnippet === "list-templates" ? <Check size={11} /> : <Copy size={11} />}
                    {copiedSnippet === "list-templates" ? "Copied" : "Copy"}
                  </button>
                </div>
              );
            })()}
          </div>

          {/* Send notification */}
          <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-3">Send a notification</p>
            <p className="text-sm text-gray-700 mb-3">
              <code className="bg-gray-100 px-1 py-0.5 rounded text-xs font-mono">POST /v1/notifications</code> — queues an email delivery and returns immediately.
            </p>
            <ul className="text-xs text-gray-500 space-y-1 mb-3 list-disc list-inside">
              <li><code className="bg-gray-100 px-1 py-0.5 rounded font-mono">channel</code> — <code className="bg-gray-100 px-1 py-0.5 rounded font-mono">"email"</code>, <code className="bg-gray-100 px-1 py-0.5 rounded font-mono">"sms"</code>, or <code className="bg-gray-100 px-1 py-0.5 rounded font-mono">"push"</code></li>
              <li><code className="bg-gray-100 px-1 py-0.5 rounded font-mono">recipient</code> — use <code className="bg-gray-100 px-1 py-0.5 rounded font-mono">email</code> for an address or <code className="bg-gray-100 px-1 py-0.5 rounded font-mono">external_id</code> for your own user ID</li>
              <li><code className="bg-gray-100 px-1 py-0.5 rounded font-mono">template.key</code> — the key from your template (see List templates above)</li>
              <li><code className="bg-gray-100 px-1 py-0.5 rounded font-mono">template.data</code> — variables injected into the template body</li>
              <li><code className="bg-gray-100 px-1 py-0.5 rounded font-mono">scheduled_at</code> — optional ISO 8601 datetime to deliver in the future; omit to send immediately</li>
            </ul>
            {(() => {
              const snippet = `curl -X POST ${API_BASE}/v1/notifications \\
  -H "Authorization: Bearer <your-api-key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "channel": "email",
    "recipient": { "email": "user@example.com" },
    "template": {
      "key": "welcome",
      "data": { "name": "Alice" }
    },
    "scheduled_at": "2026-06-01T09:00:00Z"
  }'`;
              return (
                <div className="relative group">
                  <pre className="bg-gray-950 text-gray-100 text-xs font-mono rounded-md px-4 py-3 overflow-x-auto whitespace-pre">{snippet}</pre>
                  <button
                    onClick={() => handleCopySnippet("send-notification", snippet)}
                    className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-white bg-gray-800 rounded px-2 py-1 text-xs flex items-center gap-1"
                  >
                    {copiedSnippet === "send-notification" ? <Check size={11} /> : <Copy size={11} />}
                    {copiedSnippet === "send-notification" ? "Copied" : "Copy"}
                  </button>
                </div>
              );
            })()}
          </div>

          {/* Check delivery status */}
          <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-3">Check delivery status</p>
            <p className="text-sm text-gray-700 mb-3">
              Use the <code className="bg-gray-100 px-1 py-0.5 rounded text-xs font-mono">notification_id</code> returned by the send call to poll status.
            </p>
            {(() => {
              const snippet = `curl ${API_BASE}/v1/notifications/<notification_id> \\
  -H "Authorization: Bearer <your-api-key>"`;
              return (
                <div className="relative group">
                  <pre className="bg-gray-950 text-gray-100 text-xs font-mono rounded-md px-4 py-3 overflow-x-auto whitespace-pre">{snippet}</pre>
                  <button
                    onClick={() => handleCopySnippet("get-notification", snippet)}
                    className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-white bg-gray-800 rounded px-2 py-1 text-xs flex items-center gap-1"
                  >
                    {copiedSnippet === "get-notification" ? <Check size={11} /> : <Copy size={11} />}
                    {copiedSnippet === "get-notification" ? "Copied" : "Copy"}
                  </button>
                </div>
              );
            })()}
          </div>
        </div>
      </div>
    </div>
  );
}
