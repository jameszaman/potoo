"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { Template } from "@/lib/types";
import { Plus, Pencil } from "lucide-react";
import NoApiKey, { isApiKeyError } from "@/components/NoApiKey";

export default function TemplatesPage() {
  const router = useRouter();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    apiFetch<{ data: Template[] }>("/v1/templates")
      .then((res) => setTemplates(res.data))
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false));
  }, []);

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
          {templates.map((t) => (
            <div
              key={t.key}
              className="bg-white border border-gray-200 rounded-lg shadow-sm px-5 py-4 flex items-center justify-between"
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
              <button
                onClick={() => router.push(`/dashboard/templates/${t.key}/edit`)}
                className="flex items-center gap-1.5 text-sm text-gray-500 px-3 py-1.5 rounded-md hover:bg-gray-100 transition-colors"
              >
                <Pencil size={13} /> Edit
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
