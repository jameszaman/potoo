"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { Template } from "@/lib/types";
import { Plus, ChevronDown, ChevronUp } from "lucide-react";
import NoApiKey, { isApiKeyError } from "@/components/NoApiKey";

export default function TemplatesPage() {
  const router = useRouter();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<string | null>(null);

  const [versionForm, setVersionForm] = useState<Record<string, { subject: string; html_body: string; text_body: string }>>({});
  const [saving, setSaving] = useState(false);

  const load = async () => {
    try {
      const res = await apiFetch<{ data: Template[] }>("/v1/templates");
      setTemplates(res.data);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCreateVersion = async (templateKey: string) => {
    const f = versionForm[templateKey];
    if (!f) return;
    setSaving(true);
    setError("");
    try {
      const version = await apiFetch<{ version_number: number }>(`/v1/templates/${templateKey}/versions`, {
        method: "POST",
        body: JSON.stringify({ subject: f.subject, html_body: f.html_body, text_body: f.text_body }),
      });
      await apiFetch(`/v1/templates/${templateKey}/activate`, {
        method: "POST",
        body: JSON.stringify({ version_number: version.version_number }),
      });
      setVersionForm((v) => ({ ...v, [templateKey]: { subject: "", html_body: "", text_body: "" } }));
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold mb-0.5">Templates</h1>
          <p className="text-sm text-gray-500">Create and manage notification templates.</p>
        </div>
        <button
          onClick={() => router.push("/dashboard/templates/new")}
          className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-3 py-2 rounded-md hover:bg-gray-700 transition-colors"
        >
          <Plus size={14} /> New template
        </button>
      </div>

      {error && (isApiKeyError(error) ? <NoApiKey /> : <p className="text-sm text-red-600 mb-4">{error}</p>)}

      {loading ? (
        <p className="text-sm text-gray-400">Loading…</p>
      ) : templates.length === 0 ? (
        <p className="text-sm text-gray-500">No templates yet.</p>
      ) : (
        <div className="space-y-2">
          {templates.map((t) => {
            const open = expanded === t.key;
            const vf = versionForm[t.key] ?? { subject: "", html_body: "", text_body: "" };
            return (
              <div key={t.key} className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
                <button
                  onClick={() => setExpanded(open ? null : t.key)}
                  className="w-full flex items-center justify-between px-5 py-4 text-left"
                >
                  <div>
                    <p className="text-sm font-medium">{t.name}</p>
                    <p className="text-xs text-gray-500 mt-0.5 flex items-center gap-2">
                      <span className="font-mono bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">{t.key}</span>
                      <span>{t.channel}</span>
                      {t.active_version != null ? (
                        <span className="bg-green-50 text-green-700 px-1.5 py-0.5 rounded">v{t.active_version} active</span>
                      ) : (
                        <span className="bg-yellow-50 text-yellow-700 px-1.5 py-0.5 rounded">no active version</span>
                      )}
                    </p>
                  </div>
                  {open ? <ChevronUp size={15} className="text-gray-400" /> : <ChevronDown size={15} className="text-gray-400" />}
                </button>

                {open && (
                  <div className="border-t border-gray-100 px-5 py-4">
                    <p className="text-xs font-medium text-gray-700 mb-3">Add version (auto-activates)</p>
                    <div className="space-y-3">
                      <div>
                        <label className="block text-xs text-gray-500 mb-1">Subject</label>
                        <input
                          value={vf.subject}
                          onChange={(e) => setVersionForm((v) => ({ ...v, [t.key]: { ...vf, subject: e.target.value } }))}
                          placeholder="Welcome to {{workspace_name}}"
                          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-gray-500 mb-1">HTML body</label>
                        <textarea
                          rows={4}
                          value={vf.html_body}
                          onChange={(e) => setVersionForm((v) => ({ ...v, [t.key]: { ...vf, html_body: e.target.value } }))}
                          placeholder="<p>Hi {{first_name}},</p>"
                          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-gray-500 mb-1">Text body <span className="text-gray-400">(optional)</span></label>
                        <textarea
                          rows={2}
                          value={vf.text_body}
                          onChange={(e) => setVersionForm((v) => ({ ...v, [t.key]: { ...vf, text_body: e.target.value } }))}
                          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900"
                        />
                      </div>
                    </div>
                    <button
                      onClick={() => handleCreateVersion(t.key)}
                      disabled={saving || !vf.subject || !vf.html_body}
                      className="mt-3 bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
                    >
                      {saving ? "Saving…" : "Save & activate"}
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
