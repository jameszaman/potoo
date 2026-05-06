"use client";

import { useState } from "react";
import { apiFetch } from "@/lib/api";
import ApiKeyGate, { useApiKey } from "@/components/ApiKeyGate";
import { CheckCircle } from "lucide-react";

function SendContent() {
  const { apiKey } = useApiKey();
  const [form, setForm] = useState({
    template_key: "",
    recipient_email: "",
    data: "",
  });
  const [sending, setSending] = useState(false);
  const [result, setResult] = useState<{ notification_id: string; delivery_id: string } | null>(null);
  const [error, setError] = useState("");

  const handleSend = async () => {
    setSending(true);
    setError("");
    setResult(null);
    try {
      let templateData: Record<string, unknown> = {};
      if (form.data.trim()) {
        templateData = JSON.parse(form.data);
      }

      const res = await apiFetch<{ notification_id: string; delivery_id: string }>(
        "/v1/notifications",
        apiKey,
        {
          method: "POST",
          body: JSON.stringify({
            channel: "email",
            recipient: { email: form.recipient_email },
            template: { key: form.template_key, data: templateData },
          }),
        }
      );
      setResult(res);
    } catch (e) {
      setError(String(e));
    } finally {
      setSending(false);
    }
  };

  return (
    <div>
      <h1 className="text-xl font-semibold mb-0.5">Send</h1>
      <p className="text-sm text-gray-500 mb-6">Send a test notification.</p>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {result ? (
        <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <CheckCircle size={18} className="text-green-500" />
            <span className="text-sm font-medium text-green-700">Notification queued</span>
          </div>
          <dl className="space-y-2 text-sm">
            <div className="flex gap-2">
              <dt className="text-gray-500 w-36 shrink-0">Notification ID</dt>
              <dd className="font-mono text-xs bg-gray-50 px-2 py-0.5 rounded">{result.notification_id}</dd>
            </div>
            <div className="flex gap-2">
              <dt className="text-gray-500 w-36 shrink-0">Delivery ID</dt>
              <dd className="font-mono text-xs bg-gray-50 px-2 py-0.5 rounded">{result.delivery_id}</dd>
            </div>
          </dl>
          <div className="mt-4 flex gap-2">
            <a
              href={`/dashboard/deliveries?id=${result.delivery_id}`}
              className="text-sm text-gray-900 underline underline-offset-2"
            >
              View delivery →
            </a>
            <button
              onClick={() => setResult(null)}
              className="text-sm text-gray-500 hover:text-gray-900"
            >
              Send another
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm max-w-lg">
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Template key</label>
              <input
                value={form.template_key}
                onChange={(e) => setForm({ ...form, template_key: e.target.value })}
                placeholder="welcome_email"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Recipient email</label>
              <input
                type="email"
                value={form.recipient_email}
                onChange={(e) => setForm({ ...form, recipient_email: e.target.value })}
                placeholder="user@example.com"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">
                Template data <span className="text-gray-400">(JSON, optional)</span>
              </label>
              <textarea
                rows={3}
                value={form.data}
                onChange={(e) => setForm({ ...form, data: e.target.value })}
                placeholder={`{"first_name": "James"}`}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>
          </div>
          <button
            onClick={handleSend}
            disabled={sending || !form.template_key || !form.recipient_email}
            className="mt-5 bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
          >
            {sending ? "Sending…" : "Send notification"}
          </button>
        </div>
      )}
    </div>
  );
}

export default function SendPage() {
  return <ApiKeyGate><SendContent /></ApiKeyGate>;
}
