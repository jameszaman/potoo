"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { ArrowLeft } from "lucide-react";

export default function NewTemplatePage() {
  const router = useRouter();
  const [form, setForm] = useState({ name: "", key: "", channel: "email" });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const handleNameChange = (name: string) => {
    const autoKey = name.toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_|_$/g, "");
    setForm((f) => ({ ...f, name, key: autoKey }));
  };

  const handleSubmit = async () => {
    setSaving(true);
    setError("");
    try {
      await apiFetch("/v1/templates", { method: "POST", body: JSON.stringify(form) });
      router.push(`/dashboard/templates/${form.key}/edit`);
    } catch (e) {
      setError(String(e));
      setSaving(false);
    }
  };

  return (
    <div className="max-w-lg">
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => router.push("/dashboard/templates")}
          className="text-gray-400 hover:text-gray-700 transition-colors"
        >
          <ArrowLeft size={18} />
        </button>
        <div>
          <h1 className="text-xl font-semibold">New template</h1>
          <p className="text-sm text-gray-500">Templates define the content sent to recipients.</p>
        </div>
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm space-y-4">
        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">
            Name <span className="text-gray-400 font-normal">— human-readable label shown in the dashboard</span>
          </label>
          <input
            autoFocus
            value={form.name}
            onChange={(e) => handleNameChange(e.target.value)}
            placeholder="Welcome Email"
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">
            Key <span className="text-gray-400 font-normal">— used in API calls, cannot be changed later</span>
          </label>
          <input
            value={form.key}
            onChange={(e) => setForm((f) => ({ ...f, key: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, "") }))}
            placeholder="welcome_email"
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          <p className="mt-1 text-xs text-gray-400">Lowercase letters, numbers and underscores only.</p>
        </div>

        <div className="flex gap-2 pt-1">
          <button
            onClick={handleSubmit}
            disabled={saving || !form.name || !form.key}
            className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
          >
            {saving ? "Creating…" : "Create template"}
          </button>
          <button
            onClick={() => router.push("/dashboard/templates")}
            className="text-sm text-gray-500 px-4 py-2 rounded-md hover:bg-gray-100"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
