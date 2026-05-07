"use client";

/**
 * Block-based rich text editor that serializes to email-safe HTML.
 *
 * Design rules that prevent the contentEditable cursor-jump bug:
 *  1. dangerouslySetInnerHTML is only used on initial mount (via key= trick).
 *  2. After mount the DOM owns the text content; React only controls block
 *     structure (add / remove / reorder blocks).
 *  3. Inline marks (bold/italic/link) are applied imperatively via execCommand
 *     or manual range manipulation, then read back on the next input event.
 */

import {
  useRef,
  useState,
  useEffect,
  useCallback,
  useImperativeHandle,
  forwardRef,
  KeyboardEvent,
} from "react";
import {
  Bold,
  Italic,
  Underline as UnderlineIcon,
  Strikethrough,
  Link,
  Link2Off,
  Pencil,
  ImageIcon,
  Minus,
  List,
  ListOrdered,
  ChevronDown,
  Square,
  Type,
  Plus,
  X,
  Quote,
  AlignLeft,
  AlignCenter,
  AlignRight,
} from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type InlineMark = "bold" | "italic" | "underline" | "strikethrough";

export type BlockType =
  | "paragraph"
  | "h1" | "h2" | "h3" | "h4" | "h5" | "h6"
  | "bullet_list"
  | "ordered_list"
  | "blockquote"
  | "divider"
  | "image"
  | "button";

export type BlockAlign = "left" | "center" | "right";

export interface Block {
  id: string;
  type: BlockType;
  align?: BlockAlign;   // text alignment (default: left)
  html?: string;        // innerHTML for text-based blocks (paragraph, headings, blockquote)
  items?: string[];     // innerHTML per list item
  src?: string;         // image
  alt?: string;         // image
  label?: string;       // button
  href?: string;        // button
}

// ---------------------------------------------------------------------------
// ID generator
// ---------------------------------------------------------------------------

let _id = 0;
function uid() { return `b${++_id}`; }

// ---------------------------------------------------------------------------
// HTML serialiser (email-safe, inline styles)
// ---------------------------------------------------------------------------

// Strip all HTML tags, returning plain text content.
function stripTags(html: string): string {
  return html.replace(/<[^>]*>/g, "");
}

export function blocksToHtml(blocks: Block[]): string {
  const rows: string[] = [];
  rows.push(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body style="margin:0;padding:0;background:#ffffff;">`);
  rows.push(`<div style="font-family:sans-serif;font-size:15px;line-height:1.6;color:#111827;max-width:600px;margin:0 auto;padding:24px;">`);
  for (const b of blocks) {
    const ta = b.align && b.align !== "left" ? `text-align:${b.align};` : "";
    const inner = b.html ?? "";
    switch (b.type) {
      case "h1": rows.push(`<h1 style="font-size:28px;font-weight:700;margin:0 0 12px;${ta}">${inner}</h1>`); break;
      case "h2": rows.push(`<h2 style="font-size:24px;font-weight:700;margin:0 0 12px;${ta}">${inner}</h2>`); break;
      case "h3": rows.push(`<h3 style="font-size:20px;font-weight:600;margin:0 0 10px;${ta}">${inner}</h3>`); break;
      case "h4": rows.push(`<h4 style="font-size:18px;font-weight:600;margin:0 0 10px;${ta}">${inner}</h4>`); break;
      case "h5": rows.push(`<h5 style="font-size:16px;font-weight:600;margin:0 0 8px;${ta}">${inner}</h5>`); break;
      case "h6": rows.push(`<h6 style="font-size:14px;font-weight:600;margin:0 0 8px;text-transform:uppercase;letter-spacing:.05em;${ta}">${inner}</h6>`); break;
      case "paragraph": rows.push(`<p style="margin:0 0 12px;${ta}">${inner || "&nbsp;"}</p>`); break;
      case "bullet_list": {
        const li = (b.items ?? []).map((s) => `<li style="margin-bottom:4px;">${s}</li>`).join("");
        rows.push(`<ul style="margin:0 0 12px;padding-left:20px;${ta}">${li}</ul>`);
        break;
      }
      case "ordered_list": {
        const li = (b.items ?? []).map((s) => `<li style="margin-bottom:4px;">${s}</li>`).join("");
        rows.push(`<ol style="margin:0 0 12px;padding-left:20px;${ta}">${li}</ol>`);
        break;
      }
      case "blockquote": rows.push(`<blockquote style="border-left:3px solid #d1d5db;margin:0 0 12px;padding:4px 12px;color:#6b7280;${ta}">${inner}</blockquote>`); break;
      case "divider": rows.push(`<hr style="border:none;border-top:1px solid #e5e7eb;margin:16px 0;" />`); break;
      case "image": {
        const src = b.src?.trim();
        if (src) {
          const imgAlign = b.align === "center" ? "margin:0 auto 12px;" : b.align === "right" ? "margin:0 0 12px auto;" : "margin:0 0 12px;";
          rows.push(`<img src="${src}" alt="${b.alt ?? ""}" style="max-width:100%;display:block;${imgAlign}" />`);
        }
        break;
      }
      case "button": {
        const btnAlign = b.align === "center" ? "margin:0 auto 16px;" : b.align === "right" ? "margin:0 0 16px auto;" : "margin:0 0 16px;";
        rows.push(`<table style="${btnAlign}" cellpadding="0" cellspacing="0"><tr><td style="background:#111827;border-radius:6px;padding:10px 20px;"><a href="${b.href ?? "#"}" style="color:#ffffff;font-weight:600;text-decoration:none;font-size:14px;">${b.label ?? "Click here"}</a></td></tr></table>`);
        break;
      }
    }
  }
  rows.push("</div>");
  rows.push("</body></html>");
  return rows.join("\n");
}

export function blocksToText(blocks: Block[]): string {
  return blocks.flatMap((b) => {
    switch (b.type) {
      case "h1": case "h2": case "h3": case "h4": case "h5": case "h6":
      case "paragraph": case "blockquote":
        return [stripTags(b.html ?? ""), ""];
      case "bullet_list":  return [...(b.items ?? []).map((s) => `• ${stripTags(s)}`), ""];
      case "ordered_list": return [...(b.items ?? []).map((s, i) => `${i + 1}. ${stripTags(s)}`), ""];
      case "divider":  return ["---", ""];
      case "button":   return [`${b.label ?? ""}: ${b.href ?? ""}`, ""];
      case "image":    return [b.src ?? "", ""];
      default: return [];
    }
  }).join("\n").trim();
}

// ---------------------------------------------------------------------------
// Inline mark helpers (module-level — no component deps)
// ---------------------------------------------------------------------------


// ---------------------------------------------------------------------------
// Focus intent & cursor helpers
// ---------------------------------------------------------------------------

type FocusAt = "start" | "end" | { offset: number };

interface FocusIntent {
  id: string;
  at: FocusAt;
}

// Count text characters in an HTML string (for locating merge join point).
function htmlTextLength(html: string): number {
  const tmp = document.createElement("div");
  tmp.innerHTML = html;
  return tmp.textContent?.length ?? 0;
}

// Place cursor at start, end, or a text-character offset inside el.
function placeCursor(el: HTMLElement, at: FocusAt) {
  el.focus();
  const sel = window.getSelection();
  if (!sel) return;
  const range = document.createRange();

  if (at === "start") {
    range.setStart(el, 0);
    range.collapse(true);
  } else if (at === "end") {
    range.selectNodeContents(el);
    range.collapse(false);
  } else {
    // Walk text nodes accumulating length until we reach the offset.
    let remaining = at.offset;
    const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT);
    let placed = false;
    while (walker.nextNode()) {
      const node = walker.currentNode as Text;
      if (remaining <= node.length) {
        range.setStart(node, remaining);
        range.collapse(true);
        placed = true;
        break;
      }
      remaining -= node.length;
    }
    if (!placed) {
      range.selectNodeContents(el);
      range.collapse(false);
    }
  }
  sel.removeAllRanges();
  sel.addRange(range);
}

// True if the cursor is visually on the first line of el.
function isOnFirstLine(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0) return true;
  const range = sel.getRangeAt(0).cloneRange();
  range.collapse(true);
  const cursorRect = range.getBoundingClientRect();
  const elRect = el.getBoundingClientRect();
  // Zero rect = empty element; treat as first line.
  if (cursorRect.top === 0 && cursorRect.left === 0) return true;
  return cursorRect.top < elRect.top + 8;
}

// True if the cursor is visually on the last line of el.
function isOnLastLine(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0) return true;
  const range = sel.getRangeAt(0).cloneRange();
  range.collapse(true);
  const cursorRect = range.getBoundingClientRect();
  const elRect = el.getBoundingClientRect();
  if (cursorRect.top === 0 && cursorRect.left === 0) return true;
  return cursorRect.bottom > elRect.bottom - 8;
}

// True if the cursor is at the very start of el's text content.
function isCursorAtStart(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0 || !sel.isCollapsed) return false;
  const range = sel.getRangeAt(0);
  // Clone and try to move backwards; if we can't, we're at the start.
  const testRange = document.createRange();
  testRange.setStart(el, 0);
  testRange.setEnd(range.startContainer, range.startOffset);
  return testRange.toString().length === 0;
}

// True if the cursor is at the very end of el's text content.
function isCursorAtEnd(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0 || !sel.isCollapsed) return false;
  const range = sel.getRangeAt(0);
  const testRange = document.createRange();
  testRange.setStart(range.endContainer, range.endOffset);
  testRange.selectNodeContents(el);
  testRange.setStart(range.endContainer, range.endOffset);
  return testRange.toString().length === 0;
}

// ---------------------------------------------------------------------------
// Block type definitions (for menus)
// ---------------------------------------------------------------------------

interface BlockTypeDef { type: BlockType; label: string; icon: React.ReactNode }

const BLOCK_TYPES: BlockTypeDef[] = [
  { type: "paragraph",     label: "Text",          icon: <Type size={13} /> },
  { type: "h1",            label: "Heading 1",     icon: <span className="text-[11px] font-black">H1</span> },
  { type: "h2",            label: "Heading 2",     icon: <span className="text-[11px] font-black">H2</span> },
  { type: "h3",            label: "Heading 3",     icon: <span className="text-[11px] font-black">H3</span> },
  { type: "h4",            label: "Heading 4",     icon: <span className="text-[11px] font-black">H4</span> },
  { type: "h5",            label: "Heading 5",     icon: <span className="text-[11px] font-black">H5</span> },
  { type: "h6",            label: "Heading 6",     icon: <span className="text-[11px] font-black">H6</span> },
  { type: "bullet_list",   label: "Bullet list",   icon: <List size={13} /> },
  { type: "ordered_list",  label: "Numbered list", icon: <ListOrdered size={13} /> },
  { type: "blockquote",    label: "Quote",         icon: <Quote size={13} /> },
  { type: "divider",       label: "Divider",       icon: <Minus size={13} /> },
  { type: "image",         label: "Image",         icon: <ImageIcon size={13} /> },
  { type: "button",        label: "Button",        icon: <Square size={13} /> },
];

function makeEmpty(type: BlockType = "paragraph"): Block {
  const id = uid();
  if (type === "divider") return { id, type };
  if (type === "image")   return { id, type, src: "", alt: "" };
  if (type === "button")  return { id, type, label: "Click here", href: "" };
  if (type === "bullet_list" || type === "ordered_list") return { id, type, items: [""] };
  return { id, type, html: "" };
}

// ---------------------------------------------------------------------------
// Small reusable dropdown menu
// ---------------------------------------------------------------------------

function DropdownMenu({ trigger, children }: { trigger: React.ReactNode; children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: globalThis.MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div ref={ref} className="relative">
      <div onMouseDown={(e) => { e.preventDefault(); setOpen((o) => !o); }}>{trigger}</div>
      {open && (
        <div
          className="absolute left-0 top-full mt-1 w-44 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1 overflow-hidden"
          onMouseDown={(e) => e.preventDefault()}
        >
          {children}
        </div>
      )}
    </div>
  );
}


// ---------------------------------------------------------------------------
// ContentEditable block — isolates DOM ownership
// ---------------------------------------------------------------------------

interface EditableProps {
  initialHtml: string;
  className?: string;
  placeholder?: string;
  onHtmlChange: (html: string) => void;
  onKeyDown: (e: KeyboardEvent<HTMLDivElement>) => void;
  onFocus?: () => void;
  autoFocus?: boolean;
  shouldFocus?: boolean;
  divRef?: React.RefObject<HTMLDivElement | null>;
}

function Editable({ initialHtml, className, placeholder, onHtmlChange, onKeyDown, onFocus, autoFocus, shouldFocus, divRef }: EditableProps) {
  const innerRef = useRef<HTMLDivElement>(null);
  const ref = divRef ?? innerRef;

  // Set content once on mount only — never touch innerHTML again from React
  useEffect(() => {
    if (ref.current) ref.current.innerHTML = initialHtml;
    if (autoFocus || shouldFocus) {
      ref.current?.focus();
      // Place cursor at start
      const el = ref.current;
      if (el) {
        const range = document.createRange();
        range.setStart(el, 0);
        range.collapse(true);
        const sel = window.getSelection();
        sel?.removeAllRanges();
        sel?.addRange(range);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleInput = () => {
    onHtmlChange(ref.current?.innerHTML ?? "");
  };

  return (
    <div
      ref={ref}
      contentEditable
      suppressContentEditableWarning
      data-placeholder={placeholder}
      className={`outline-none min-h-[1.4em] ${className ?? ""} empty:before:content-[attr(data-placeholder)] empty:before:text-gray-300`}
      onInput={handleInput}
      onKeyDown={onKeyDown}
      onFocus={onFocus}
    />
  );
}

// ---------------------------------------------------------------------------
// Image block — upload + URL tabs
// ---------------------------------------------------------------------------

const ALLOWED_TYPES = ["image/jpeg", "image/png", "image/gif", "image/webp"];
const MAX_SIZE = 1 * 1024 * 1024; // 1 MB
const EMAIL_MAX_WIDTH = 600;

interface ImageBlockProps {
  block: Block;
  onChange: (b: Block) => void;
  onDelete: () => void;
  toolbar: React.ReactNode;
}

function ImageBlock({ block, onChange, onDelete, toolbar }: ImageBlockProps) {
  void onDelete;
  const [tab, setTab] = useState<"upload" | "url">(block.src ? "url" : "upload");
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [widthWarning, setWidthWarning] = useState("");
  const dropRef = useRef<HTMLLabelElement>(null);
  const [dragging, setDragging] = useState(false);

  // Suppress unused warning — onDelete is used via deleteBtn but TS doesn't see it
  void onDelete;

  const checkWidth = (url: string) => {
    const img = new window.Image();
    img.onload = () => {
      if (img.naturalWidth > EMAIL_MAX_WIDTH) {
        setWidthWarning(`Image is ${img.naturalWidth}px wide — email clients cap content at ${EMAIL_MAX_WIDTH}px. It will be scaled down.`);
      } else {
        setWidthWarning("");
      }
    };
    img.src = url;
  };

  const handleFile = async (file: File) => {
    setUploadError("");
    setWidthWarning("");

    if (!ALLOWED_TYPES.includes(file.type)) {
      setUploadError("Unsupported type — use jpg, png, gif or webp.");
      return;
    }
    if (file.size > MAX_SIZE) {
      setUploadError(`File too large (${(file.size / 1024).toFixed(0)} KB) — maximum is 1 MB.`);
      return;
    }

    setUploading(true);
    try {
      const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`${API_BASE}/v1/uploads/image`, { method: "POST", body: fd, credentials: "include" });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error ?? `Upload failed (${res.status})`);
      }
      const { url } = await res.json();
      onChange({ ...block, src: url });
      checkWidth(url);
      setTab("url"); // switch to url tab so preview is visible
    } catch (e) {
      setUploadError(String(e));
    } finally {
      setUploading(false);
    }
  };

  const handleUrlChange = (url: string) => {
    const trimmed = url.trim();
    onChange({ ...block, src: trimmed });
    if (trimmed) checkWidth(trimmed);
    else setWidthWarning("");
  };

  const isLocalhost = block.src?.includes("localhost") || block.src?.includes("127.0.0.1");

  return (
    <div className="group relative py-1 pt-8">
      {toolbar}
      <div className="space-y-2">
        {/* Tab bar */}
        <div className="flex gap-1 border-b border-gray-100 pb-1">
          {(["upload", "url"] as const).map((t) => (
            <button
              key={t}
              onMouseDown={(e) => { e.preventDefault(); setTab(t); }}
              className={`text-xs px-2.5 py-1 rounded-t font-medium transition-colors ${tab === t ? "text-gray-900 border-b-2 border-gray-900 -mb-px" : "text-gray-400 hover:text-gray-600"}`}
            >
              {t === "upload" ? "Upload" : "URL"}
            </button>
          ))}
        </div>

        {tab === "upload" && (
          <label
            ref={dropRef}
            onDragEnter={(e) => { e.preventDefault(); setDragging(true); }}
            onDragLeave={(e) => { e.preventDefault(); setDragging(false); }}
            onDragOver={(e) => e.preventDefault()}
            onDrop={(e) => {
              e.preventDefault();
              setDragging(false);
              const file = e.dataTransfer.files[0];
              if (file) handleFile(file);
            }}
            className={`flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-lg px-4 py-6 cursor-pointer transition-colors ${dragging ? "border-gray-400 bg-gray-50" : "border-gray-200 hover:border-gray-300"}`}
          >
            <input
              type="file"
              accept=".jpg,.jpeg,.png,.gif,.webp"
              className="sr-only"
              onChange={(e) => { const f = e.target.files?.[0]; if (f) handleFile(f); }}
            />
            {uploading ? (
              <p className="text-xs text-gray-400">Uploading…</p>
            ) : (
              <>
                <ImageIcon size={20} className="text-gray-300" />
                <p className="text-xs text-gray-500 text-center">
                  <span className="font-medium text-gray-700">Click to upload</span> or drag and drop
                </p>
                <p className="text-xs text-gray-400">JPG, PNG, GIF, WEBP · max 1 MB</p>
              </>
            )}
          </label>
        )}

        {tab === "url" && (
          <input
            value={block.src ?? ""}
            onChange={(e) => handleUrlChange(e.target.value)}
            placeholder="https://example.com/image.png"
            className="w-full border border-gray-200 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
        )}

        {/* Alt text — always visible */}
        <input
          value={block.alt ?? ""}
          onChange={(e) => onChange({ ...block, alt: e.target.value })}
          placeholder="Alt text (required for accessibility)"
          className={`w-full border rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 ${!block.alt ? "border-amber-200" : "border-gray-200"}`}
        />
        {!block.alt && (
          <p className="text-xs text-amber-600">Alt text is required — email clients show it when images are blocked.</p>
        )}

        {/* Errors and warnings */}
        {uploadError && <p className="text-xs text-red-600">{uploadError}</p>}
        {widthWarning && <p className="text-xs text-amber-600">{widthWarning}</p>}
        {isLocalhost && (
          <div className="rounded-md bg-red-50 border border-red-200 px-3 py-2">
            <p className="text-xs text-red-700 font-medium">Image will not appear in sent emails</p>
            <p className="text-xs text-red-600 mt-0.5">
              This image is stored on localhost and is only reachable on your machine.
              Email clients cannot fetch it. In production, set <code className="font-mono bg-red-100 px-1 rounded">API_BASE_URL</code> to your public domain.
              For now, use the URL tab and paste a publicly hosted image instead.
            </p>
          </div>
        )}

        {/* Preview */}
        {block.src && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={block.src}
            alt={block.alt ?? ""}
            className="max-w-full rounded border border-gray-200"
            style={{ maxWidth: EMAIL_MAX_WIDTH }}
          />
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Block row
// ---------------------------------------------------------------------------

// Handle exposed by BlockRow for imperative focus/content updates.
interface BlockRowHandle {
  focus(at: FocusAt): void;
  setContent(html: string, thenFocusAt?: FocusAt): void;
}

interface RowProps {
  block: Block;
  isFirst: boolean;
  focused: boolean;
  onChange: (b: Block) => void;
  onDelete: () => void;
  onAddAfter: (type?: BlockType) => void;
  onSplit: (before: string, after: string) => void;
  onFocus: () => void;
  applyMark: (mark: InlineMark) => void;
  focusIntent: FocusIntent | null;
  onFocusHandled: () => void;
  onNavigatePrev: () => void;
  onNavigateNext: () => void;
  onMergePrev: (myHtml: string) => void;
  onMergeNext: (myHtml: string) => void;
}

// Single toolbar that appears above a block on hover.
// When text is selected inside the block, mark/link controls become visible.
function BlockToolbar({
  block,
  onChangeType,
  onChangeAlign,
  onDelete,
  onAddAfter,
  showAlign,
  hasSelection,
  onMark,
  savedRange,
  onUnlink,
}: {
  block: Block;
  onChangeType: (t: BlockType) => void;
  onChangeAlign: (a: BlockAlign) => void;
  onDelete: () => void;
  onAddAfter: () => void;
  showAlign: boolean;
  hasSelection: boolean;
  onMark: (m: InlineMark) => void;
  savedRange: React.RefObject<Range | null>;
  onUnlink: (a: HTMLAnchorElement) => void;
}) {
  const currentAlign = block.align ?? "left";
  const [linkMode, setLinkMode] = useState(false);
  const [linkUrl, setLinkUrl] = useState("");
  const linkRange = useRef<Range | null>(null);
  const toolbarRef = useRef<HTMLDivElement>(null);

  // True if any part of the saved selection touches an <a> element.
  const selectionTouchesLink = (): boolean => {
    const range = savedRange.current;
    if (!range) return false;
    const ancestor = range.commonAncestorContainer;
    const el = ancestor.nodeType === Node.TEXT_NODE ? ancestor.parentElement : ancestor as Element;
    return !!el?.closest("a");
  };

  const openLink = (e: React.MouseEvent) => {
    e.preventDefault();
    const sel = window.getSelection();
    const range = (sel && !sel.isCollapsed && sel.rangeCount > 0)
      ? sel.getRangeAt(0).cloneRange()
      : savedRange.current?.cloneRange() ?? null;
    linkRange.current = range;

    let existingHref = "";
    if (range) {
      const ancestor = range.commonAncestorContainer;
      const el = ancestor.nodeType === Node.TEXT_NODE ? ancestor.parentElement : ancestor as Element;
      const enclosingAnchor = el?.closest("a") as HTMLAnchorElement | null;
      if (enclosingAnchor) existingHref = enclosingAnchor.getAttribute("href") ?? "";
    }

    setLinkUrl(existingHref);
    setLinkMode(true);
  };

  const applyLink = () => {
    if (!linkUrl || !linkRange.current) { cancelLink(); return; }
    const range = linkRange.current;
    const fragment = range.extractContents();
    fragment.querySelectorAll("a").forEach((el) => {
      const parent = el.parentNode!;
      while (el.firstChild) parent.insertBefore(el.firstChild, el);
      parent.removeChild(el);
    });
    const a = document.createElement("a");
    a.href = linkUrl.startsWith("http") ? linkUrl : `https://${linkUrl}`;
    a.target = "_blank";
    a.rel = "noopener noreferrer";
    a.style.cssText = "color:#2563eb;text-decoration:underline;";
    a.appendChild(fragment);
    range.insertNode(a);
    const host = a.closest("[contenteditable]") as HTMLElement | null;
    if (host) host.dispatchEvent(new Event("input", { bubbles: true }));
    cancelLink();
  };

  const cancelLink = () => {
    setLinkMode(false);
    setLinkUrl("");
    linkRange.current = null;
  };

  // Close link mode when clicking outside the toolbar.
  useEffect(() => {
    if (!linkMode) return;
    const handler = (e: MouseEvent) => {
      if (toolbarRef.current && !toolbarRef.current.contains(e.target as Node)) {
        cancelLink();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [linkMode]);

  return (
    <div ref={toolbarRef} className={`absolute -top-7 left-0 right-0 flex items-center gap-0.5 transition-opacity z-20 pointer-events-none group-hover:pointer-events-auto ${linkMode ? "opacity-100 pointer-events-auto" : "opacity-0 group-hover:opacity-100"}`}>
      <div className="flex items-center gap-0.5 bg-white border border-gray-200 rounded-lg shadow-sm px-1 py-0.5">
        {/* Type picker */}
        <DropdownMenu
          trigger={
            <button className="flex items-center gap-0.5 px-1.5 py-1 rounded text-xs text-gray-500 hover:bg-gray-100">
              {BLOCK_TYPES.find((t) => t.type === block.type)?.icon}
              <ChevronDown size={9} className="mt-px opacity-60" />
            </button>
          }
        >
          {BLOCK_TYPES.map((bt) => (
            <button
              key={bt.type}
              onMouseDown={(e) => { e.preventDefault(); onChangeType(bt.type); }}
              className={`w-full flex items-center gap-2 px-3 py-1.5 text-xs text-left hover:bg-gray-50 ${block.type === bt.type ? "font-semibold text-gray-900" : "text-gray-600"}`}
            >
              <span className="w-4 flex-none flex justify-center text-gray-400">{bt.icon}</span>
              {bt.label}
            </button>
          ))}
        </DropdownMenu>

        {/* Align buttons — only for blocks that support alignment */}
        {showAlign && (
          <>
            <div className="w-px h-3.5 bg-gray-200 mx-0.5" />
            {(["left", "center", "right"] as BlockAlign[]).map((a) => (
              <button
                key={a}
                onMouseDown={(e) => { e.preventDefault(); onChangeAlign(a); }}
                className={`p-1 rounded transition-colors ${currentAlign === a ? "text-gray-900 bg-gray-100" : "text-gray-400 hover:text-gray-700 hover:bg-gray-50"}`}
                title={`Align ${a}`}
              >
                {a === "left"   && <AlignLeft size={12} />}
                {a === "center" && <AlignCenter size={12} />}
                {a === "right"  && <AlignRight size={12} />}
              </button>
            ))}
          </>
        )}

        {/* Inline mark + link controls — visible when text is selected, or while link input is open */}
        {(hasSelection || linkMode) && (
          <>
            <div className="w-px h-3.5 bg-gray-200 mx-0.5" />
            {linkMode ? (
              <>
                <input
                  autoFocus
                  value={linkUrl}
                  onChange={(e) => setLinkUrl(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") { e.preventDefault(); applyLink(); }
                    if (e.key === "Escape") cancelLink();
                  }}
                  placeholder="https://example.com"
                  className="text-xs border border-gray-200 rounded px-2 py-0.5 w-40 outline-none focus:ring-1 focus:ring-gray-400"
                />
                <button onMouseDown={(e) => { e.preventDefault(); applyLink(); }} className="text-xs text-white bg-gray-900 hover:bg-gray-700 px-2 py-0.5 rounded ml-0.5">Apply</button>
              </>
            ) : (
              <>
                {(["bold", "italic", "underline", "strikethrough"] as InlineMark[]).map((m) => (
                  <button
                    key={m}
                    onMouseDown={(e) => { e.preventDefault(); onMark(m); }}
                    className="p-1 rounded text-gray-500 hover:text-gray-900 hover:bg-gray-100 transition-colors"
                    title={m.charAt(0).toUpperCase() + m.slice(1)}
                  >
                    {m === "bold"          && <Bold size={12} />}
                    {m === "italic"        && <Italic size={12} />}
                    {m === "underline"     && <UnderlineIcon size={12} />}
                    {m === "strikethrough" && <Strikethrough size={12} />}
                  </button>
                ))}
                <button
                  onMouseDown={openLink}
                  className="p-1 rounded text-gray-500 hover:text-gray-900 hover:bg-gray-100 transition-colors"
                  title={selectionTouchesLink() ? "Edit link" : "Add link"}
                >
                  {selectionTouchesLink() ? <Pencil size={12} /> : <Link size={12} />}
                </button>
                {selectionTouchesLink() && (
                  <button
                    onMouseDown={(e) => {
                      e.preventDefault();
                      const node = savedRange.current?.commonAncestorContainer;
                      if (!node) return;
                      const el = node.nodeType === Node.TEXT_NODE ? (node as Text).parentElement : node as Element;
                      const anchor = el?.closest("a") as HTMLAnchorElement | null;
                      if (anchor) onUnlink(anchor);
                    }}
                    className="p-1 rounded text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors"
                    title="Remove link"
                  >
                    <Link2Off size={12} />
                  </button>
                )}
              </>
            )}
          </>
        )}


        <div className="w-px h-3.5 bg-gray-200 mx-0.5" />

        {/* Delete */}
        <button
          onMouseDown={(e) => { e.preventDefault(); onDelete(); }}
          className="p-1 rounded text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors"
          title="Delete block"
        >
          <X size={12} />
        </button>

        {/* Add below */}
        <button
          onMouseDown={(e) => { e.preventDefault(); onAddAfter(); }}
          className="p-1 rounded text-gray-400 hover:text-gray-700 hover:bg-gray-50 transition-colors"
          title="Add block below"
        >
          <Plus size={12} />
        </button>
      </div>
    </div>
  );
}

const BlockRow = forwardRef<BlockRowHandle, RowProps>(function BlockRow(
  { block, onChange, onDelete, onAddAfter, onSplit, onFocus, isFirst,
    applyMark, focusIntent, onFocusHandled,
    onNavigatePrev, onNavigateNext, onMergePrev, onMergeNext },
  ref
) {
  const editableRef = useRef<HTMLDivElement>(null);
  const blockRef = useRef<HTMLDivElement>(null);
  const [hasSelection, setHasSelection] = useState(false);
  const savedRange = useRef<Range | null>(null);
  const [hoveredAnchor, setHoveredAnchor] = useState<HTMLAnchorElement | null>(null);

  // Expose imperative handle for focus and content updates from parent.
  useImperativeHandle(ref, () => ({
    focus(at: FocusAt) {
      const el = editableRef.current;
      if (el) placeCursor(el, at);
    },
    setContent(html: string, thenFocusAt?: FocusAt) {
      const el = editableRef.current;
      if (!el) return;
      el.innerHTML = html;
      el.dispatchEvent(new Event("input", { bubbles: true }));
      if (thenFocusAt !== undefined) placeCursor(el, thenFocusAt);
    },
  }));

  // Apply focus intent when this block is targeted.
  useEffect(() => {
    if (!focusIntent || focusIntent.id !== block.id) return;
    const el = editableRef.current;
    if (el) placeCursor(el, focusIntent.at);
    onFocusHandled();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [focusIntent]);

  // Track which <a> the pointer is hovering inside this block.
  useEffect(() => {
    const el = blockRef.current;
    if (!el) return;
    const onOver = (e: MouseEvent) => {
      const a = (e.target as Element).closest("a") as HTMLAnchorElement | null;
      setHoveredAnchor(a && el.contains(a) ? a : null);
    };
    const onOut = (e: MouseEvent) => {
      const to = e.relatedTarget as Node | null;
      if (!to || !el.contains(to)) setHoveredAnchor(null);
    };
    el.addEventListener("mouseover", onOver);
    el.addEventListener("mouseout", onOut);
    return () => { el.removeEventListener("mouseover", onOver); el.removeEventListener("mouseout", onOut); };
  }, []);

  useEffect(() => {
    const onSelectionChange = () => {
      const sel = window.getSelection();
      if (!sel || sel.isCollapsed || sel.rangeCount === 0) {
        setHasSelection(false);
        savedRange.current = null;
        return;
      }
      const range = sel.getRangeAt(0);
      const container = blockRef.current;
      if (container && container.contains(range.commonAncestorContainer)) {
        savedRange.current = range.cloneRange();
        setHasSelection(true);
      } else {
        setHasSelection(false);
        savedRange.current = null;
      }
    };
    document.addEventListener("selectionchange", onSelectionChange);
    return () => document.removeEventListener("selectionchange", onSelectionChange);
  }, []);

  const changeType = (type: BlockType) => {
    const text = stripTags(block.html ?? "");
    onChange(Object.assign(makeEmpty(type), type === "bullet_list" || type === "ordered_list"
      ? { items: [text] }
      : type === "divider" || type === "image" || type === "button"
        ? {}
        : { html: text }
    ));
  };

  const handleHtmlChange = (html: string) => {
    onChange({ ...block, html });
  };

  const baseKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    const el = editableRef.current;
    const isMod = e.metaKey || e.ctrlKey;

    // ── Cmd/Ctrl shortcuts ──────────────────────────────────────────────────
    if (isMod && e.key === "b") { e.preventDefault(); applyMark("bold"); return; }
    if (isMod && e.key === "i") { e.preventDefault(); applyMark("italic"); return; }
    if (isMod && e.key === "u") { e.preventDefault(); applyMark("underline"); return; }

    // ── Enter: split block ──────────────────────────────────────────────────
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!el) { onSplit("", ""); return; }
      const sel = window.getSelection();
      let afterHtml = "";
      if (sel && sel.rangeCount > 0) {
        const range = sel.getRangeAt(0);
        const postRange = document.createRange();
        postRange.selectNodeContents(el);
        postRange.setStart(range.startContainer, range.startOffset);
        const fragment = postRange.extractContents();
        const tmp = document.createElement("div");
        tmp.appendChild(fragment);
        afterHtml = tmp.innerHTML;
      }
      onChange({ ...block, html: el.innerHTML });
      onSplit("", afterHtml);
      return;
    }

    // ── Backspace: delete empty block OR merge with previous ────────────────
    if (e.key === "Backspace") {
      if (!el) return;
      if (el.innerText === "") {
        e.preventDefault();
        onDelete();
        return;
      }
      if (isCursorAtStart(el)) {
        e.preventDefault();
        onMergePrev(el.innerHTML);
        return;
      }
    }

    // ── Delete: merge next block into current ───────────────────────────────
    if (e.key === "Delete") {
      if (!el) return;
      if (isCursorAtEnd(el)) {
        e.preventDefault();
        onMergeNext(el.innerHTML);
        return;
      }
    }

    // ── Arrow up: navigate to previous block ────────────────────────────────
    if (e.key === "ArrowUp") {
      if (!el) return;
      if (isOnFirstLine(el)) {
        e.preventDefault();
        onNavigatePrev();
        return;
      }
    }

    // ── Arrow down: navigate to next block ──────────────────────────────────
    if (e.key === "ArrowDown") {
      if (!el) return;
      if (isOnLastLine(el)) {
        e.preventDefault();
        onNavigateNext();
        return;
      }
    }

    // ── Tab: insert two spaces (prevent focus jump) ─────────────────────────
    if (e.key === "Tab") {
      e.preventDefault();
      const sel = window.getSelection();
      if (sel && sel.rangeCount > 0) {
        const range = sel.getRangeAt(0);
        range.deleteContents();
        const spaces = document.createTextNode("    ");
        range.insertNode(spaces);
        range.setStartAfter(spaces);
        range.collapse(true);
        sel.removeAllRanges();
        sel.addRange(range);
        el?.dispatchEvent(new Event("input", { bubbles: true }));
      }
    }
  };

  const unlinkAnchor = (anchor: HTMLAnchorElement) => {
    const parent = anchor.parentNode!;
    while (anchor.firstChild) parent.insertBefore(anchor.firstChild, anchor);
    parent.removeChild(anchor);
    setHoveredAnchor(null);
    const host = editableRef.current;
    if (host) host.dispatchEvent(new Event("input", { bubbles: true }));
  };

  const toolbar = (
    <BlockToolbar
      block={block}
      onChangeType={changeType}
      onChangeAlign={(a) => onChange({ ...block, align: a })}
      onDelete={onDelete}
      onAddAfter={() => onAddAfter("paragraph")}
      showAlign={block.type !== "divider"}
      hasSelection={hasSelection}
      onMark={applyMark}
      savedRange={savedRange}
      onUnlink={unlinkAnchor}
    />
  );

  // ---- Divider ----
  if (block.type === "divider") {
    return (
      <div ref={blockRef} className="group relative py-3 pt-8">
        {toolbar}
        <hr className="border-gray-200" />
      </div>
    );
  }

  // ---- Image ----
  if (block.type === "image") {
    return (
      <div ref={blockRef}>
        <ImageBlock
          block={block}
          onChange={onChange}
          onDelete={onDelete}
          toolbar={toolbar}
        />
      </div>
    );
  }

  // ---- Button ----
  if (block.type === "button") {
    const btnJustify = block.align === "center" ? "justify-center" : block.align === "right" ? "justify-end" : "justify-start";
    return (
      <div ref={blockRef} className="group relative py-1 pt-8">
        {toolbar}
        <div className="space-y-1.5">
          <input
            value={block.label ?? ""}
            onChange={(e) => onChange({ ...block, label: e.target.value })}
            placeholder="Button label"
            className="w-full border border-gray-200 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          <input
            value={block.href ?? ""}
            onChange={(e) => onChange({ ...block, href: e.target.value })}
            placeholder="URL (https://…)"
            className="w-full border border-gray-200 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          <div className={`flex ${btnJustify}`}>
            <span className="inline-block bg-gray-900 text-white text-sm font-semibold px-5 py-2 rounded-md select-none">
              {block.label || "Click here"}
            </span>
          </div>
        </div>
      </div>
    );
  }

  // ---- Lists ----
  if (block.type === "bullet_list" || block.type === "ordered_list") {
    const items: string[] = block.items ?? [""];
    const Tag = block.type === "bullet_list" ? "ul" : "ol";
    const listClass = block.type === "bullet_list" ? "list-disc pl-5" : "list-decimal pl-5";
    const alignStyle = block.align && block.align !== "left" ? { textAlign: block.align as "center" | "right" } : undefined;
    return (
      <div ref={blockRef} className="group relative py-0.5 pt-8">
        {toolbar}
        <Tag className={`${listClass} space-y-0.5`} style={alignStyle}>
          {items.map((item, i) => (
            <li key={i}>
              <Editable
                initialHtml={item}
                className="text-[15px] leading-relaxed"
                onHtmlChange={(html) => {
                  const next = items.map((s, j) => j === i ? html : s);
                  onChange({ ...block, items: next });
                }}
                onKeyDown={(e) => {
                  const isMod = e.metaKey || e.ctrlKey;
                  if (isMod && e.key === "b") { e.preventDefault(); applyMark("bold"); return; }
                  if (isMod && e.key === "i") { e.preventDefault(); applyMark("italic"); return; }
                  if (isMod && e.key === "u") { e.preventDefault(); applyMark("underline"); return; }
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    const next = [...items.slice(0, i + 1), "", ...items.slice(i + 1)];
                    onChange({ ...block, items: next });
                  }
                  if (e.key === "Backspace" && (e.currentTarget as HTMLDivElement).innerText === "") {
                    e.preventDefault();
                    if (items.length === 1) onDelete();
                    else onChange({ ...block, items: items.filter((_, j) => j !== i) });
                  }
                  if (e.key === "ArrowUp" && i === 0) {
                    const el = e.currentTarget as HTMLDivElement;
                    if (isOnFirstLine(el)) { e.preventDefault(); onNavigatePrev(); }
                  }
                  if (e.key === "ArrowDown" && i === items.length - 1) {
                    const el = e.currentTarget as HTMLDivElement;
                    if (isOnLastLine(el)) { e.preventDefault(); onNavigateNext(); }
                  }
                }}
                onFocus={onFocus}
              />
            </li>
          ))}
        </Tag>
      </div>
    );
  }

  // ---- Text-based (paragraph, headings, blockquote) ----
  const editorClass = (() => {
    switch (block.type) {
      case "h1": return "text-3xl font-bold leading-tight";
      case "h2": return "text-2xl font-bold leading-tight";
      case "h3": return "text-xl font-semibold leading-snug";
      case "h4": return "text-lg font-semibold leading-snug";
      case "h5": return "text-base font-semibold";
      case "h6": return "text-sm font-semibold uppercase tracking-wide text-gray-500";
      case "blockquote": return "border-l-4 border-gray-200 pl-4 text-gray-500 italic";
      default:  return "text-[15px] leading-relaxed";
    }
  })();

  const alignStyle = block.align && block.align !== "left" ? { textAlign: block.align as "center" | "right" } : undefined;

  return (
    <div ref={blockRef} className="group relative py-0.5 pt-8">
      {toolbar}
      <div className="relative" style={alignStyle}>
        <Editable
          divRef={editableRef}
          initialHtml={block.html ?? ""}
          className={editorClass}
          placeholder={!isFirst ? undefined : "Start writing…"}
          onHtmlChange={handleHtmlChange}
          onKeyDown={baseKeyDown}
          onFocus={onFocus}
          autoFocus={isFirst && !block.html}
        />
        {/* Unlink popover — appears when hovering an <a> inside this block */}
        {hoveredAnchor && (() => {
          const rect = hoveredAnchor.getBoundingClientRect();
          const containerRect = blockRef.current?.getBoundingClientRect();
          if (!containerRect) return null;
          const top = rect.bottom - containerRect.top + 4;
          const left = rect.left - containerRect.left;
          return (
            <div
              style={{ top, left }}
              className="absolute z-30 flex items-center gap-1.5 bg-white border border-gray-200 rounded-lg shadow-md px-2 py-1 text-xs text-gray-600 pointer-events-auto"
              onMouseEnter={() => setHoveredAnchor(hoveredAnchor)}
              onMouseLeave={() => setHoveredAnchor(null)}
            >
              <span className="max-w-45 truncate text-blue-600">{hoveredAnchor.getAttribute("href")}</span>
              <div className="w-px h-3 bg-gray-200" />
              <button
                onMouseDown={(e) => { e.preventDefault(); unlinkAnchor(hoveredAnchor); }}
                className="text-gray-400 hover:text-red-500 transition-colors"
                title="Remove link"
              >
                <Link2Off size={12} />
              </button>
            </div>
          );
        })()}
      </div>
    </div>
  );
});

// ---------------------------------------------------------------------------
// Main editor
// ---------------------------------------------------------------------------

export interface TemplateEditorProps {
  initialBlocks?: Block[];
  onChange: (blocks: Block[], html: string, text: string) => void;
}

export default function TemplateEditor({ initialBlocks, onChange }: TemplateEditorProps) {
  const [blocks, setBlocks] = useState<Block[]>(
    initialBlocks && initialBlocks.length > 0 ? initialBlocks : [makeEmpty("paragraph")]
  );
  const [focusedId, setFocusedId] = useState<string | null>(null);

  const update = useCallback((next: Block[]) => {
    setBlocks(next);
    onChange(next, blocksToHtml(next), blocksToText(next));
  }, [onChange]);


  // Toggle an inline mark on the current selection, preserving the selection
  // afterward so the toolbar stays visible.
  //
  // Strategy:
  //  - fullyMarked → remove: split wrapper elements at selection boundaries so
  //    only the selected portion is unwrapped, leaving text outside untouched.
  //  - not fullyMarked → add via execCommand (handles partial selections), then
  //    normalise <b>→<strong> and <i>→<em>.
  //  Selection is restored by character offset after any DOM mutation.
  const applyMark = useCallback((mark: InlineMark) => {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || sel.rangeCount === 0) return;

    const range = sel.getRangeAt(0);
    const anchor = range.commonAncestorContainer;
    const host = (anchor.nodeType === Node.ELEMENT_NODE ? anchor as Element : anchor.parentElement)
      ?.closest("[contenteditable]") as HTMLElement | null;
    if (!host) return;

    const tag = mark === "bold" ? "strong" : mark === "italic" ? "em" : mark === "strikethrough" ? "del" : "u";

    // Compute character offset of a (node, offset) position within host's text.
    const charOffset = (node: Node, off: number): number => {
      let count = 0;
      const w = document.createTreeWalker(host, NodeFilter.SHOW_TEXT);
      while (w.nextNode()) {
        const t = w.currentNode as Text;
        if (t === node) return count + off;
        count += t.length;
      }
      return count;
    };

    // Restore selection from character offsets into host's text nodes.
    const restoreSelection = (startChar: number, endChar: number) => {
      let s: [Node, number] | null = null;
      let e: [Node, number] | null = null;
      let count = 0;
      const w = document.createTreeWalker(host, NodeFilter.SHOW_TEXT);
      while (w.nextNode()) {
        const t = w.currentNode as Text;
        if (!s && count + t.length >= startChar) s = [t, startChar - count];
        if (!e && count + t.length >= endChar) { e = [t, endChar - count]; break; }
        count += t.length;
      }
      if (!s || !e) return;
      try {
        const r = document.createRange();
        r.setStart(s[0], s[1]);
        r.setEnd(e[0], e[1]);
        sel.removeAllRanges();
        sel.addRange(r);
      } catch { /* ignore if nodes shifted */ }
    };

    // Collect all text nodes inside the host that overlap the live selection.
    const texts: Text[] = [];
    const walker = document.createTreeWalker(host, NodeFilter.SHOW_TEXT);
    while (walker.nextNode()) {
      const node = walker.currentNode as Text;
      if (range.intersectsNode(node)) texts.push(node);
    }
    const fullyMarked = texts.length > 0 && texts.every(n => !!n.parentElement?.closest(tag));

    // Save character offsets before any mutation.
    const startChar = charOffset(range.startContainer, range.startOffset);
    const endChar   = charOffset(range.endContainer,   range.endOffset);

    if (fullyMarked) {
      // ── Remove ────────────────────────────────────────────────────────────
      // Split wrappers at selection boundaries so only the selected span is
      // unwrapped; text outside stays inside its wrapper.

      // Split at start: if cursor is mid-text-node inside a wrapper, split so
      // the pre-selection portion remains wrapped.
      const liveRange = sel.getRangeAt(0);
      if (liveRange.startContainer.nodeType === Node.TEXT_NODE && liveRange.startOffset > 0) {
        const t = liveRange.startContainer as Text;
        if (t.parentElement?.closest(tag)) {
          const after = t.splitText(liveRange.startOffset);
          liveRange.setStart(after, 0);
        }
      }
      // Split at end.
      if (liveRange.endContainer.nodeType === Node.TEXT_NODE) {
        const t = liveRange.endContainer as Text;
        if (t.parentElement?.closest(tag) && liveRange.endOffset < t.length) {
          t.splitText(liveRange.endOffset);
          // range end offset is unchanged — it's within the original node
        }
      }

      // Re-collect selected text nodes after splitting.
      const selectedTexts: Text[] = [];
      const w2 = document.createTreeWalker(host, NodeFilter.SHOW_TEXT);
      while (w2.nextNode()) {
        const n = w2.currentNode as Text;
        if (liveRange.intersectsNode(n)) selectedTexts.push(n);
      }

      // Unwrap only wrappers that now sit entirely within the selection.
      const wrappers = new Set<Element>();
      selectedTexts.forEach(n => {
        const el = n.parentElement?.closest(tag);
        if (el) wrappers.add(el);
      });
      wrappers.forEach(el => {
        const parent = el.parentNode!;
        while (el.firstChild) parent.insertBefore(el.firstChild, el);
        parent.removeChild(el);
      });
      host.normalize();

      restoreSelection(startChar, endChar);
    } else {
      // ── Add ───────────────────────────────────────────────────────────────
      // execCommand handles partial selections correctly and preserves selection.
      if (mark === "bold")               document.execCommand("bold");
      else if (mark === "italic")        document.execCommand("italic");
      else if (mark === "underline")     document.execCommand("underline");
      else if (mark === "strikethrough") document.execCommand("strikeThrough");

      // Normalise browser-emitted <b>/<i> to semantic <strong>/<em>.
      if (mark === "bold") {
        host.querySelectorAll("b:not(strong)").forEach(el => {
          const strong = document.createElement("strong");
          while (el.firstChild) strong.appendChild(el.firstChild);
          el.parentNode!.replaceChild(strong, el);
        });
      } else if (mark === "italic") {
        host.querySelectorAll("i:not(em)").forEach(el => {
          const em = document.createElement("em");
          while (el.firstChild) em.appendChild(el.firstChild);
          el.parentNode!.replaceChild(em, el);
        });
      }

      // execCommand preserves selection, but tag renaming may lose it — restore.
      restoreSelection(startChar, endChar);
    }

    host.dispatchEvent(new Event("input", { bubbles: true }));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const [focusIntent, setFocusIntent] = useState<FocusIntent | null>(null);
  // Refs to each mounted BlockRow for imperative content updates (merges).
  const blockRefs = useRef<Map<string, BlockRowHandle>>(new Map());

  const addAfter = (index: number, type: BlockType = "paragraph", carryHtml = "") => {
    const newBlock = makeEmpty(type);
    if (carryHtml && (type === "paragraph" || type.startsWith("h") || type === "blockquote")) {
      newBlock.html = carryHtml;
    }
    const next = [...blocks.slice(0, index + 1), newBlock, ...blocks.slice(index + 1)];
    setFocusIntent({ id: newBlock.id, at: "start" });
    update(next);
  };

  // Use a stable ref to blocks so callbacks don't close over stale values.
  const blocksRef = useRef(blocks);
  useEffect(() => { blocksRef.current = blocks; }, [blocks]);

  return (
    <div>
      <div className="space-y-1">
        {blocks.map((block, index) => (
          <BlockRow
            key={block.id}
            ref={(handle) => {
              if (handle) blockRefs.current.set(block.id, handle);
              else blockRefs.current.delete(block.id);
            }}
            block={block}
            isFirst={index === 0}
            focused={focusedId === block.id}
            focusIntent={focusIntent?.id === block.id ? focusIntent : null}
            onFocusHandled={() => setFocusIntent(null)}
            onFocus={() => setFocusedId(block.id)}
            applyMark={applyMark}
            onChange={(b) => {
              const next = blocksRef.current.map((old, i) => i === index ? b : old);
              update(next);
            }}
            onDelete={() => {
              const cur = blocksRef.current;
              if (cur.length === 1) { update([makeEmpty()]); return; }
              // Focus the previous block (or next if first).
              const targetIndex = index > 0 ? index - 1 : 1;
              const targetId = cur.filter((_, i) => i !== index)[Math.min(targetIndex, cur.length - 2)]?.id;
              if (targetId) setFocusIntent({ id: targetId, at: index > 0 ? "end" : "start" });
              update(cur.filter((_, i) => i !== index));
            }}
            onAddAfter={(type) => addAfter(index, type)}
            onSplit={(_before, afterHtml) => addAfter(index, "paragraph", afterHtml)}
            onNavigatePrev={() => {
              const cur = blocksRef.current;
              if (index === 0) return;
              setFocusIntent({ id: cur[index - 1].id, at: "end" });
            }}
            onNavigateNext={() => {
              const cur = blocksRef.current;
              if (index === cur.length - 1) return;
              setFocusIntent({ id: cur[index + 1].id, at: "start" });
            }}
            onMergePrev={(myHtml) => {
              const cur = blocksRef.current;
              if (index === 0) return;
              const prev = cur[index - 1];
              if (prev.type === "divider" || prev.type === "image" || prev.type === "button") {
                setFocusIntent({ id: prev.id, at: "end" });
                return;
              }
              const prevHtml = prev.html ?? "";
              const joinOffset = htmlTextLength(prevHtml);
              const mergedHtml = prevHtml + myHtml;
              // Imperatively update the prev block's DOM, then remove this block.
              blockRefs.current.get(prev.id)?.setContent(mergedHtml, { offset: joinOffset });
              const next = cur
                .map((b, i) => i === index - 1 ? { ...b, html: mergedHtml } : b)
                .filter((_, i) => i !== index);
              update(next);
            }}
            onMergeNext={(myHtml) => {
              const cur = blocksRef.current;
              if (index === cur.length - 1) return;
              const next = cur[index + 1];
              if (next.type === "divider" || next.type === "image" || next.type === "button") {
                setFocusIntent({ id: next.id, at: "start" });
                return;
              }
              const joinOffset = htmlTextLength(myHtml);
              const mergedHtml = myHtml + (next.html ?? "");
              blockRefs.current.get(block.id)?.setContent(mergedHtml, { offset: joinOffset });
              const nextBlocks = cur
                .map((b, i) => i === index ? { ...b, html: mergedHtml } : b)
                .filter((_, i) => i !== index + 1);
              update(nextBlocks);
            }}
          />
        ))}
      </div>
      {/* Click below last block to append a paragraph */}
      <div
        className="h-20 cursor-text"
        onClick={() => {
          const last = blocks[blocks.length - 1];
          const isEmpty = last.type === "paragraph" && !last.html;
          if (!isEmpty) addAfter(blocks.length - 1);
        }}
      />
    </div>
  );
}
