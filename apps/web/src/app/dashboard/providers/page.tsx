"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { ProviderConnection } from "@/lib/types";
import { Plus, Trash2 } from "lucide-react";

export default function ProvidersPage() {
  const [connections, setConnections] = useState<ProviderConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({
    provider_type: "resend",
    display_name: "",
    api_key: "",
    webhook_key: "",
    is_default: true,
  });
  const [saving, setSaving] = useState(false);

  const load = async () => {
    try {
      const res = await apiFetch<{ data: ProviderConnection[] }>("/v1/provider-connections");
      setConnections(res.data);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCreate = async () => {
    setSaving(true);
    setError("");
    try {
      await apiFetch("/v1/provider-connections", {
        method: "POST",
        body: JSON.stringify({
          provider_type: form.provider_type,
          channel: "email",
          display_name: form.display_name,
          credentials: { api_key: form.api_key, webhook_key: form.webhook_key },
          is_default: form.is_default,
        }),
      });
      setShowForm(false);
      setForm({ provider_type: "resend", display_name: "", api_key: "", webhook_key: "", is_default: true });
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this provider connection?")) return;
    try {
      await apiFetch(`/v1/provider-connections/${id}`, { method: "DELETE" });
      await load();
    } catch (e) {
      setError(String(e));
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold mb-0.5">Providers</h1>
          <p className="text-sm text-gray-500">Connect email delivery providers.</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors"
        >
          <Plus size={14} /> Add provider
        </button>
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold mb-4">New provider connection</h2>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Provider</label>
              <select
                value={form.provider_type}
                onChange={(e) => setForm({ ...form, provider_type: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              >
                <option value="resend">Resend</option>
                <option value="sendgrid">SendGrid</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Display name</label>
              <input
                value={form.display_name}
                onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                placeholder="Resend Production"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API key</label>
              <input
                type="password"
                value={form.api_key}
                onChange={(e) => setForm({ ...form, api_key: e.target.value })}
                placeholder="re_..."
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Webhook key <span className="text-gray-400">(optional)</span></label>
              <input
                type="password"
                value={form.webhook_key}
                onChange={(e) => setForm({ ...form, webhook_key: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                id="is_default"
                type="checkbox"
                checked={form.is_default}
                onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
              />
              <label htmlFor="is_default" className="text-sm text-gray-700">Set as default</label>
            </div>
          </div>
          <div className="flex gap-2 mt-4">
            <button
              onClick={handleCreate}
              disabled={saving || !form.display_name || !form.api_key}
              className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
            >
              {saving ? "Saving…" : "Save"}
            </button>
            <button
              onClick={() => setShowForm(false)}
              className="text-sm text-gray-500 px-4 py-2 rounded-md hover:bg-gray-100 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-400">Loading…</p>
      ) : connections.length === 0 ? (
        <p className="text-sm text-gray-500">No provider connections yet.</p>
      ) : (
        <div className="space-y-2">
          {connections.map((c) => (
            <div key={c.id} className="bg-white border border-gray-200 rounded-lg px-5 py-4 flex items-center justify-between shadow-sm">
              <div>
                <p className="text-sm font-medium">{c.display_name}</p>
                <p className="text-xs text-gray-500 mt-0.5">
                  {c.provider_type} · {c.channel}
                  {c.is_default && <span className="ml-2 text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">default</span>}
                </p>
              </div>
              <button
                onClick={() => handleDelete(c.id)}
                className="text-gray-400 hover:text-red-500 transition-colors"
              >
                <Trash2 size={15} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
