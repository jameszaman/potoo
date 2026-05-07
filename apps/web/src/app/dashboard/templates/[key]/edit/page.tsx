"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { Template } from "@/lib/types";
import { ArrowLeft, Eye, EyeOff, Save } from "lucide-react";
import TemplateEditor, { Block, blocksToHtml, blocksToText } from "@/components/TemplateEditor";

export default function EditTemplatePage() {
  const { key } = useParams<{ key: string }>();
  const router = useRouter();

  const [template, setTemplate] = useState<Template | null>(null);
  const [loadError, setLoadError] = useState("");
  const [loading, setLoading] = useState(true);

  const [initialBlocks, setInitialBlocks] = useState<Block[] | undefined>(undefined);
  const [initialSubject, setInitialSubject] = useState("");

  const blocksRef = useRef<Block[]>([]);
  const [subject, setSubject] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [saved, setSaved] = useState(false);
  const [preview, setPreview] = useState(false);
  const [previewHtml, setPreviewHtml] = useState("");

  useEffect(() => {
    apiFetch<Template & {
      active_version_subject?: string;
      active_version_html_body?: string;
      active_version_blocks?: string;
    }>(`/v1/templates/${key}`)
      .then((t) => {
        setTemplate(t);
        const subj = t.active_version_subject ?? "";
        setInitialSubject(subj);
        setSubject(subj);
        if (t.active_version_blocks) {
          try {
            // Migrate old `spans` shape → `html`, and old `items: InlineSpan[][]` → `items: string[]`
            const migrate = (b: Block): Block => {
              const anyB = b as Block & { spans?: { text: string; marks?: string[]; href?: string }[]; };
              if (anyB.spans) {
                const html = anyB.spans.map((s) => {
                  let t = s.text;
                  if (s.marks?.includes("bold"))          t = `<strong>${t}</strong>`;
                  if (s.marks?.includes("italic"))        t = `<em>${t}</em>`;
                  if (s.marks?.includes("underline"))     t = `<u>${t}</u>`;
                  if (s.marks?.includes("strikethrough")) t = `<del>${t}</del>`;
                  if (s.href) t = `<a href="${s.href}">${t}</a>`;
                  return t;
                }).join("");
                return { ...b, html, spans: undefined } as Block;
              }
              if (b.items && b.items.length > 0 && typeof b.items[0] !== "string") {
                const items = (b.items as unknown as { text: string }[][]).map((row) => row.map((s) => s.text).join(""));
                return { ...b, items };
              }
              return b;
            };
            const parsed = (JSON.parse(t.active_version_blocks) as Block[]).map(migrate);
            setInitialBlocks(parsed);
            blocksRef.current = parsed;
          } catch {
            setInitialBlocks(undefined);
          }
        }
      })
      .catch((e) => setLoadError(String(e)))
      .finally(() => setLoading(false));
  }, [key]);

  const handleEditorChange = useCallback((blocks: Block[], _html: string, _text: string) => {
    blocksRef.current = blocks;
  }, []);

  const handleSave = async () => {
    setSaving(true);
    setSaveError("");
    setSaved(false);
    try {
      const blocks = blocksRef.current;
      const html = blocksToHtml(blocks);
      const text = blocksToText(blocks);
      const editorBlocks = JSON.stringify(blocks);

      const version = await apiFetch<{ version_number: number }>(
        `/v1/templates/${key}/versions`,
        {
          method: "POST",
          body: JSON.stringify({
            subject,
            html_body: html,
            text_body: text,
            editor_blocks: editorBlocks,
          }),
        }
      );
      await apiFetch(`/v1/templates/${key}/activate`, {
        method: "POST",
        body: JSON.stringify({ version_number: version.version_number }),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (e) {
      setSaveError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const togglePreview = () => {
    if (!preview) setPreviewHtml(blocksToHtml(blocksRef.current));
    setPreview((v) => !v);
  };

  if (loadError) {
    return (
      <div className="p-8">
        <p className="text-sm text-red-600">{loadError}</p>
        <button onClick={() => router.push("/dashboard/templates")} className="mt-4 text-sm text-gray-500 underline">
          Back to templates
        </button>
      </div>
    );
  }

  if (loading) {
    return <div className="p-8 text-sm text-gray-400">Loading…</div>;
  }

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Top bar */}
      <div className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between sticky top-0 z-40">
        <div className="flex items-center gap-3">
          <button
            onClick={() => router.push("/dashboard/templates")}
            className="text-gray-400 hover:text-gray-700 transition-colors"
          >
            <ArrowLeft size={18} />
          </button>
          <div>
            <p className="text-sm font-semibold text-gray-900">{template?.name ?? key}</p>
            <p className="text-xs text-gray-400 font-mono">{key}</p>
          </div>
          {template?.active_version != null && (
            <span className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">
              v{template.active_version} active
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {saved && <span className="text-xs text-green-600 font-medium">Saved &amp; activated</span>}
          {saveError && <span className="text-xs text-red-600">{saveError}</span>}
          {!subject && !saved && <span className="text-xs text-amber-500">Subject required to save</span>}
          <button
            onClick={togglePreview}
            className="flex items-center gap-1.5 text-sm text-gray-500 px-3 py-1.5 rounded-md hover:bg-gray-100 transition-colors"
          >
            {preview ? <EyeOff size={14} /> : <Eye size={14} />}
            {preview ? "Edit" : "Preview"}
          </button>
          <button
            onClick={handleSave}
            disabled={saving || !subject}
            className="flex items-center gap-1.5 bg-gray-900 text-white text-sm font-medium px-4 py-1.5 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Save size={13} />
            {saving ? "Saving…" : "Save & activate"}
          </button>
        </div>
      </div>

      {/* Subject line */}
      <div className={`bg-white border-b px-8 py-3 ${!subject ? "border-amber-200" : "border-gray-100"}`}>
        <div className="max-w-2xl mx-auto flex items-center gap-3">
          <span className={`text-xs w-14 shrink-0 ${!subject ? "text-amber-500" : "text-gray-400"}`}>Subject</span>
          <input
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            placeholder="Required — supports {{variables}}"
            className="flex-1 text-sm outline-none bg-transparent placeholder:text-amber-300"
          />
        </div>
      </div>

      {/* Body */}
      <div className="flex-1 px-8 py-8">
        <div className="max-w-2xl mx-auto">
          {/* Preview pane — only rendered when active */}
          {preview && (
            <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
              <div className="bg-gray-50 border-b border-gray-200 px-4 py-2 text-xs text-gray-400 flex items-center gap-2">
                <Eye size={12} />
                Email preview
                <span className="ml-auto font-mono text-gray-300">Subject: {subject || "(none)"}</span>
              </div>
              <iframe
                srcDoc={previewHtml}
                className="w-full"
                style={{ height: "600px", border: "none" }}
                title="Email preview"
              />
            </div>
          )}
          {/* Editor — always mounted, hidden during preview so DOM state is preserved */}
          <div className={`bg-white rounded-xl border border-gray-200 shadow-sm px-8 py-8 min-h-125 ${preview ? "hidden" : ""}`}>
            {/* key= forces a full remount only when initial data first arrives */}
            <TemplateEditor
              key={initialSubject + String(!!initialBlocks)}
              initialBlocks={initialBlocks}
              onChange={handleEditorChange}
            />
          </div>
        </div>
      </div>

      {!preview && (
        <div className="px-8 pb-6">
          <div className="max-w-2xl mx-auto">
            <p className="text-xs text-gray-400">
              Hover a block to change its type or delete it.
              Click <strong>+</strong> to insert a block below.
              Select text to apply bold, italic, underline, or a link.
              Use <code className="bg-gray-100 px-1 rounded font-mono">{"{{variable}}"}</code> for dynamic content.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
