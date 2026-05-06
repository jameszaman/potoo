"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { ProviderConnection } from "@/lib/types";
import { Plus, Trash2, Pencil } from "lucide-react";
import NoApiKey, { isApiKeyError } from "@/components/NoApiKey";

type ProviderType = "resend" | "sendgrid" | "smtp";

const PROVIDER_LABELS: Record<ProviderType, string> = {
  resend: "Resend",
  sendgrid: "SendGrid",
  smtp: "SMTP (Gmail / custom)",
};

const isSmtp = (t: string) => t === "smtp";

const defaultForm = {
  provider_type: "resend" as ProviderType,
  display_name: "",
  // API key providers
  api_key: "",
  webhook_key: "",
  // SMTP
  host: "",
  port: "587",
  username: "",
  password: "",
  from_email: "",
  from_name: "",
  is_default: true,
};

type EditForm = { display_name: string; from_email: string; from_name: string };

export default function ProvidersPage() {
  const [connections, setConnections] = useState<ProviderConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(defaultForm);
  const [saving, setSaving] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<EditForm>({ display_name: "", from_email: "", from_name: "" });

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
      const credentials: Record<string, string> = isSmtp(form.provider_type)
        ? {
            host: form.host,
            port: form.port,
            username: form.username,
            password: form.password,
            ...(form.from_email && { from_email: form.from_email }),
            ...(form.from_name && { from_name: form.from_name }),
          }
        : {
            api_key: form.api_key,
            ...(form.webhook_key && { webhook_key: form.webhook_key }),
            ...(form.from_email && { from_email: form.from_email }),
            ...(form.from_name && { from_name: form.from_name }),
          };

      await apiFetch("/v1/provider-connections", {
        method: "POST",
        body: JSON.stringify({
          provider_type: form.provider_type,
          channel: "email",
          display_name: form.display_name,
          credentials,
          is_default: form.is_default,
        }),
      });
      setShowForm(false);
      setForm(defaultForm);
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleEdit = (c: ProviderConnection) => {
    setEditingId(c.id);
    setEditForm({ display_name: c.display_name, from_email: "", from_name: "" });
  };

  const handleUpdate = async () => {
    if (!editingId) return;
    setSaving(true);
    setError("");
    try {
      await apiFetch(`/v1/provider-connections/${editingId}`, {
        method: "PUT",
        body: JSON.stringify({
          display_name: editForm.display_name,
          credentials: {
            ...(editForm.from_email && { from_email: editForm.from_email }),
            ...(editForm.from_name && { from_name: editForm.from_name }),
          },
        }),
      });
      setEditingId(null);
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      await apiFetch(`/v1/provider-connections/${id}`, { method: "PATCH" });
      await load();
    } catch (e) {
      setError(String(e));
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

  const smtp = isSmtp(form.provider_type);
  const saveDisabled = saving || !form.display_name || !form.from_email || (smtp
    ? !form.host || !form.port || !form.username || !form.password
    : !form.api_key);

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

      {error && (isApiKeyError(error) ? <NoApiKey /> : <p className="text-sm text-red-600 mb-4">{error}</p>)}

      {!error && showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold mb-4">New provider connection</h2>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Provider</label>
              <select
                value={form.provider_type}
                onChange={(e) => setForm({ ...defaultForm, provider_type: e.target.value as ProviderType })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              >
                {(Object.entries(PROVIDER_LABELS) as [ProviderType, string][]).map(([value, label]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Display name</label>
              <input
                value={form.display_name}
                onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                placeholder={smtp ? "Gmail Workspace" : form.provider_type === "resend" ? "Resend Production" : "SendGrid Production"}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
              <p className="text-xs text-gray-400 mt-1">Only visible to you in this dashboard.</p>
            </div>

            {smtp ? (
              <>
                <div className="grid grid-cols-3 gap-3">
                  <div className="col-span-2">
                    <label className="block text-xs font-medium text-gray-700 mb-1">SMTP host</label>
                    <input
                      value={form.host}
                      onChange={(e) => setForm({ ...form, host: e.target.value })}
                      placeholder="smtp.gmail.com"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Port</label>
                    <input
                      value={form.port}
                      onChange={(e) => setForm({ ...form, port: e.target.value })}
                      placeholder="587"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Username</label>
                  <input
                    value={form.username}
                    onChange={(e) => setForm({ ...form, username: e.target.value })}
                    placeholder="you@gmail.com"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Password / App Password</label>
                  <input
                    type="password"
                    value={form.password}
                    onChange={(e) => setForm({ ...form, password: e.target.value })}
                    placeholder="Gmail App Password"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">From email</label>
                    <input
                      value={form.from_email}
                      onChange={(e) => setForm({ ...form, from_email: e.target.value })}
                      placeholder="you@gmail.com"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                    <p className="text-xs text-gray-400 mt-1">Address emails are sent from.</p>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">From name <span className="text-gray-400">(optional)</span></label>
                    <input
                      value={form.from_name}
                      onChange={(e) => setForm({ ...form, from_name: e.target.value })}
                      placeholder="Potoo"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                    <p className="text-xs text-gray-400 mt-1">Sender name shown in the inbox.</p>
                  </div>
                </div>
              </>
            ) : (
              <>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">API key</label>
                  <input
                    type="password"
                    value={form.api_key}
                    onChange={(e) => setForm({ ...form, api_key: e.target.value })}
                    placeholder={form.provider_type === "resend" ? "re_..." : "SG...."}
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
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">From email</label>
                    <input
                      value={form.from_email}
                      onChange={(e) => setForm({ ...form, from_email: e.target.value })}
                      placeholder="notifications@yourdomain.com"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                    <p className="text-xs text-gray-400 mt-1">Address emails are sent from.</p>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">From name <span className="text-gray-400">(optional)</span></label>
                    <input
                      value={form.from_name}
                      onChange={(e) => setForm({ ...form, from_name: e.target.value })}
                      placeholder="Potoo"
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                    <p className="text-xs text-gray-400 mt-1">Sender name shown in the inbox.</p>
                  </div>
                </div>
              </>
            )}

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
              disabled={saveDisabled}
              className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
            >
              {saving ? "Saving…" : "Save"}
            </button>
            <button
              onClick={() => { setShowForm(false); setForm(defaultForm); }}
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
            <div key={c.id} className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
              <div className="px-5 py-4 flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">{c.display_name}</p>
                  <p className="text-xs text-gray-500 mt-0.5">
                    {PROVIDER_LABELS[c.provider_type as ProviderType] ?? c.provider_type} · {c.channel}
                    {c.is_default && <span className="ml-2 text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">default</span>}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  {!c.is_default && (
                    <button
                      onClick={() => handleSetDefault(c.id)}
                      className="text-xs text-gray-500 hover:text-gray-900 transition-colors"
                    >
                      Set as default
                    </button>
                  )}
                  <button
                    onClick={() => editingId === c.id ? setEditingId(null) : handleEdit(c)}
                    className="text-gray-400 hover:text-gray-700 transition-colors"
                  >
                    <Pencil size={14} />
                  </button>
                  <button
                    onClick={() => handleDelete(c.id)}
                    disabled={c.is_default}
                    title={c.is_default ? "Set another provider as default before deleting" : undefined}
                    className="text-gray-400 hover:text-red-500 transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:text-gray-400"
                  >
                    <Trash2 size={15} />
                  </button>
                </div>
              </div>

              {editingId === c.id && (
                <div className="border-t border-gray-100 px-5 py-4 bg-gray-50 space-y-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Display name</label>
                    <input
                      value={editForm.display_name}
                      onChange={(e) => setEditForm({ ...editForm, display_name: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                    />
                    <p className="text-xs text-gray-400 mt-1">Only visible to you in this dashboard.</p>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">From email</label>
                      <input
                        value={editForm.from_email}
                        onChange={(e) => setEditForm({ ...editForm, from_email: e.target.value })}
                        placeholder="notifications@yourdomain.com"
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                      />
                      <p className="text-xs text-gray-400 mt-1">Address emails are sent from.</p>
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">From name <span className="text-gray-400">(optional)</span></label>
                      <input
                        value={editForm.from_name}
                        onChange={(e) => setEditForm({ ...editForm, from_name: e.target.value })}
                        placeholder="Potoo"
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                      />
                      <p className="text-xs text-gray-400 mt-1">Sender name shown in the inbox.</p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleUpdate}
                      disabled={saving || !editForm.display_name}
                      className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
                    >
                      {saving ? "Saving…" : "Save"}
                    </button>
                    <button
                      onClick={() => setEditingId(null)}
                      className="text-sm text-gray-500 px-3 py-2 hover:bg-gray-100 rounded-md transition-colors"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
