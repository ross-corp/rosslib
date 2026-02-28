"use client";

import { useState } from "react";
import Link from "next/link";

type SavedSearch = {
  id: string;
  name: string;
  query: string;
  filters: {
    sort?: string;
    year_min?: string;
    year_max?: string;
    subject?: string;
    language?: string;
    tab?: string;
  } | null;
  created_at: string;
};

function buildHref(search: SavedSearch): string {
  const p = new URLSearchParams();
  p.set("q", search.query);
  if (search.filters) {
    if (search.filters.tab) p.set("type", search.filters.tab);
    if (search.filters.sort) p.set("sort", search.filters.sort);
    if (search.filters.year_min) p.set("year_min", search.filters.year_min);
    if (search.filters.year_max) p.set("year_max", search.filters.year_max);
    if (search.filters.subject) p.set("subject", search.filters.subject);
    if (search.filters.language) p.set("language", search.filters.language);
  }
  return `/search?${p}`;
}

export default function SavedSearches({
  searches: initialSearches,
  currentQuery,
  currentFilters,
}: {
  searches: SavedSearch[];
  currentQuery: string;
  currentFilters: {
    sort: string;
    year_min: string;
    year_max: string;
    subject: string;
    language: string;
    tab: string;
  };
}) {
  const [searches, setSearches] = useState(initialSearches);
  const [saving, setSaving] = useState(false);
  const [showNameInput, setShowNameInput] = useState(false);
  const [name, setName] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const hasActiveFilters =
    currentQuery.trim() !== "" &&
    (currentFilters.sort ||
      currentFilters.year_min ||
      currentFilters.year_max ||
      currentFilters.subject ||
      currentFilters.language);

  const canSave = currentQuery.trim() !== "";

  async function handleSave() {
    if (!name.trim()) return;
    setSaving(true);
    const filters: Record<string, string> = {};
    if (currentFilters.sort) filters.sort = currentFilters.sort;
    if (currentFilters.year_min) filters.year_min = currentFilters.year_min;
    if (currentFilters.year_max) filters.year_max = currentFilters.year_max;
    if (currentFilters.subject) filters.subject = currentFilters.subject;
    if (currentFilters.language) filters.language = currentFilters.language;
    if (currentFilters.tab) filters.tab = currentFilters.tab;

    const res = await fetch("/api/me/saved-searches", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: name.trim(),
        query: currentQuery,
        filters: Object.keys(filters).length > 0 ? filters : null,
      }),
    });
    if (res.ok) {
      const data = await res.json();
      setSearches((prev) => [data, ...prev]);
      setName("");
      setShowNameInput(false);
    }
    setSaving(false);
  }

  async function handleDelete(id: string) {
    setDeletingId(id);
    const res = await fetch(`/api/me/saved-searches/${id}`, {
      method: "DELETE",
    });
    if (res.ok) {
      setSearches((prev) => prev.filter((s) => s.id !== id));
    }
    setDeletingId(null);
  }

  return (
    <div className="mb-6">
      {/* Saved search chips */}
      {searches.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap mb-3">
          <span className="text-xs text-text-secondary">Saved:</span>
          {searches.map((search) => (
            <div key={search.id} className="flex items-center gap-0.5 group">
              <Link
                href={buildHref(search)}
                className="text-xs px-2.5 py-1 rounded-l-full border border-r-0 border-border text-text-primary hover:border-border-strong hover:text-text-primary transition-colors bg-surface-2"
              >
                {search.name}
              </Link>
              <button
                onClick={() => handleDelete(search.id)}
                disabled={deletingId === search.id}
                className="text-xs px-1.5 py-1 rounded-r-full border border-border text-text-tertiary hover:text-text-primary hover:border-border-strong transition-colors bg-surface-2 disabled:opacity-50"
                title="Remove saved search"
              >
                &times;
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Save button */}
      {canSave && !showNameInput && (
        <button
          onClick={() => setShowNameInput(true)}
          className="text-xs text-text-secondary hover:text-text-primary underline transition-colors"
        >
          {hasActiveFilters
            ? "Save this search with filters"
            : "Save this search"}
        </button>
      )}

      {/* Name input */}
      {showNameInput && (
        <div className="flex items-center gap-2 mt-1">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleSave();
              if (e.key === "Escape") {
                setShowNameInput(false);
                setName("");
              }
            }}
            placeholder="Name this search..."
            maxLength={100}
            autoFocus
            className="px-2 py-1 text-xs border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent w-48"
          />
          <button
            onClick={handleSave}
            disabled={saving || !name.trim()}
            className="text-xs px-2.5 py-1 rounded border border-accent bg-accent text-text-inverted hover:bg-accent-hover transition-colors disabled:opacity-50"
          >
            {saving ? "..." : "Save"}
          </button>
          <button
            onClick={() => {
              setShowNameInput(false);
              setName("");
            }}
            className="text-xs text-text-tertiary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
        </div>
      )}
    </div>
  );
}
