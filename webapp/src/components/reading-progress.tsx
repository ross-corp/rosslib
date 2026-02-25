"use client";

import { useState } from "react";

type Props = {
  openLibraryId: string;
  initialPages: number | null;
  initialPercent: number | null;
  pageCount: number | null;
  initialDeviceTotalPages: number | null;
};

export default function ReadingProgress({
  openLibraryId,
  initialPages,
  initialPercent,
  pageCount,
  initialDeviceTotalPages,
}: Props) {
  const [mode, setMode] = useState<"page" | "percent">(
    initialPages != null ? "page" : "percent"
  );
  const [pages, setPages] = useState(initialPages?.toString() ?? "");
  const [percent, setPercent] = useState(initialPercent?.toString() ?? "");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const [deviceTotalPages, setDeviceTotalPages] = useState(
    initialDeviceTotalPages?.toString() ?? ""
  );
  const [editingDevice, setEditingDevice] = useState(false);
  const [savingDevice, setSavingDevice] = useState(false);

  // Effective total pages: prefer user's device page count, fall back to catalog
  const effectiveTotal = initialDeviceTotalPages ?? pageCount;

  const currentPercent =
    initialPercent ??
    (initialPages != null && effectiveTotal
      ? Math.min(100, Math.round((initialPages / effectiveTotal) * 100))
      : null);

  async function saveProgress() {
    setSaving(true);
    setMessage(null);

    const body: Record<string, unknown> = {};
    if (mode === "page") {
      const p = pages === "" ? null : parseInt(pages, 10);
      if (p !== null && (isNaN(p) || p < 0)) {
        setMessage("Invalid page number");
        setSaving(false);
        return;
      }
      body.progress_pages = p;
      // Auto-calculate percent using effective total
      if (p != null && effectiveTotal) {
        body.progress_percent = Math.min(100, Math.round((p / effectiveTotal) * 100));
      }
    } else {
      const pct = percent === "" ? null : parseInt(percent, 10);
      if (pct !== null && (isNaN(pct) || pct < 0 || pct > 100)) {
        setMessage("Must be 0-100");
        setSaving(false);
        return;
      }
      body.progress_percent = pct;
      body.progress_pages = null;
    }

    const res = await fetch(`/api/me/books/${openLibraryId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    setSaving(false);
    if (res.ok) {
      setMessage("Saved");
      setTimeout(() => setMessage(null), 2000);
    } else {
      setMessage("Failed to save");
    }
  }

  async function saveDeviceTotalPages() {
    setSavingDevice(true);
    const dtp = deviceTotalPages === "" ? null : parseInt(deviceTotalPages, 10);
    if (dtp !== null && (isNaN(dtp) || dtp < 1)) {
      setMessage("Must be at least 1");
      setSavingDevice(false);
      return;
    }

    const res = await fetch(`/api/me/books/${openLibraryId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ device_total_pages: dtp }),
    });

    setSavingDevice(false);
    if (res.ok) {
      setEditingDevice(false);
      // Reload to reflect new percentage calculations
      window.location.reload();
    } else {
      setMessage("Failed to save");
    }
  }

  return (
    <div>
      {/* Progress bar */}
      {currentPercent != null && (
        <div className="mb-3">
          <div className="flex items-center justify-between text-xs text-stone-400 mb-1">
            <span>{currentPercent}% complete</span>
            {initialPages != null && effectiveTotal && (
              <span>
                p. {initialPages} of {effectiveTotal}
              </span>
            )}
          </div>
          <div className="w-full h-1.5 bg-stone-100 rounded-full overflow-hidden">
            <div
              className="h-full bg-stone-400 rounded-full transition-all duration-300"
              style={{ width: `${currentPercent}%` }}
            />
          </div>
        </div>
      )}

      {/* Input */}
      <div className="flex items-center gap-2">
        <div className="flex items-center border border-stone-200 rounded overflow-hidden text-xs">
          <button
            type="button"
            onClick={() => setMode("page")}
            className={`px-2 py-1 transition-colors ${
              mode === "page"
                ? "bg-stone-900 text-white"
                : "text-stone-400 hover:text-stone-600"
            }`}
          >
            Page
          </button>
          <button
            type="button"
            onClick={() => setMode("percent")}
            className={`px-2 py-1 transition-colors ${
              mode === "percent"
                ? "bg-stone-900 text-white"
                : "text-stone-400 hover:text-stone-600"
            }`}
          >
            %
          </button>
        </div>

        {mode === "page" ? (
          <div className="flex items-center gap-1">
            <input
              type="number"
              min={0}
              max={effectiveTotal ?? undefined}
              value={pages}
              onChange={(e) => setPages(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && saveProgress()}
              disabled={saving}
              placeholder="Page"
              className="w-20 border border-stone-200 rounded px-2 py-1 text-xs text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            />
            {effectiveTotal && (
              <span className="text-xs text-stone-400">/ {effectiveTotal}</span>
            )}
          </div>
        ) : (
          <div className="flex items-center gap-1">
            <input
              type="number"
              min={0}
              max={100}
              value={percent}
              onChange={(e) => setPercent(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && saveProgress()}
              disabled={saving}
              placeholder="0"
              className="w-16 border border-stone-200 rounded px-2 py-1 text-xs text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            />
            <span className="text-xs text-stone-400">%</span>
          </div>
        )}

        <button
          type="button"
          onClick={saveProgress}
          disabled={saving}
          className="text-xs px-2.5 py-1 rounded bg-stone-900 text-white hover:bg-stone-700 disabled:opacity-50 transition-colors"
        >
          {saving ? "..." : "Update"}
        </button>

        {message && (
          <span className="text-xs text-stone-500">{message}</span>
        )}
      </div>

      {/* Device page count setting */}
      <div className="mt-2">
        {editingDevice ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-stone-400">My edition has</span>
            <input
              type="number"
              min={1}
              value={deviceTotalPages}
              onChange={(e) => setDeviceTotalPages(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && saveDeviceTotalPages()}
              disabled={savingDevice}
              placeholder="pages"
              className="w-20 border border-stone-200 rounded px-2 py-1 text-xs text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            />
            <span className="text-xs text-stone-400">pages</span>
            <button
              type="button"
              onClick={saveDeviceTotalPages}
              disabled={savingDevice}
              className="text-xs text-stone-600 hover:text-stone-900 transition-colors"
            >
              {savingDevice ? "..." : "Save"}
            </button>
            <button
              type="button"
              onClick={() => {
                setEditingDevice(false);
                setDeviceTotalPages(initialDeviceTotalPages?.toString() ?? "");
              }}
              className="text-xs text-stone-400 hover:text-stone-600 transition-colors"
            >
              Cancel
            </button>
          </div>
        ) : (
          <button
            type="button"
            onClick={() => setEditingDevice(true)}
            className="text-xs text-stone-400 hover:text-stone-600 transition-colors"
          >
            {initialDeviceTotalPages
              ? `My edition: ${initialDeviceTotalPages} pages (edit)`
              : "Set my edition's page count"}
          </button>
        )}
      </div>
    </div>
  );
}
