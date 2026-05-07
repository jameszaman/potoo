"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { Plus, Copy, Check, Trash2 } from "lucide-react";

interface InviteResponse {
  id: string;
  org_id: string;
  org_name?: string;
  token: string;
  used_at?: string;
  expires_at: string;
  deleted_at?: string;
  created_at: string;
}

function inviteStatus(inv: InviteResponse): { label: string; className: string } {
  if (inv.used_at) return { label: "Used", className: "bg-gray-100 text-gray-500" };
  if (new Date(inv.expires_at) < new Date()) return { label: "Expired", className: "bg-red-50 text-red-500" };
  return { label: "Active", className: "bg-green-50 text-green-700" };
}

export default function InvitesPage() {
  const [invites, setInvites] = useState<InviteResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const load = async () => {
    try {
      const res = await apiFetch<{ data: InviteResponse[] }>("/v1/invites");
      setInvites(res.data);
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
      await apiFetch("/v1/invites", { method: "POST" });
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this invite? The link will stop working immediately.")) return;
    try {
      await apiFetch(`/v1/invites/${id}`, { method: "DELETE" });
      await load();
    } catch (e) {
      setError(String(e));
    }
  };

  const inviteUrl = (token: string) =>
    `${window.location.origin}/invite/${token}`;

  const handleCopy = (inv: InviteResponse) => {
    navigator.clipboard.writeText(inviteUrl(inv.token));
    setCopiedId(inv.id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  if (loading) return <p className="text-sm text-gray-400">Loading…</p>;

  if (error) return <p className="text-sm text-red-600">{error}</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold mb-0.5">Invites</h1>
          <p className="text-sm text-gray-500">Create single-use invite links for your organization.</p>
        </div>
        <button
          onClick={handleCreate}
          disabled={saving}
          className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
        >
          <Plus size={14} /> {saving ? "Creating…" : "New invite"}
        </button>
      </div>

      {invites.length === 0 ? (
        <p className="text-sm text-gray-500">No invites yet. Create one to invite someone to your organization.</p>
      ) : (
        <div className="space-y-2">
          {invites.map((inv) => {
            const status = inviteStatus(inv);
            const isActive = !inv.used_at && new Date(inv.expires_at) >= new Date();
            return (
              <div key={inv.id} className="bg-white border border-gray-200 rounded-lg px-5 py-4 flex items-center justify-between shadow-sm">
                <div>
                  <p className="text-sm font-medium text-gray-800 font-mono">{inv.token.slice(0, 16)}…</p>
                  <p className="text-xs text-gray-400 mt-0.5">
                    Created {new Date(inv.created_at).toLocaleDateString()}
                    {" · "}
                    {inv.used_at
                      ? <span className="text-gray-500">Used {new Date(inv.used_at).toLocaleDateString()}</span>
                      : new Date(inv.expires_at) < new Date()
                        ? <span className="text-red-400">Expired {new Date(inv.expires_at).toLocaleDateString()}</span>
                        : <span>Expires {new Date(inv.expires_at).toLocaleDateString()}</span>
                    }
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${status.className}`}>{status.label}</span>
                  {isActive && (
                    <button onClick={() => handleCopy(inv)} className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-900 transition-colors">
                      {copiedId === inv.id ? <Check size={13} /> : <Copy size={13} />}
                      {copiedId === inv.id ? "Copied" : "Copy link"}
                    </button>
                  )}
                  {!inv.used_at && (
                    <button onClick={() => handleDelete(inv.id)} className="text-gray-400 hover:text-red-500 transition-colors">
                      <Trash2 size={14} />
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
