"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { Delivery, DeliveryEvent } from "@/lib/types";
import clsx from "clsx";

const STATUS_COLORS: Record<string, string> = {
  queued: "bg-gray-100 text-gray-600",
  processing: "bg-blue-50 text-blue-700",
  sent: "bg-blue-100 text-blue-700",
  delivered: "bg-green-50 text-green-700",
  bounced: "bg-red-50 text-red-700",
  failed_temporary: "bg-yellow-50 text-yellow-700",
  failed_permanent: "bg-red-100 text-red-700",
  complained: "bg-orange-50 text-orange-700",
};

function DeliveryDetail({ deliveryId }: { deliveryId: string }) {
  const [delivery, setDelivery] = useState<Delivery | null>(null);
  const [events, setEvents] = useState<DeliveryEvent[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      apiFetch<Delivery>(`/v1/deliveries/${deliveryId}`),
      apiFetch<{ data: DeliveryEvent[] }>(`/v1/deliveries/${deliveryId}/events`),
    ])
      .then(([d, e]) => { setDelivery(d); setEvents(e.data); })
      .catch((e) => setError(String(e)));
  }, [deliveryId]);

  if (error) return <p className="text-sm text-red-600">{error}</p>;
  if (!delivery) return <p className="text-sm text-gray-400">Loading…</p>;

  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      <div className="px-5 py-4 border-b border-gray-100">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium font-mono">{delivery.id}</p>
          <span className={clsx("text-xs px-2 py-0.5 rounded font-medium", STATUS_COLORS[delivery.status] ?? "bg-gray-100 text-gray-600")}>
            {delivery.status}
          </span>
        </div>
        <dl className="mt-3 grid grid-cols-2 gap-x-6 gap-y-1.5 text-xs">
          <div><dt className="text-gray-400">Channel</dt><dd className="text-gray-700">{delivery.channel}</dd></div>
          <div><dt className="text-gray-400">Provider</dt><dd className="text-gray-700">{delivery.provider_type ?? "—"}</dd></div>
          <div><dt className="text-gray-400">Attempts</dt><dd className="text-gray-700">{delivery.attempt_count}</dd></div>
          <div><dt className="text-gray-400">Message ID</dt><dd className="text-gray-700 font-mono truncate">{delivery.provider_message_id ?? "—"}</dd></div>
          {delivery.last_error_message && (
            <div className="col-span-2"><dt className="text-gray-400">Last error</dt><dd className="text-red-600">{delivery.last_error_message}</dd></div>
          )}
        </dl>
      </div>
      <div className="px-5 py-4">
        <p className="text-xs font-medium text-gray-500 mb-3">Event timeline</p>
        {events.length === 0 ? (
          <p className="text-xs text-gray-400">No events yet.</p>
        ) : (
          <ol className="space-y-2">
            {events.map((e) => (
              <li key={e.id} className="flex items-start gap-3 text-xs">
                <span className="w-1.5 h-1.5 rounded-full bg-gray-400 mt-1.5 shrink-0" />
                <div>
                  <span className="font-medium">{e.event_type}</span>
                  {e.provider_type && <span className="text-gray-400 ml-1">via {e.provider_type}</span>}
                  <span className="text-gray-400 ml-2">{new Date(e.occurred_at).toLocaleString()}</span>
                </div>
              </li>
            ))}
          </ol>
        )}
      </div>
    </div>
  );
}

function DeliveriesContent() {
  const searchParams = useSearchParams();
  const selectedId = searchParams.get("id");
  const [lookupId, setLookupId] = useState(selectedId ?? "");
  const [activeId, setActiveId] = useState(selectedId ?? "");

  return (
    <div>
      <h1 className="text-xl font-semibold mb-0.5">Deliveries</h1>
      <p className="text-sm text-gray-500 mb-6">Look up a delivery by ID to see its status and event timeline.</p>
      <div className="flex gap-2 mb-6 max-w-lg">
        <input
          value={lookupId}
          onChange={(e) => setLookupId(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && setActiveId(lookupId)}
          placeholder="del_01HX1234"
          className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900"
        />
        <button
          onClick={() => setActiveId(lookupId)}
          disabled={!lookupId}
          className="bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
        >
          Look up
        </button>
      </div>
      {activeId && <DeliveryDetail key={activeId} deliveryId={activeId} />}
    </div>
  );
}

export default function DeliveriesPage() {
  return (
    <Suspense>
      <DeliveriesContent />
    </Suspense>
  );
}
