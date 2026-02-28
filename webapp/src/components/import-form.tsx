"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import { useToast } from "@/components/toast";

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

type DoneResult = { imported: number; failed: number; errors: string[]; pending_saved?: number };

type PendingImport = {
  id: string;
  title: string;
  author: string;
  isbn13: string;
  exclusive_shelf: string;
  custom_shelves: string[];
  rating: number | null;
  review_text: string;
  date_read: string;
  date_added: string;
  created: string;
};

type ShelfMapping = {
  action: "tag" | "skip" | "create_label" | "existing_label" | "map_dnf";
  label_name?: string;
  label_key_id?: string;
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
};

// ── Helpers ────────────────────────────────────────────────────────────────────

function shelfLabel(slug: string): string {
  const labels: Record<string, string> = {
    "read": "Read",
    "to-read": "Want to Read",
    "currently-reading": "Currently Reading",
    "want-to-read": "Want to Read",
    "owned-to-read": "Owned to Read",
    "dnf": "Did Not Finish",
  };
  return labels[slug] ?? slug;
}

function stars(rating: number | null): string {
  if (!rating) return "";
  return "\u2605".repeat(rating) + "\u2606".repeat(5 - rating);
}

// ── Component ─────────────────────────────────────────────────────────────────

export default function ImportForm({ username, source = "goodreads" }: { username: string; source?: "goodreads" | "storygraph" }) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState<string | null>(null);
  const toast = useToast();

  type Phase = "idle" | "previewing" | "review" | "configure" | "importing" | "done";
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
  // shelf name → mapping config
  const [shelfMappings, setShelfMappings] = useState<Map<string, ShelfMapping>>(new Map());
  // shelves currently editing a new label name (input visible until Enter/blur)
  const [editingLabel, setEditingLabel] = useState<Set<string>>(new Set());
  // User's existing tag keys (for "Add to existing label")
  const [existingTagKeys, setExistingTagKeys] = useState<TagKey[]>([]);
  // Preview progress
  const [progress, setProgress] = useState<{ current: number; total: number; title: string } | null>(null);

  // Pending imports from server
  const [pendingImports, setPendingImports] = useState<PendingImport[]>([]);

  const loadPendingImports = useCallback(async () => {
    try {
      const res = await fetch("/api/me/imports/pending");
      if (res.ok) {
        const data = await res.json();
        setPendingImports(Array.isArray(data) ? data : []);
      }
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    loadPendingImports();
  }, [loadPendingImports]);

  async function dismissPending(id: string) {
    setPendingImports((prev) => prev.filter((p) => p.id !== id));
    try {
      await fetch(`/api/me/imports/pending/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "dismiss" }),
      });
    } catch {
      // ignore
    }
  }

  async function clearAllPending() {
    const ids = pendingImports.map((p) => p.id);
    setPendingImports([]);
    for (const id of ids) {
      try {
        await fetch(`/api/me/imports/pending/${id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ action: "dismiss" }),
        });
      } catch {
        // ignore
      }
    }
  }

  // ── Upload & preview ────────────────────────────────────────────────────────

  async function handlePreview() {
    const file = fileInputRef.current?.files?.[0];
    if (!file) {
      setError("Please choose a CSV file first.");
      return;
    }
    setError(null);
    setProgress(null);
    setPhase("previewing");

    const form = new FormData();
    form.append("file", file);

    try {
      const res = await fetch(`/api/me/import/${source}/preview`, {
        method: "POST",
        body: form,
      });

      if (!res.ok) {
        const text = await res.text();
        try {
          const data = JSON.parse(text);
          setError(data.error ?? "Preview failed.");
        } catch {
          setError("Preview failed.");
        }
        setPhase("idle");
        return;
      }

      // Read NDJSON stream
      const reader = res.body!.getReader();
      const decoder = new TextDecoder();
      let buffer = "";
      let result: (PreviewResponse & { shelves?: { name: string; count: number }[] }) | null = null;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop()!;
        for (const line of lines) {
          if (!line.trim()) continue;
          try {
            const obj = JSON.parse(line);
            if (obj.type === "progress") {
              setProgress({ current: obj.current, total: obj.total, title: obj.title });
            } else if (obj.type === "result") {
              result = obj;
            }
          } catch {
            // skip malformed lines
          }
        }
      }

      if (!result) {
        setError("Failed to parse preview response.");
        setPhase("idle");
        return;
      }

      // Pre-populate state: match/ambiguous rows selected, unmatched skipped.
      const sel = new Map<number, boolean>();
      const cho = new Map<number, string>();
      for (const row of result.rows) {
        sel.set(row.row_id, row.status !== "unmatched");
        if (row.status === "ambiguous" && row.candidates?.[0]) {
          cho.set(row.row_id, row.candidates[0].ol_id);
        }
      }
      setSelected(sel);
      setChoices(cho);
      setPreview(result);
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

  function buildUnmatchedRows() {
    if (!preview) return [];
    return preview.rows
      .filter((r) => r.status === "unmatched")
      .map((r) => ({
        title: r.title,
        author: r.author,
        isbn13: r.isbn13,
        rating: r.rating,
        review_text: r.review_text,
        date_read: r.date_read,
        date_added: r.date_added,
        exclusive_shelf_slug: r.exclusive_shelf_slug,
        custom_shelves: r.custom_shelves,
      }));
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
      const res = await fetch(`/api/me/import/${source}/commit`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          rows,
          unmatched_rows: buildUnmatchedRows(),
          shelf_mappings: Array.from(shelfMappings.entries()).map(([shelf, mapping]) => ({
            shelf,
            action: mapping.action,
            label_name: mapping.label_name ?? "",
            label_key_id: mapping.label_key_id ?? "",
          })),
        }),
      });
      const data: DoneResult & { error?: string } = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Import failed.");
        setPhase("review");
        return;
      }
      // Reload pending imports from server
      await loadPendingImports();
      setDoneResult(data);
      setPhase("done");
      toast.success(`Import complete — ${data.imported} book${data.imported !== 1 ? "s" : ""} imported`);
    } catch {
      setError("Network error. Please try again.");
      toast.error("Import failed — network error");
      setPhase("review");
    }
  }

  async function enterConfigure() {
    const rows = buildCommitRows();
    // Collect unique custom shelves from selected rows with counts
    const counts = new Map<string, number>();
    for (const row of rows) {
      for (const shelf of row.custom_shelves) {
        counts.set(shelf, (counts.get(shelf) ?? 0) + 1);
      }
    }
    // Initialize all custom shelves to "tag" by default
    const mappings = new Map<string, ShelfMapping>();
    for (const shelf of counts.keys()) {
      mappings.set(shelf, shelfMappings.get(shelf) ?? { action: "tag" });
    }
    setShelfMappings(mappings);
    // Fetch existing tag keys for "Add to existing label"
    try {
      const res = await fetch("/api/me/tag-keys");
      if (res.ok) {
        const data: TagKey[] = await res.json();
        setExistingTagKeys(
          (data ?? []).filter((k) => k.slug !== "status"),
        );
      }
    } catch {
      // ignore — existing labels just won't be available
    }
    setPhase("configure");
  }

  // ── Counts ──────────────────────────────────────────────────────────────────

  const selectedCount = buildCommitRows().length;

  // ── Render ───────────────────────────────────────────────────────────────────

  if (phase === "done" && doneResult) {
    return (
      <div className="space-y-4">
        <p className="text-text-primary font-medium">
          {doneResult.imported} book{doneResult.imported !== 1 ? "s" : ""} imported successfully.
        </p>
        {doneResult.failed > 0 && (
          <p className="text-sm text-red-600">
            {doneResult.failed} book{doneResult.failed !== 1 ? "s" : ""} failed to import.
          </p>
        )}
        {(doneResult.pending_saved ?? 0) > 0 && (
          <p className="text-sm text-text-primary">
            {doneResult.pending_saved} unmatched book{doneResult.pending_saved !== 1 ? "s" : ""} saved for later.
          </p>
        )}
        {(doneResult.errors ?? []).length > 0 && (
          <details className="text-xs text-text-primary">
            <summary className="cursor-pointer hover:text-text-primary">Show errors</summary>
            <ul className="mt-2 space-y-1 list-disc list-inside">
              {doneResult.errors.map((e, i) => <li key={i}>{e}</li>)}
            </ul>
          </details>
        )}
        <Link
          href={`/${username}/library`}
          className="inline-block mt-2 text-sm text-text-primary underline hover:no-underline"
        >
          Go to your library
        </Link>
      </div>
    );
  }

  if (phase === "previewing" || phase === "importing") {
    return (
      <div className="space-y-2">
        <div className="flex items-center gap-3 text-sm text-text-primary">
          <svg className="animate-spin h-4 w-4 text-text-primary shrink-0" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {phase === "previewing"
            ? progress
              ? `Matching books (${progress.current}/${progress.total})\u2026`
              : "Matching books against Open Library\u2026"
            : "Importing\u2026"}
        </div>
        {phase === "previewing" && progress && (
          <p className="text-xs text-text-primary pl-7 truncate">
            Last looked up: {progress.title}
          </p>
        )}
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
            <span className="px-2.5 py-1 rounded-full bg-surface-2 border border-border text-text-primary">
              {preview.unmatched} not found
            </span>
          )}
        </div>

        {/* Ambiguous — always expanded (requires user input) */}
        {ambiguousRows.length > 0 && (
          <section>
            <h2 className="text-sm font-semibold text-text-primary mb-3">
              Choose edition ({ambiguousRows.length})
            </h2>
            <p className="text-xs text-text-primary mb-4">
              Multiple editions found — pick the one you read.
            </p>
            <ul className="space-y-4">
              {ambiguousRows.map((row) => (
                <li key={row.row_id} className="border border-border rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <input
                      type="checkbox"
                      checked={selected.get(row.row_id) ?? true}
                      onChange={(e) =>
                        setSelected((prev) => new Map(prev).set(row.row_id, e.target.checked))
                      }
                      className="mt-0.5 rounded border-border accent-neutral-400"
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-text-primary truncate">{row.title}</p>
                      <p className="text-xs text-text-primary">{row.author}</p>
                      {row.rating && (
                        <p className="text-xs text-text-primary mt-0.5 font-mono">{stars(row.rating)}</p>
                      )}
                      <select
                        className="mt-2 text-xs border border-border rounded px-2 py-1.5 w-full max-w-md text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong"
                        value={choices.get(row.row_id) ?? row.candidates?.[0]?.ol_id ?? ""}
                        onChange={(e) =>
                          setChoices((prev) => new Map(prev).set(row.row_id, e.target.value))
                        }
                      >
                        {(row.candidates ?? []).map((c) => (
                          <option key={c.ol_id} value={c.ol_id}>
                            {c.title}
                            {c.year ? ` (${c.year})` : ""}
                            {c.authors?.length ? ` \u2014 ${c.authors.slice(0, 2).join(", ")}` : ""}
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
              className="flex items-center gap-2 text-sm font-semibold text-text-primary hover:text-text-primary transition-colors"
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
              <ul className="mt-3 divide-y divide-border border border-border rounded-lg overflow-hidden">
                {matchedRows.map((row) => (
                  <li key={row.row_id} className="flex items-center gap-3 px-4 py-2.5 hover:bg-surface-2">
                    <input
                      type="checkbox"
                      checked={selected.get(row.row_id) ?? true}
                      onChange={(e) =>
                        setSelected((prev) => new Map(prev).set(row.row_id, e.target.checked))
                      }
                      className="rounded border-border accent-neutral-400 shrink-0"
                    />
                    <div className="flex-1 min-w-0">
                      <span className="text-sm text-text-primary truncate block">{row.title}</span>
                      <span className="text-xs text-text-primary">{row.author}</span>
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-xs text-text-primary">{shelfLabel(row.exclusive_shelf_slug)}</span>
                      {row.rating && (
                        <span className="block text-xs text-text-primary font-mono">{stars(row.rating)}</span>
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
            <h2 className="text-sm font-semibold text-text-primary mb-1">
              Not found ({unmatchedRows.length})
            </h2>
            <p className="text-xs text-text-primary mb-3">
              These books couldn&apos;t be matched to Open Library and will be saved for later.
            </p>
            <ul className="divide-y divide-border border border-border rounded-lg overflow-hidden">
              {unmatchedRows.map((row) => (
                <li key={row.row_id} className="px-4 py-2.5">
                  <span className="text-sm text-text-primary truncate block">{row.title}</span>
                  <span className="text-xs text-text-primary">{row.author}</span>
                </li>
              ))}
            </ul>
          </section>
        )}

        {error && <p className="text-sm text-red-600">{error}</p>}

        {/* Action bar */}
        <div className="flex items-center gap-4 pt-2 border-t border-border">
          <button
            type="button"
            onClick={enterConfigure}
            disabled={selectedCount === 0}
            className="px-4 py-2 bg-accent text-text-inverted text-sm rounded-lg hover:bg-accent-hover transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Next: Configure labels
          </button>
          <button
            type="button"
            onClick={() => { setPhase("idle"); setPreview(null); setError(null); }}
            className="text-sm text-text-primary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  if (phase === "configure") {
    // Compute shelf counts from selected rows
    const rows = buildCommitRows();
    const shelfCounts = new Map<string, number>();
    for (const row of rows) {
      for (const shelf of row.custom_shelves) {
        shelfCounts.set(shelf, (shelfCounts.get(shelf) ?? 0) + 1);
      }
    }
    const sortedShelves = Array.from(shelfCounts.entries()).sort((a, b) => b[1] - a[1]);
    const hasCustomShelves = sortedShelves.length > 0;

    return (
      <div className="space-y-8">
        {/* Exclusive shelf mapping (read-only, informational) */}
        <section>
          <h2 className="text-sm font-semibold text-text-primary mb-3">
            Status mapping
          </h2>
          <p className="text-xs text-text-primary mb-3">
            Reading status will be mapped to your Status label automatically.
          </p>
          <ul className="divide-y divide-border border border-border rounded-lg overflow-hidden text-sm">
            <li className="flex items-center justify-between px-4 py-2.5">
              <span className="text-text-primary">read</span>
              <span className="text-text-primary">Status: Finished</span>
            </li>
            <li className="flex items-center justify-between px-4 py-2.5">
              <span className="text-text-primary">currently-reading</span>
              <span className="text-text-primary">Status: Currently Reading</span>
            </li>
            <li className="flex items-center justify-between px-4 py-2.5">
              <span className="text-text-primary">to-read</span>
              <span className="text-text-primary">Status: Want to Read</span>
            </li>
            {source === "storygraph" && (
              <li className="flex items-center justify-between px-4 py-2.5">
                <span className="text-text-primary">did-not-finish</span>
                <span className="text-text-primary">Status: DNF</span>
              </li>
            )}
          </ul>
        </section>

        {/* Custom shelf mapping */}
        {hasCustomShelves ? (
          <section>
            <h2 className="text-sm font-semibold text-text-primary mb-3">
              Custom labels ({sortedShelves.length})
            </h2>
            <p className="text-xs text-text-primary mb-3">
              Choose how to import each {source === "storygraph" ? "StoryGraph tag" : "custom Goodreads shelf"} as a label.
            </p>
            <ul className="divide-y divide-border border border-border rounded-lg overflow-hidden">
              {(() => {
                // Collect all label names entered across all shelves (sorted for stable order)
                const labelNameSet = new Set<string>();
                for (const [, m] of shelfMappings) {
                  if (m.action === "create_label" && m.label_name?.trim()) {
                    labelNameSet.add(m.label_name.trim());
                  }
                }
                const allLabelNames = Array.from(labelNameSet).sort();
                return sortedShelves.map(([shelf, count]) => {
                  const mapping = shelfMappings.get(shelf) ?? { action: "tag" as const };
                  const isEditing = editingLabel.has(shelf);
                  const showInput = isEditing || (mapping.action === "create_label" && !mapping.label_name?.trim());
                  return (
                    <li key={shelf} className="px-4 py-2.5 space-y-2">
                      <div className="flex items-center gap-3">
                        <div className="flex-1 min-w-0">
                          <span className="text-sm text-text-primary">{shelf}</span>
                          <span className="text-xs text-text-tertiary ml-2">
                            ({count} book{count !== 1 ? "s" : ""})
                          </span>
                        </div>
                        <select
                          className="text-xs border border-border rounded px-2 py-1.5 text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong"
                          value={mapping.action}
                          onChange={(e) => {
                            const action = e.target.value as ShelfMapping["action"];
                            setShelfMappings((prev) => {
                              const next = new Map(prev);
                              next.set(shelf, { action });
                              return next;
                            });
                            if (action === "create_label") {
                              setEditingLabel((prev) => new Set(prev).add(shelf));
                            } else {
                              setEditingLabel((prev) => { const next = new Set(prev); next.delete(shelf); return next; });
                            }
                          }}
                        >
                          <option value="tag">Tag</option>
                          <option value="create_label">Create label</option>
                          {existingTagKeys.length > 0 && (
                            <option value="existing_label">Add to existing label</option>
                          )}
                          <option value="map_dnf">Map to DNF</option>
                          <option value="skip">Skip</option>
                        </select>
                      </div>
                      {/* Label pills — always visible when labels exist */}
                      {allLabelNames.length > 0 && mapping.action !== "existing_label" && mapping.action !== "skip" && mapping.action !== "map_dnf" && !showInput && (
                        <div className="flex items-center gap-1.5 flex-wrap">
                          {allLabelNames.map((name) => {
                            const isActive = mapping.action === "create_label" && mapping.label_name?.trim() === name;
                            return (
                              <button
                                key={name}
                                type="button"
                                className={isActive ? "tag-pill-active cursor-pointer" : "tag-pill cursor-pointer"}
                                onClick={() =>
                                  setShelfMappings((prev) => {
                                    const next = new Map(prev);
                                    if (isActive) {
                                      next.set(shelf, { action: "tag" });
                                    } else {
                                      next.set(shelf, { action: "create_label", label_name: name });
                                    }
                                    return next;
                                  })
                                }
                              >
                                {name}
                              </button>
                            );
                          })}
                        </div>
                      )}
                      {/* Text input for creating a new label name */}
                      {showInput && (
                        <div className="flex items-center gap-2">
                          <label className="text-xs text-text-primary shrink-0">Label name:</label>
                          <input
                            type="text"
                            placeholder="e.g. Genre, Source"
                            autoFocus
                            className="text-xs border border-border rounded px-2 py-1.5 flex-1 max-w-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong"
                            value={mapping.label_name ?? ""}
                            onChange={(e) =>
                              setShelfMappings((prev) => {
                                const next = new Map(prev);
                                next.set(shelf, { action: "create_label", label_name: e.target.value });
                                return next;
                              })
                            }
                            onKeyDown={(e) => {
                              if (e.key === "Enter") {
                                e.preventDefault();
                                (e.target as HTMLInputElement).blur();
                              }
                            }}
                            onBlur={() => {
                              setEditingLabel((prev) => { const next = new Set(prev); next.delete(shelf); return next; });
                              // If blurred with empty value, revert to tag
                              if (!mapping.label_name?.trim()) {
                                setShelfMappings((prev) => {
                                  const next = new Map(prev);
                                  next.set(shelf, { action: "tag" });
                                  return next;
                                });
                              }
                            }}
                          />
                          <span className="text-xs text-text-tertiary">
                            (becomes a value under this label)
                          </span>
                        </div>
                      )}
                      {mapping.action === "existing_label" && (
                        <div className="flex items-center gap-2">
                          <label className="text-xs text-text-primary shrink-0">Label:</label>
                          <select
                            className="text-xs border border-border rounded px-2 py-1.5 text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong"
                            value={mapping.label_key_id ?? ""}
                            onChange={(e) =>
                              setShelfMappings((prev) => {
                                const next = new Map(prev);
                                next.set(shelf, { ...mapping, label_key_id: e.target.value });
                                return next;
                              })
                            }
                          >
                            <option value="">Select a label...</option>
                            {existingTagKeys.map((k) => (
                              <option key={k.id} value={k.id}>{k.name}</option>
                            ))}
                          </select>
                          <span className="text-xs text-text-tertiary">
                            (becomes a value under this label)
                          </span>
                        </div>
                      )}
                    </li>
                  );
                });
              })()}
            </ul>
          </section>
        ) : (
          <p className="text-sm text-text-primary">
            No custom labels found — only status mapping will be applied.
          </p>
        )}

        {error && <p className="text-sm text-red-600">{error}</p>}

        {/* Action bar */}
        <div className="flex items-center gap-4 pt-2 border-t border-border">
          <button
            type="button"
            onClick={handleCommit}
            className="px-4 py-2 bg-accent text-text-inverted text-sm rounded-lg hover:bg-accent-hover transition-colors"
          >
            Import {selectedCount} book{selectedCount !== 1 ? "s" : ""}
          </button>
          <button
            type="button"
            onClick={() => setPhase("review")}
            className="text-sm text-text-primary hover:text-text-primary transition-colors"
          >
            Back
          </button>
        </div>
      </div>
    );
  }

  // idle
  return (
    <div className="space-y-6">
      {pendingImports.length > 0 && (
        <section className="border border-amber-200 rounded-lg p-4 bg-amber-50 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-amber-900">
              Failed imports ({pendingImports.length})
            </h2>
            <button
              type="button"
              onClick={clearAllPending}
              className="text-xs text-amber-700 hover:text-amber-900 transition-colors"
            >
              Dismiss all
            </button>
          </div>
          <p className="text-xs text-amber-700">
            These books couldn&apos;t be matched automatically. Search for them manually to add them.
          </p>
          <ul className="divide-y divide-amber-200">
            {pendingImports.map((item) => (
              <li key={item.id} className="flex items-center gap-3 py-2">
                <div className="flex-1 min-w-0">
                  <span className="text-sm text-text-primary truncate block">{item.title}</span>
                  <span className="text-xs text-text-primary">{item.author}</span>
                </div>
                <a
                  href={`/search?q=${encodeURIComponent(item.title)}`}
                  className="text-xs text-text-primary hover:text-text-primary underline shrink-0 transition-colors"
                >
                  Search
                </a>
                <button
                  type="button"
                  onClick={() => dismissPending(item.id)}
                  className="text-xs text-text-primary hover:text-text-primary shrink-0 transition-colors"
                  aria-label="Dismiss"
                >
                  &#10005;
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      <div className="text-sm text-text-primary space-y-1">
        {source === "storygraph" ? (
          <>
            <p>Export your library from StoryGraph:</p>
            <p className="text-text-primary">
              Settings &rarr; Manage Account &rarr; Export StoryGraph Library
            </p>
          </>
        ) : (
          <>
            <p>Export your library from Goodreads:</p>
            <p className="text-text-primary">
              My Books &rarr; Import and Export &rarr; Export Library &rarr; Download
            </p>
          </>
        )}
      </div>

      <div className="flex items-center gap-3">
        <label className="cursor-pointer">
          <span className="px-3 py-2 text-sm border border-border rounded-lg text-text-primary hover:border-border transition-colors focus-within:ring-2 focus-within:ring-accent focus-within:ring-offset-2 relative">
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
        {fileName && <span className="text-xs text-text-primary truncate max-w-xs">{fileName}</span>}
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <button
        type="button"
        onClick={handlePreview}
        disabled={!fileName}
        className="px-4 py-2 bg-accent text-text-inverted text-sm rounded-lg hover:bg-accent-hover transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Preview import
      </button>
    </div>
  );
}
