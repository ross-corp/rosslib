"use client";

import { useEffect, useRef, useState } from "react";

// ── Types ──────────────────────────────────────────────────────────────────────

type BookCandidate = {
  ol_id: string;
  title: string;
  authors: string[];
  cover_url: string | null;
  year: number | null;
};

type PreviewRow = {
  row_id: number;
  title: string;
  author: string;
  isbn13: string;
  rating: number | null;
  review_text: string | null;
  spoiler: boolean;
  date_read: string | null;
  date_added: string | null;
  exclusive_shelf_slug: string;
  custom_shelves: string[];
  status: "matched" | "ambiguous" | "unmatched";
  match?: BookCandidate;
  candidates?: BookCandidate[];
};

type PreviewResponse = {
  total: number;
  matched: number;
  ambiguous: number;
  unmatched: number;
  rows: PreviewRow[];
};

type CommitRow = {
  row_id: number;
  ol_id: string;
  title: string;
  cover_url: string | null;
  authors: string;
  publication_year: number | null;
  isbn13: string | null;
  rating: number | null;
  review_text: string | null;
  spoiler: boolean;
  date_read: string | null;
  date_added: string | null;
  exclusive_shelf_slug: string;
  custom_shelves: string[];
};

type DoneResult = { imported: number; failed: number; errors: string[] };

type SavedUnmatched = {
  title: string;
  author: string;
  isbn13: string;
  exclusive_shelf_slug: string;
};

const STORAGE_KEY = "rosslib:import:unmatched";

// ── Helpers ────────────────────────────────────────────────────────────────────

function shelfLabel(slug: string): string {
  const labels: Record<string, string> = {
    "read": "Read",
    "currently-reading": "Currently Reading",
    "want-to-read": "Want to Read",
    "owned-to-read": "Owned to Read",
    "dnf": "Did Not Finish",
  };
  return labels[slug] ?? slug;
}

function stars(rating: number | null): string {
  if (!rating) return "";
  return "★".repeat(rating) + "☆".repeat(5 - rating);
}

function mergeUnmatched(existing: SavedUnmatched[], incoming: SavedUnmatched[]): SavedUnmatched[] {
  const seen = new Set(existing.map((r) => `${r.title}|||${r.author}`));
  const added = incoming.filter((r) => !seen.has(`${r.title}|||${r.author}`));
  return [...existing, ...added];
}

// ── Component ─────────────────────────────────────────────────────────────────

export default function ImportForm() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState<string | null>(null);

  type Phase = "idle" | "previewing" | "review" | "importing" | "done";
  const [phase, setPhase] = useState<Phase>("idle");
  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [doneResult, setDoneResult] = useState<DoneResult | null>(null);

  // row_id → included in import (true) or skipped (false)
  const [selected, setSelected] = useState<Map<number, boolean>>(new Map());
  // row_id → chosen candidate ol_id (for ambiguous rows)
  const [choices, setChoices] = useState<Map<number, string>>(new Map());
  // matched section expanded
  const [matchedExpanded, setMatchedExpanded] = useState(false);

  // Unmatched books persisted from previous imports
  const [savedUnmatched, setSavedUnmatched] = useState<SavedUnmatched[]>([]);

  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) setSavedUnmatched(JSON.parse(raw));
    } catch {
      // ignore parse errors
    }
  }, []);

  function dismissUnmatched(index: number) {
    const updated = savedUnmatched.filter((_, i) => i !== index);
    setSavedUnmatched(updated);
    try {
      if (updated.length === 0) localStorage.removeItem(STORAGE_KEY);
      else localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
    } catch { /* ignore */ }
  }

  function clearAllUnmatched() {
    setSavedUnmatched([]);
    try { localStorage.removeItem(STORAGE_KEY); } catch { /* ignore */ }
  }

  // ── Upload & preview ────────────────────────────────────────────────────────

  async function handlePreview() {
    const file = fileInputRef.current?.files?.[0];
    if (!file) {
      setError("Please choose a CSV file first.");
      return;
    }
    setError(null);
    setPhase("previewing");

    const form = new FormData();
    form.append("file", file);

    try {
      const res = await fetch("/api/me/import/goodreads/preview", {
        method: "POST",
        body: form,
      });
      const data: PreviewResponse & { error?: string } = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Preview failed.");
        setPhase("idle");
        return;
      }

      // Pre-populate state: match/ambiguous rows selected, unmatched skipped.
      const sel = new Map<number, boolean>();
      const cho = new Map<number, string>();
      for (const row of data.rows) {
        sel.set(row.row_id, row.status !== "unmatched");
        if (row.status === "ambiguous" && row.candidates?.[0]) {
          cho.set(row.row_id, row.candidates[0].ol_id);
        }
      }
      setSelected(sel);
      setChoices(cho);
      setPreview(data);
      setPhase("review");
    } catch {
      setError("Network error. Please try again.");
      setPhase("idle");
    }
  }

  // ── Commit ──────────────────────────────────────────────────────────────────

  function buildCommitRows(): CommitRow[] {
    if (!preview) return [];
    const rows: CommitRow[] = [];

    for (const row of preview.rows) {
      if (!selected.get(row.row_id)) continue;

      let candidate: BookCandidate | null = null;
      if (row.status === "matched" && row.match) {
        candidate = row.match;
      } else if (row.status === "ambiguous") {
        const chosenId = choices.get(row.row_id);
        candidate =
          row.candidates?.find((c) => c.ol_id === chosenId) ??
          row.candidates?.[0] ??
          null;
      }
      if (!candidate) continue;

      rows.push({
        row_id: row.row_id,
        ol_id: candidate.ol_id,
        title: candidate.title || row.title,
        cover_url: candidate.cover_url ?? null,
        authors: (candidate.authors ?? []).join(", "),
        publication_year: candidate.year ?? null,
        isbn13: row.isbn13 || null,
        rating: row.rating,
        review_text: row.review_text,
        spoiler: row.spoiler,
        date_read: row.date_read,
        date_added: row.date_added,
        exclusive_shelf_slug: row.exclusive_shelf_slug,
        custom_shelves: row.custom_shelves,
      });
    }
    return rows;
  }

  async function handleCommit() {
    const rows = buildCommitRows();
    if (rows.length === 0) {
      setError("No books selected to import.");
      return;
    }
    setError(null);
    setPhase("importing");

    try {
      const res = await fetch("/api/me/import/goodreads/commit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ rows }),
      });
      const data: DoneResult & { error?: string } = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Import failed.");
        setPhase("review");
        return;
      }
      // Persist any unmatched rows to localStorage for later manual lookup
      if (preview) {
        const unmatched: SavedUnmatched[] = preview.rows
          .filter((r) => r.status === "unmatched")
          .map((r) => ({
            title: r.title,
            author: r.author,
            isbn13: r.isbn13,
            exclusive_shelf_slug: r.exclusive_shelf_slug,
          }));
        const merged = mergeUnmatched(savedUnmatched, unmatched);
        try {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(merged));
          setSavedUnmatched(merged);
        } catch {
          // quota exceeded or SSR — ignore
        }
      }
      setDoneResult(data);
      setPhase("done");
    } catch {
      setError("Network error. Please try again.");
      setPhase("review");
    }
  }

  // ── Counts ──────────────────────────────────────────────────────────────────

  const selectedCount = buildCommitRows().length;

  // ── Render ───────────────────────────────────────────────────────────────────

  if (phase === "done" && doneResult) {
    return (
      <div className="space-y-4">
        <p className="text-stone-900 font-medium">
          {doneResult.imported} book{doneResult.imported !== 1 ? "s" : ""} imported successfully.
        </p>
        {doneResult.failed > 0 && (
          <p className="text-sm text-red-600">
            {doneResult.failed} book{doneResult.failed !== 1 ? "s" : ""} failed to import.
          </p>
        )}
        {(doneResult.errors ?? []).length > 0 && (
          <details className="text-xs text-stone-500">
            <summary className="cursor-pointer hover:text-stone-700">Show errors</summary>
            <ul className="mt-2 space-y-1 list-disc list-inside">
              {doneResult.errors.map((e, i) => <li key={i}>{e}</li>)}
            </ul>
          </details>
        )}
        <a
          href="/me/shelves"
          className="inline-block mt-2 text-sm text-stone-900 underline hover:no-underline"
        >
          Go to your shelves
        </a>
      </div>
    );
  }

  if (phase === "previewing" || phase === "importing") {
    return (
      <div className="flex items-center gap-3 text-sm text-stone-600">
        <svg className="animate-spin h-4 w-4 text-stone-400" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        {phase === "previewing"
          ? "Matching books against Open Library — this can take up to 30 seconds…"
          : "Importing…"}
      </div>
    );
  }

  if (phase === "review" && preview) {
    const matchedRows = preview.rows.filter((r) => r.status === "matched");
    const ambiguousRows = preview.rows.filter((r) => r.status === "ambiguous");
    const unmatchedRows = preview.rows.filter((r) => r.status === "unmatched");

    return (
      <div className="space-y-8">
        {/* Summary */}
        <div className="flex flex-wrap gap-4 text-sm">
          <span className="px-2.5 py-1 rounded-full bg-green-50 border border-green-200 text-green-800">
            {preview.matched} matched
          </span>
          {preview.ambiguous > 0 && (
            <span className="px-2.5 py-1 rounded-full bg-amber-50 border border-amber-200 text-amber-800">
              {preview.ambiguous} need a choice
            </span>
          )}
          {preview.unmatched > 0 && (
            <span className="px-2.5 py-1 rounded-full bg-stone-100 border border-stone-200 text-stone-600">
              {preview.unmatched} not found
            </span>
          )}
        </div>

        {/* Ambiguous — always expanded (requires user input) */}
        {ambiguousRows.length > 0 && (
          <section>
            <h2 className="text-sm font-semibold text-stone-700 mb-3">
              Choose edition ({ambiguousRows.length})
            </h2>
            <p className="text-xs text-stone-400 mb-4">
              Multiple editions found — pick the one you read.
            </p>
            <ul className="space-y-4">
              {ambiguousRows.map((row) => (
                <li key={row.row_id} className="border border-stone-200 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <input
                      type="checkbox"
                      checked={selected.get(row.row_id) ?? true}
                      onChange={(e) =>
                        setSelected((prev) => new Map(prev).set(row.row_id, e.target.checked))
                      }
                      className="mt-0.5 rounded border-stone-300 accent-stone-900"
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-stone-900 truncate">{row.title}</p>
                      <p className="text-xs text-stone-500">{row.author}</p>
                      {row.rating && (
                        <p className="text-xs text-stone-400 mt-0.5 font-mono">{stars(row.rating)}</p>
                      )}
                      <select
                        className="mt-2 text-xs border border-stone-200 rounded px-2 py-1.5 w-full max-w-md text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400"
                        value={choices.get(row.row_id) ?? row.candidates?.[0]?.ol_id ?? ""}
                        onChange={(e) =>
                          setChoices((prev) => new Map(prev).set(row.row_id, e.target.value))
                        }
                      >
                        {(row.candidates ?? []).map((c) => (
                          <option key={c.ol_id} value={c.ol_id}>
                            {c.title}
                            {c.year ? ` (${c.year})` : ""}
                            {c.authors?.length ? ` — ${c.authors.slice(0, 2).join(", ")}` : ""}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Matched — collapsed by default */}
        {matchedRows.length > 0 && (
          <section>
            <button
              type="button"
              onClick={() => setMatchedExpanded((v) => !v)}
              className="flex items-center gap-2 text-sm font-semibold text-stone-700 hover:text-stone-900 transition-colors"
            >
              <span>Matched ({matchedRows.length})</span>
              <svg
                className={`w-4 h-4 transition-transform ${matchedExpanded ? "rotate-180" : ""}`}
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {matchedExpanded && (
              <ul className="mt-3 divide-y divide-stone-100 border border-stone-100 rounded-lg overflow-hidden">
                {matchedRows.map((row) => (
                  <li key={row.row_id} className="flex items-center gap-3 px-4 py-2.5 hover:bg-stone-50">
                    <input
                      type="checkbox"
                      checked={selected.get(row.row_id) ?? true}
                      onChange={(e) =>
                        setSelected((prev) => new Map(prev).set(row.row_id, e.target.checked))
                      }
                      className="rounded border-stone-300 accent-stone-900 shrink-0"
                    />
                    <div className="flex-1 min-w-0">
                      <span className="text-sm text-stone-900 truncate block">{row.title}</span>
                      <span className="text-xs text-stone-400">{row.author}</span>
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-xs text-stone-400">{shelfLabel(row.exclusive_shelf_slug)}</span>
                      {row.rating && (
                        <span className="block text-xs text-stone-300 font-mono">{stars(row.rating)}</span>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>
        )}

        {/* Unmatched */}
        {unmatchedRows.length > 0 && (
          <section>
            <h2 className="text-sm font-semibold text-stone-700 mb-1">
              Not found ({unmatchedRows.length})
            </h2>
            <p className="text-xs text-stone-400 mb-3">
              These books couldn&apos;t be matched to Open Library and will be skipped.
            </p>
            <ul className="divide-y divide-stone-100 border border-stone-100 rounded-lg overflow-hidden">
              {unmatchedRows.map((row) => (
                <li key={row.row_id} className="px-4 py-2.5">
                  <span className="text-sm text-stone-500 truncate block">{row.title}</span>
                  <span className="text-xs text-stone-400">{row.author}</span>
                </li>
              ))}
            </ul>
          </section>
        )}

        {error && <p className="text-sm text-red-600">{error}</p>}

        {/* Commit bar */}
        <div className="flex items-center gap-4 pt-2 border-t border-stone-100">
          <button
            type="button"
            onClick={handleCommit}
            disabled={selectedCount === 0}
            className="px-4 py-2 bg-stone-900 text-white text-sm rounded-lg hover:bg-stone-700 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Import {selectedCount} book{selectedCount !== 1 ? "s" : ""}
          </button>
          <button
            type="button"
            onClick={() => { setPhase("idle"); setPreview(null); setError(null); }}
            className="text-sm text-stone-400 hover:text-stone-700 transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  // idle
  return (
    <div className="space-y-6">
      {savedUnmatched.length > 0 && (
        <section className="border border-amber-200 rounded-lg p-4 bg-amber-50 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-amber-900">
              Not found in previous import ({savedUnmatched.length})
            </h2>
            <button
              type="button"
              onClick={clearAllUnmatched}
              className="text-xs text-amber-700 hover:text-amber-900 transition-colors"
            >
              Clear all
            </button>
          </div>
          <p className="text-xs text-amber-700">
            These books couldn&apos;t be matched automatically. Search for them manually to add them.
          </p>
          <ul className="divide-y divide-amber-200">
            {savedUnmatched.map((book, i) => (
              <li key={i} className="flex items-center gap-3 py-2">
                <div className="flex-1 min-w-0">
                  <span className="text-sm text-stone-800 truncate block">{book.title}</span>
                  <span className="text-xs text-stone-500">{book.author}</span>
                </div>
                <a
                  href={`/search?q=${encodeURIComponent(book.title)}`}
                  className="text-xs text-stone-500 hover:text-stone-900 underline shrink-0 transition-colors"
                >
                  Search
                </a>
                <button
                  type="button"
                  onClick={() => dismissUnmatched(i)}
                  className="text-xs text-stone-400 hover:text-stone-600 shrink-0 transition-colors"
                  aria-label="Dismiss"
                >
                  ✕
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      <div className="text-sm text-stone-500 space-y-1">
        <p>Export your library from Goodreads:</p>
        <p className="text-stone-400">
          My Books → Import and Export → Export Library → Download
        </p>
      </div>

      <div className="flex items-center gap-3">
        <label className="cursor-pointer">
          <span className="px-3 py-2 text-sm border border-stone-200 rounded-lg text-stone-700 hover:border-stone-400 transition-colors focus-within:ring-2 focus-within:ring-stone-900 focus-within:ring-offset-2 relative">
            {fileName ? "Change file" : "Choose file"}
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              className="sr-only"
              onChange={(e) => setFileName(e.target.files?.[0]?.name ?? null)}
            />
          </span>
        </label>
        {fileName && <span className="text-xs text-stone-400 truncate max-w-xs">{fileName}</span>}
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <button
        type="button"
        onClick={handlePreview}
        disabled={!fileName}
        className="px-4 py-2 bg-stone-900 text-white text-sm rounded-lg hover:bg-stone-700 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Preview import
      </button>
    </div>
  );
}
