"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { Template } from "@/lib/types";
import { CheckCircle, Clock, X, CalendarClock, Timer, Check } from "lucide-react";
import NoApiKey, { isApiKeyError } from "@/components/NoApiKey";

// ── Exact date/time picker ────────────────────────────────────────────────────

const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

function ExactPicker({
  value,
  onChange,
  onClose,
}: {
  value: Date | null;
  onChange: (d: Date) => void;
  onClose: () => void;
}) {
  const now = new Date();
  const initial = value ?? now;

  const [viewYear, setViewYear] = useState(initial.getFullYear());
  const [viewMonth, setViewMonth] = useState(initial.getMonth());
  const [selectedDate, setSelectedDate] = useState<Date | null>(value);
  const [timeInput, setTimeInput] = useState(() => {
    if (value) {
      return `${String(value.getHours()).padStart(2, "0")}:${String(value.getMinutes()).padStart(2, "0")}`;
    }
    return "09:00";
  });

  const parseTime = (s: string): { h: number; m: number } | null => {
    const match = s.match(/^(\d{1,2}):(\d{2})$/);
    if (!match) return null;
    const h = parseInt(match[1], 10);
    const m = parseInt(match[2], 10);
    if (h > 23 || m > 59) return null;
    return { h, m };
  };

  const emit = (d: Date | null, time: string) => {
    if (!d) return;
    const parsed = parseTime(time);
    if (!parsed) return;
    const out = new Date(d);
    out.setHours(parsed.h, parsed.m, 0, 0);
    onChange(out);
  };

  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();
  const firstDayOfWeek = new Date(viewYear, viewMonth, 1).getDay();

  const prevMonth = () => {
    if (viewMonth === 0) { setViewYear(y => y - 1); setViewMonth(11); }
    else setViewMonth(m => m - 1);
  };
  const nextMonth = () => {
    if (viewMonth === 11) { setViewYear(y => y + 1); setViewMonth(0); }
    else setViewMonth(m => m + 1);
  };

  const selectDay = (day: number) => {
    const d = new Date(viewYear, viewMonth, day);
    setSelectedDate(d);
    emit(d, timeInput);
    onClose();
  };

  const isPast = (day: number) => {
    const d = new Date(viewYear, viewMonth, day);
    const today = new Date(); today.setHours(0, 0, 0, 0);
    return d < today;
  };

  const isSelected = (day: number) =>
    selectedDate?.getFullYear() === viewYear &&
    selectedDate?.getMonth() === viewMonth &&
    selectedDate?.getDate() === day;

  const isToday = (day: number) => {
    const t = new Date();
    return t.getFullYear() === viewYear && t.getMonth() === viewMonth && t.getDate() === day;
  };

  const cells: (number | null)[] = [
    ...Array(firstDayOfWeek).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => i + 1),
  ];

  const timeValid = parseTime(timeInput) !== null;

  return (
    <div className="bg-white border border-gray-200 rounded-xl shadow-lg flex overflow-hidden">
      {/* Calendar */}
      <div className="p-4 w-64 shrink-0">
        <div className="flex items-center justify-between mb-3">
          <button type="button" onClick={prevMonth}
            className="w-7 h-7 flex items-center justify-center rounded hover:bg-gray-100 text-gray-500 hover:text-gray-800 transition-colors text-base leading-none">
            ‹
          </button>
          <span className="text-sm font-semibold text-gray-800">
            {MONTHS[viewMonth]} {viewYear}
          </span>
          <button type="button" onClick={nextMonth}
            className="w-7 h-7 flex items-center justify-center rounded hover:bg-gray-100 text-gray-500 hover:text-gray-800 transition-colors text-base leading-none">
            ›
          </button>
        </div>

        <div className="grid grid-cols-7 mb-1">
          {["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"].map((d) => (
            <div key={d} className="text-center text-xs text-gray-400 font-medium py-0.5">{d}</div>
          ))}
        </div>

        <div className="grid grid-cols-7 gap-y-0.5">
          {cells.map((day, i) => {
            if (!day) return <div key={i} />;
            const past = isPast(day);
            const sel = isSelected(day);
            const today = isToday(day);
            return (
              <button key={i} type="button" disabled={past} onClick={() => selectDay(day)}
                className={[
                  "text-xs h-7 w-7 mx-auto rounded-full flex items-center justify-center transition-colors font-medium",
                  past ? "text-gray-300 cursor-not-allowed" :
                  sel ? "bg-gray-900 text-white" :
                  today ? "ring-1 ring-gray-400 text-gray-800 hover:bg-gray-100" :
                  "text-gray-700 hover:bg-gray-100",
                ].join(" ")}
              >
                {day}
              </button>
            );
          })}
        </div>
      </div>

      {/* Time input */}
      <div className="border-l border-gray-100 flex flex-col w-36 shrink-0 p-4">
        <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Time</p>
        <input
          type="text"
          value={timeInput}
          onChange={(e) => {
            setTimeInput(e.target.value);
            if (selectedDate) emit(selectedDate, e.target.value);
          }}
          placeholder="HH:MM"
          className={[
            "w-full border rounded-md px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2",
            timeValid ? "border-gray-300 focus:ring-gray-900" : "border-red-300 focus:ring-red-400",
          ].join(" ")}
        />
        <p className="text-xs text-gray-400 mt-2">24-hour format</p>
        {!timeValid && timeInput !== "" && (
          <p className="text-xs text-red-500 mt-1">Use HH:MM (e.g. 14:30)</p>
        )}
        <p className="text-xs text-gray-400 mt-3 leading-relaxed">
          Set the time then pick a day to confirm.
        </p>
      </div>
    </div>
  );
}

// ── Relative time picker ──────────────────────────────────────────────────────

function RelativePicker({
  offsetMinutes,
  onChange,
}: {
  offsetMinutes: number;
  onChange: (minutes: number) => void;
}) {
  const totalMins = offsetMinutes;
  const days = Math.floor(totalMins / (60 * 24));
  const hours = Math.floor((totalMins % (60 * 24)) / 60);
  const mins = totalMins % 60;

  const update = (d: number, h: number, m: number) => {
    const total = d * 24 * 60 + h * 60 + m;
    onChange(Math.max(1, total));
  };

  const presets = [
    { label: "15 min", value: 15 },
    { label: "1 hour", value: 60 },
    { label: "3 hours", value: 180 },
    { label: "1 day", value: 60 * 24 },
  ];

  const previewLabel = () => {
    const parts: string[] = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (mins > 0) parts.push(`${mins}m`);
    return parts.length ? `in ${parts.join(" ")}` : "now";
  };

  return (
    <div className="bg-white border border-gray-200 rounded-xl shadow-lg p-5 w-80">
      {/* Presets */}
      <div className="flex gap-2 mb-5">
        {presets.map((p) => (
          <button
            key={p.value}
            type="button"
            onClick={() => onChange(p.value)}
            className={[
              "flex-1 text-xs font-medium py-1.5 rounded-md border transition-colors",
              offsetMinutes === p.value
                ? "bg-gray-900 text-white border-gray-900"
                : "border-gray-200 text-gray-600 hover:border-gray-400 hover:text-gray-900",
            ].join(" ")}
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* Sliders */}
      <div className="space-y-4">
        <SliderRow label="Days" value={days} max={30} onChange={(v) => update(v, hours, mins)} />
        <SliderRow label="Hours" value={hours} max={23} onChange={(v) => update(days, v, mins)} />
        <SliderRow label="Minutes" value={mins} max={59} onChange={(v) => update(days, hours, v)} />
      </div>

      {/* Preview */}
      <div className="mt-5 pt-4 border-t border-gray-100 flex items-center justify-between">
        <span className="text-xs text-gray-400">Sends</span>
        <span className="text-sm font-semibold text-gray-900">{previewLabel()}</span>
      </div>
    </div>
  );
}

function SliderRow({
  label,
  value,
  max,
  onChange,
}: {
  label: string;
  value: number;
  max: number;
  onChange: (v: number) => void;
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-1.5">
        <span className="text-xs font-medium text-gray-600">{label}</span>
        <span className="text-xs font-mono text-gray-900 w-6 text-right">{value}</span>
      </div>
      <input
        type="range"
        min={0}
        max={max}
        value={value}
        onChange={(e) => onChange(parseInt(e.target.value, 10))}
        className="w-full h-1.5 rounded-full appearance-none cursor-pointer accent-gray-900"
      />
    </div>
  );
}

// ── Schedule section ──────────────────────────────────────────────────────────

type ScheduleMode = "exact" | "relative";

function ScheduleSection({
  scheduledAt,
  onSchedule,
  onClear,
}: {
  scheduledAt: Date | null;
  onSchedule: (d: Date) => void;
  onClear: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState<ScheduleMode>("exact");
  const [offsetMinutes, setOffsetMinutes] = useState(60);

  const applyRelative = (minutes: number) => {
    setOffsetMinutes(minutes);
    const d = new Date(Date.now() + minutes * 60 * 1000);
    onSchedule(d);
  };

  return (
    <div className="border-t border-gray-100 pt-4">
      <div className="flex items-center justify-between mb-2">
        <label className="text-xs font-medium text-gray-700">Schedule for later <span className="text-gray-400 font-normal">(optional)</span></label>
        {scheduledAt && (
          <button
            type="button"
            onClick={() => { onClear(); setOpen(false); }}
            className="flex items-center gap-1 text-xs text-gray-400 hover:text-red-500 transition-colors"
          >
            <X size={11} /> Clear
          </button>
        )}
      </div>

      {/* Collapsed summary */}
      {scheduledAt && !open ? (
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="w-full flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-md text-sm text-gray-800 hover:border-gray-400 transition-colors bg-white"
        >
          <Clock size={14} className="text-gray-400 shrink-0" />
          {scheduledAt.toLocaleString(undefined, {
            weekday: "short", month: "short", day: "numeric",
            hour: "2-digit", minute: "2-digit",
          })}
        </button>
      ) : !open ? (
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="flex items-center gap-2 text-xs text-gray-500 hover:text-gray-800 transition-colors border border-dashed border-gray-300 rounded-md px-3 py-2 w-full hover:border-gray-400"
        >
          <CalendarClock size={13} />
          Pick a date and time
        </button>
      ) : null}

      {/* Expanded picker */}
      {open && (
        <div>
          {/* Mode tabs */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex gap-1 p-1 bg-gray-100 rounded-lg">
              <button
                type="button"
                onClick={() => setMode("exact")}
                className={[
                  "flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-md transition-colors",
                  mode === "exact" ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
                ].join(" ")}
              >
                <CalendarClock size={12} /> Exact time
              </button>
              <button
                type="button"
                onClick={() => setMode("relative")}
                className={[
                  "flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-md transition-colors",
                  mode === "relative" ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
                ].join(" ")}
              >
                <Timer size={12} /> Relative
              </button>
            </div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => setOpen(false)}
                title="Confirm"
                className="w-7 h-7 flex items-center justify-center rounded hover:bg-green-50 text-gray-400 hover:text-green-600 transition-colors"
              >
                <Check size={14} />
              </button>
              <button
                type="button"
                onClick={() => { onClear(); setOpen(false); }}
                title="Cancel"
                className="w-7 h-7 flex items-center justify-center rounded hover:bg-red-50 text-gray-400 hover:text-red-500 transition-colors"
              >
                <X size={14} />
              </button>
            </div>
          </div>

          {mode === "exact" ? (
            <ExactPicker
              value={scheduledAt}
              onChange={(d) => { onSchedule(d); }}
              onClose={() => setOpen(false)}
            />
          ) : (
            <RelativePicker
              offsetMinutes={offsetMinutes}
              onChange={(mins) => applyRelative(mins)}
            />
          )}
        </div>
      )}
    </div>
  );
}

// ── Send page ─────────────────────────────────────────────────────────────────

export default function SendPage() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [form, setForm] = useState({
    template_key: "",
    recipient_email: "",
    data: "",
    cc: "",
    bcc: "",
  });
  const [showCcBcc, setShowCcBcc] = useState(false);
  const [scheduledAt, setScheduledAt] = useState<Date | null>(null);
  const [sending, setSending] = useState(false);
  const [result, setResult] = useState<{
    notification_id: string;
    delivery_id: string;
    scheduled_at?: string;
  } | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    apiFetch<{ data: Template[] }>("/v1/templates")
      .then((res) => setTemplates(res.data))
      .catch(() => {});
  }, []);

  const selectedTemplate = templates.find((t) => t.key === form.template_key);

  const handleSend = async () => {
    setSending(true);
    setError("");
    setResult(null);
    try {
      let templateData: Record<string, unknown> = {};
      if (form.data.trim()) templateData = JSON.parse(form.data);

      const parseEmails = (s: string) =>
        s.split(",").map((e) => e.trim()).filter(Boolean);

      const body: Record<string, unknown> = {
        channel: "email",
        recipient: { email: form.recipient_email },
        template: { key: form.template_key, data: templateData },
      };
      if (form.cc.trim()) body.cc = parseEmails(form.cc);
      if (form.bcc.trim()) body.bcc = parseEmails(form.bcc);
      if (scheduledAt) body.scheduled_at = scheduledAt.toISOString();

      const res = await apiFetch<{ notification_id: string; delivery_id: string }>(
        "/v1/notifications",
        { method: "POST", body: JSON.stringify(body) }
      );
      setResult({ ...res, scheduled_at: scheduledAt?.toISOString() });
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

      {error && (isApiKeyError(error) ? <NoApiKey /> : <p className="text-sm text-red-600 mb-4">{error}</p>)}

      {result ? (
        <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm max-w-lg">
          <div className="flex items-center gap-2 mb-4">
            {result.scheduled_at ? (
              <>
                <Clock size={18} className="text-blue-500" />
                <span className="text-sm font-medium text-blue-700">Notification scheduled</span>
              </>
            ) : (
              <>
                <CheckCircle size={18} className="text-green-500" />
                <span className="text-sm font-medium text-green-700">Notification queued</span>
              </>
            )}
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
            {result.scheduled_at && (
              <div className="flex gap-2">
                <dt className="text-gray-500 w-36 shrink-0">Scheduled for</dt>
                <dd className="text-xs text-gray-700">{new Date(result.scheduled_at).toLocaleString()}</dd>
              </div>
            )}
          </dl>
          <div className="mt-4 flex gap-3">
            <a
              href={`/dashboard/deliveries?id=${result.delivery_id}`}
              className="text-sm text-gray-900 underline underline-offset-2"
            >
              View delivery →
            </a>
            <button onClick={() => setResult(null)} className="text-sm text-gray-500 hover:text-gray-900">
              Send another
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm max-w-lg">
          <div className="space-y-4">
            {/* Template */}
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Template</label>
              {templates.length === 0 ? (
                <p className="text-xs text-gray-400 py-2">
                  No templates yet.{" "}
                  <a href="/dashboard/templates" className="underline text-gray-600">Create one first →</a>
                </p>
              ) : (
                <>
                  <select
                    value={form.template_key}
                    onChange={(e) => setForm({ ...form, template_key: e.target.value })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 bg-white"
                  >
                    <option value="">Select a template…</option>
                    {templates.map((t) => (
                      <option key={t.key} value={t.key}>{t.name}</option>
                    ))}
                  </select>
                  {selectedTemplate && (
                    <p className="mt-1 text-xs text-gray-400">
                      Key: <span className="font-mono">{selectedTemplate.key}</span>
                      {selectedTemplate.active_version != null
                        ? ` · v${selectedTemplate.active_version} active`
                        : " · no active version"}
                      </p>
                  )}
                </>
              )}
            </div>

            {/* Recipient */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="text-xs font-medium text-gray-700">Recipient email</label>
                {!showCcBcc && (
                  <button
                    type="button"
                    onClick={() => setShowCcBcc(true)}
                    className="text-xs text-gray-400 hover:text-gray-700 transition-colors"
                  >
                    + CC / BCC
                  </button>
                )}
              </div>
              <input
                type="email"
                value={form.recipient_email}
                onChange={(e) => setForm({ ...form, recipient_email: e.target.value })}
                placeholder="user@example.com"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              />
            </div>

            {/* CC / BCC */}
            {showCcBcc && (
              <div className="space-y-3">
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">
                    CC <span className="text-gray-400 font-normal">(comma-separated)</span>
                  </label>
                  <input
                    type="text"
                    value={form.cc}
                    onChange={(e) => setForm({ ...form, cc: e.target.value })}
                    placeholder="alice@example.com, bob@example.com"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                  />
                </div>
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <label className="text-xs font-medium text-gray-700">
                      BCC <span className="text-gray-400 font-normal">(comma-separated)</span>
                    </label>
                    <button
                      type="button"
                      onClick={() => { setShowCcBcc(false); setForm({ ...form, cc: "", bcc: "" }); }}
                      className="text-xs text-gray-400 hover:text-red-500 transition-colors"
                    >
                      Remove
                    </button>
                  </div>
                  <input
                    type="text"
                    value={form.bcc}
                    onChange={(e) => setForm({ ...form, bcc: e.target.value })}
                    placeholder="audit@example.com"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                  />
                </div>
              </div>
            )}

            {/* Template data */}
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

            <ScheduleSection
              scheduledAt={scheduledAt}
              onSchedule={setScheduledAt}
              onClear={() => setScheduledAt(null)}
            />
          </div>

          <button
            onClick={handleSend}
            disabled={sending || !form.template_key || !form.recipient_email}
            className="mt-5 bg-gray-900 text-white text-sm font-medium px-4 py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
          >
            {sending ? "Sending…" : scheduledAt ? "Schedule notification" : "Send notification"}
          </button>
        </div>
      )}
    </div>
  );
}
