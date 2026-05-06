"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { listMyOrgs, OrgSummary } from "@/lib/auth";
import { Plus, Copy, Check } from "lucide-react";

interface InviteResponse {
  id: string;
  org_id: string;
  org_name?: string;
  token: string;
  used_at?: string;
  expires_at: string;
  created_at: string;
}

function inviteStatus(inv: InviteResponse): { label: string; className: string } {
  if (inv.used_at) return { label: "Used", className: "bg-gray-100 text-gray-500" };
  if (new Date(inv.expires_at) < new Date()) return { label: "Expired", className: "bg-red-50 text-red-500" };
  return { label: "Active", className: "bg-green-50 text-green-700" };
}

export default function InvitesPage() {
  const [invites, setInvites] = useState<InviteResponse[]>([]);
  const [orgs, setOrgs] = useState<OrgSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [saving, setSaving] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const load = async () => {
    try {
      const [invRes, orgRes] = await Promise.all([
        apiFetch<{ data: InviteResponse[] }>("/v1/platform/invites"),
        listMyOrgs(),
      ]);
      setInvites(invRes.data);
      const customerOrgs = orgRes.filter((o) => o.type === "customer");
      setOrgs(customerOrgs);
      if (customerOrgs.length > 0) setSelectedOrgId(customerOrgs[0].id);
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
      await apiFetch("/v1/platform/invites", {
        method: "POST",
        body: JSON.stringify({ org_id: selectedOrgId }),
      });
      setShowForm(false);
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
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
          <p className="text-sm text-gray-500">Create single-use invite links for customer organizations.</p>
        </div>
        {orgs.length > 0 && (
          <button
            onClick={() => setShowForm(true)}
            className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors"
          >
            <Plus size={14} /> New invite
          </button>
        )}
      </div>

      {showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 mb-6 shadow-sm">
          <h2 className="text-sm font-semibold mb-3">New invite</h2>
          <form onSubmit={handleCreate} className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 mb-1">Organization</label>
              <select
                value={selectedOrgId}
                onChange={(e) => setSelectedOrgId(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              >
                {orgs.map((o) => (
                  <option key={o.id} value={o.id}>{o.name}</option>
                ))}
              </select>
            </div>
            <button
              type="submit"
              disabled={saving || !selectedOrgId}
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

      {orgs.length === 0 && invites.length === 0 ? (
        <p className="text-sm text-gray-500">No customer organizations yet. Create one before sending invites.</p>
      ) : invites.length === 0 ? (
        <p className="text-sm text-gray-500">No invites yet. Create one to invite someone to a customer organization.</p>
      ) : (
        <div className="space-y-2">
          {invites.map((inv) => {
            const status = inviteStatus(inv);
            const isActive = !inv.used_at && new Date(inv.expires_at) >= new Date();
            return (
              <div key={inv.id} className="bg-white border border-gray-200 rounded-lg px-5 py-4 flex items-center justify-between shadow-sm">
                <div>
                  <p className="text-sm font-medium">{inv.org_name ?? inv.org_id}</p>
                  <p className="text-xs text-gray-400 mt-0.5">
                    Expires {new Date(inv.expires_at).toLocaleDateString()}
                    {inv.used_at && <> · Used {new Date(inv.used_at).toLocaleDateString()}</>}
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
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
