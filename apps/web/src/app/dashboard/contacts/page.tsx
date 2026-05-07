"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { Plus, Trash2, Pencil, X, Check, Tag, Users } from "lucide-react";

interface Contact {
  id: string;
  email: string;
  name?: string;
  phone?: string;
  tag?: string;
  created_at: string;
}

interface ContactForm {
  email: string;
  name: string;
  phone: string;
  tag: string;
}

const emptyForm: ContactForm = { email: "", name: "", phone: "", tag: "" };

export default function ContactsPage() {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [filterTag, setFilterTag] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<ContactForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<Omit<ContactForm, "email">>({ name: "", phone: "", tag: "" });

  const load = async (tag?: string) => {
    try {
      const url = tag ? `/v1/contacts?tag=${encodeURIComponent(tag)}` : "/v1/contacts";
      const [res, tagsRes] = await Promise.all([
        apiFetch<{ data: Contact[] }>(url),
        apiFetch<{ data: string[] }>("/v1/contacts/tags"),
      ]);
      setContacts(res.data);
      setTags(tagsRes.data);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(filterTag || undefined); }, [filterTag]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setFormError("");
    try {
      await apiFetch("/v1/contacts", {
        method: "POST",
        body: JSON.stringify({
          email: form.email,
          name: form.name || undefined,
          phone: form.phone || undefined,
          tag: form.tag || undefined,
        }),
      });
      setForm(emptyForm);
      setShowForm(false);
      await load(filterTag || undefined);
    } catch (e) {
      setFormError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this contact?")) return;
    try {
      await apiFetch(`/v1/contacts/${id}`, { method: "DELETE" });
      await load(filterTag || undefined);
    } catch (e) {
      setError(String(e));
    }
  };

  const startEdit = (c: Contact) => {
    setEditingId(c.id);
    setEditForm({ name: c.name ?? "", phone: c.phone ?? "", tag: c.tag ?? "" });
  };

  const handleUpdate = async (id: string) => {
    setSaving(true);
    try {
      await apiFetch(`/v1/contacts/${id}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: editForm.name || undefined,
          phone: editForm.phone || undefined,
          tag: editForm.tag || undefined,
        }),
      });
      setEditingId(null);
      await load(filterTag || undefined);
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const displayed = contacts;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold mb-0.5">Contacts</h1>
          <p className="text-sm text-gray-500">Save recipients to reuse when sending notifications.</p>
        </div>
        <button
          onClick={() => { setShowForm(true); setFormError(""); }}
          className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors"
        >
          <Plus size={14} /> New contact
        </button>
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {/* New contact form */}
      {showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold mb-3">New contact</h2>
          <form onSubmit={handleCreate} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Email <span className="text-red-400">*</span></label>
                <input
                  type="email" required value={form.email}
                  onChange={(e) => setForm({ ...form, email: e.target.value })}
                  placeholder="alice@example.com"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Name <span className="text-gray-400">(optional)</span></label>
                <input
                  type="text" value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="Alice Smith"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Phone <span className="text-gray-400">(optional)</span></label>
                <input
                  type="text" value={form.phone}
                  onChange={(e) => setForm({ ...form, phone: e.target.value })}
                  placeholder="+15551234567"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Tag <span className="text-gray-400">(optional)</span></label>
                <input
                  type="text" value={form.tag}
                  onChange={(e) => setForm({ ...form, tag: e.target.value })}
                  placeholder="vip, newsletter, …"
                  list="tag-suggestions"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                />
                <datalist id="tag-suggestions">
                  {tags.map((t) => <option key={t} value={t} />)}
                </datalist>
              </div>
            </div>
            {formError && <p className="text-xs text-red-600">{formError}</p>}
            <div className="flex gap-2 pt-1">
              <button type="submit" disabled={saving || !form.email}
                className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50">
                {saving ? "Saving…" : "Save contact"}
              </button>
              <button type="button" onClick={() => setShowForm(false)}
                className="text-sm text-gray-500 px-3 py-2 hover:bg-gray-100 rounded-md transition-colors">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Tag filter */}
      {tags.length > 0 && (
        <div className="flex items-center gap-2 mb-4 flex-wrap">
          <Tag size={13} className="text-gray-400 shrink-0" />
          <button
            onClick={() => setFilterTag("")}
            className={[
              "text-xs px-2.5 py-1 rounded-full border transition-colors",
              filterTag === "" ? "bg-gray-900 text-white border-gray-900" : "border-gray-200 text-gray-600 hover:border-gray-400",
            ].join(" ")}
          >
            All
          </button>
          {tags.map((t) => (
            <button key={t} onClick={() => setFilterTag(t === filterTag ? "" : t)}
              className={[
                "text-xs px-2.5 py-1 rounded-full border transition-colors",
                filterTag === t ? "bg-gray-900 text-white border-gray-900" : "border-gray-200 text-gray-600 hover:border-gray-400",
              ].join(" ")}
            >
              {t}
            </button>
          ))}
        </div>
      )}

      {/* Contact list */}
      {loading ? (
        <p className="text-sm text-gray-400">Loading…</p>
      ) : displayed.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Users size={32} className="text-gray-200 mb-3" />
          <p className="text-sm text-gray-500">
            {filterTag ? `No contacts with tag "${filterTag}".` : "No contacts yet. Add one to get started."}
          </p>
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-100">
                <th className="text-left text-xs font-medium text-gray-500 px-4 py-3">Email</th>
                <th className="text-left text-xs font-medium text-gray-500 px-4 py-3">Name</th>
                <th className="text-left text-xs font-medium text-gray-500 px-4 py-3">Phone</th>
                <th className="text-left text-xs font-medium text-gray-500 px-4 py-3">Tag</th>
                <th className="text-left text-xs font-medium text-gray-500 px-4 py-3">Added</th>
                <th className="w-16" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {displayed.map((c) => (
                <tr key={c.id} className="hover:bg-gray-50 transition-colors group">
                  {editingId === c.id ? (
                    <>
                      <td className="px-4 py-2 text-xs text-gray-500 font-mono">{c.email}</td>
                      <td className="px-4 py-2">
                        <input value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                          placeholder="Name" className="w-full border border-gray-300 rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-gray-900" />
                      </td>
                      <td className="px-4 py-2">
                        <input value={editForm.phone} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })}
                          placeholder="Phone" className="w-full border border-gray-300 rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-gray-900" />
                      </td>
                      <td className="px-4 py-2">
                        <input value={editForm.tag} onChange={(e) => setEditForm({ ...editForm, tag: e.target.value })}
                          placeholder="Tag" list="tag-suggestions" className="w-full border border-gray-300 rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-gray-900" />
                      </td>
                      <td className="px-4 py-2 text-xs text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
                      <td className="px-4 py-2">
                        <div className="flex items-center gap-1">
                          <button onClick={() => handleUpdate(c.id)} disabled={saving}
                            className="p-1 rounded hover:bg-green-50 text-gray-400 hover:text-green-600 transition-colors">
                            <Check size={13} />
                          </button>
                          <button onClick={() => setEditingId(null)}
                            className="p-1 rounded hover:bg-gray-100 text-gray-400 hover:text-gray-700 transition-colors">
                            <X size={13} />
                          </button>
                        </div>
                      </td>
                    </>
                  ) : (
                    <>
                      <td className="px-4 py-3 font-mono text-xs text-gray-700">{c.email}</td>
                      <td className="px-4 py-3 text-gray-700">{c.name ?? <span className="text-gray-300">—</span>}</td>
                      <td className="px-4 py-3 text-gray-500 text-xs">{c.phone ?? <span className="text-gray-300">—</span>}</td>
                      <td className="px-4 py-3">
                        {c.tag
                          ? <span className="text-xs bg-gray-100 text-gray-700 px-2 py-0.5 rounded-full">{c.tag}</span>
                          : <span className="text-gray-300">—</span>}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button onClick={() => startEdit(c)}
                            className="p-1 rounded hover:bg-gray-100 text-gray-400 hover:text-gray-700 transition-colors">
                            <Pencil size={13} />
                          </button>
                          <button onClick={() => handleDelete(c.id)}
                            className="p-1 rounded hover:bg-red-50 text-gray-400 hover:text-red-500 transition-colors">
                            <Trash2 size={13} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
