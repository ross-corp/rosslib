"use client";

import { useState } from "react";

type Edition = {
  key: string;
  title: string;
  publisher: string | null;
  publish_date: string;
  page_count: number | null;
  isbn: string | null;
  cover_url: string | null;
  format: string;
  language: string;
};

const LANG_NAMES: Record<string, string> = {
  eng: "English",
  spa: "Spanish",
  fre: "French",
  ger: "German",
  por: "Portuguese",
  ita: "Italian",
  dut: "Dutch",
  rus: "Russian",
  jpn: "Japanese",
  chi: "Chinese",
  kor: "Korean",
};

function langName(code: string): string {
  return LANG_NAMES[code] ?? code;
}

function formatLabel(format: string): string {
  if (!format) return "";
  const lower = format.toLowerCase();
  if (lower.includes("hardcover") || lower.includes("hardback")) return "Hardcover";
  if (lower.includes("paperback") || lower.includes("softcover") || lower === "mass market") return "Paperback";
  if (lower.includes("ebook") || lower.includes("e-book") || lower.includes("kindle")) return "eBook";
  if (lower.includes("audio")) return "Audiobook";
  return format.charAt(0).toUpperCase() + format.slice(1);
}

export default function EditionSelector({
  workId,
  openLibraryId,
  editions,
  totalEditions,
  currentEditionKey,
  defaultCoverUrl,
  onEditionChanged,
}: {
  workId: string;
  openLibraryId: string;
  editions: Edition[];
  totalEditions: number;
  currentEditionKey: string | null;
  defaultCoverUrl: string | null;
  onEditionChanged?: (editionKey: string | null, coverUrl: string | null) => void;
}) {
  const [open, setOpen] = useState(false);
  const [allEditions, setAllEditions] = useState<Edition[]>(editions);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState<string | null>(null);
  const [selectedKey, setSelectedKey] = useState<string | null>(currentEditionKey);

  async function loadMore() {
    setLoading(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/editions?limit=50&offset=${allEditions.length}`
      );
      if (res.ok) {
        const data = await res.json();
        if (data.editions) {
          setAllEditions((prev) => [...prev, ...data.editions]);
        }
      }
    } finally {
      setLoading(false);
    }
  }

  async function selectEdition(edition: Edition | null) {
    const editionKey = edition?.key ?? "";
    const coverUrl = edition?.cover_url ?? "";
    setSaving(editionKey || "__reset");
    try {
      const res = await fetch(`/api/me/books/${openLibraryId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          selected_edition_key: editionKey,
          selected_edition_cover_url: coverUrl,
        }),
      });
      if (res.ok) {
        setSelectedKey(editionKey || null);
        onEditionChanged?.(editionKey || null, coverUrl || null);
        setOpen(false);
      }
    } finally {
      setSaving(null);
    }
  }

  const hasMore = totalEditions > allEditions.length;

  return (
    <div>
      <button
        onClick={() => setOpen(!open)}
        className="text-xs text-text-secondary hover:text-text-primary transition-colors"
      >
        {selectedKey ? "Change edition" : "Select edition"}
      </button>

      {open && (
        <div className="mt-3 border border-border rounded-lg bg-surface-0 max-h-[400px] overflow-y-auto">
          {/* Reset to default option */}
          {selectedKey && (
            <button
              onClick={() => selectEdition(null)}
              disabled={saving !== null}
              className="w-full flex items-center gap-3 p-3 hover:bg-surface-2/50 transition-colors border-b border-border text-left"
            >
              {defaultCoverUrl ? (
                <img
                  src={defaultCoverUrl}
                  alt="Default cover"
                  className="w-10 h-[60px] shrink-0 rounded object-cover bg-surface-2"
                />
              ) : (
                <div className="w-10 h-[60px] shrink-0 rounded bg-surface-2" />
              )}
              <div className="flex-1 min-w-0">
                <span className="text-sm font-medium text-text-primary">
                  Default cover
                </span>
                <p className="text-xs text-text-secondary">
                  Use the original work cover
                </p>
              </div>
              {saving === "__reset" && (
                <span className="text-xs text-text-secondary">Saving...</span>
              )}
            </button>
          )}

          {allEditions.map((ed) => {
            const isSelected = selectedKey === ed.key;
            return (
              <button
                key={ed.key}
                onClick={() => selectEdition(ed)}
                disabled={saving !== null || isSelected}
                className={`w-full flex items-center gap-3 p-3 hover:bg-surface-2/50 transition-colors border-b border-border last:border-b-0 text-left ${
                  isSelected ? "bg-surface-2/80" : ""
                }`}
              >
                {ed.cover_url ? (
                  <img
                    src={ed.cover_url}
                    alt={ed.title}
                    className="w-10 h-[60px] shrink-0 rounded object-cover bg-surface-2"
                  />
                ) : (
                  <div className="w-10 h-[60px] shrink-0 rounded bg-surface-2" />
                )}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-text-primary truncate">
                      {ed.title || "Untitled"}
                    </span>
                    {ed.format && (
                      <span className="text-[10px] font-medium text-text-secondary border border-border rounded px-1.5 py-0.5 leading-none shrink-0">
                        {formatLabel(ed.format)}
                      </span>
                    )}
                    {isSelected && (
                      <span className="text-[10px] font-medium text-text-primary border border-border rounded px-1.5 py-0.5 leading-none shrink-0 bg-surface-2">
                        Selected
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-text-secondary mt-0.5">
                    {[
                      ed.publisher,
                      ed.publish_date,
                      ed.page_count ? `${ed.page_count} pp` : null,
                      ed.language ? langName(ed.language) : null,
                    ]
                      .filter(Boolean)
                      .join(" · ")}
                  </p>
                  {ed.isbn && (
                    <p className="text-[11px] text-text-secondary mt-0.5 font-mono">
                      ISBN {ed.isbn}
                    </p>
                  )}
                </div>
                {saving === ed.key && (
                  <span className="text-xs text-text-secondary shrink-0">
                    Saving...
                  </span>
                )}
              </button>
            );
          })}

          {hasMore && (
            <div className="p-3 text-center">
              <button
                onClick={loadMore}
                disabled={loading}
                className="text-xs text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50"
              >
                {loading
                  ? "Loading..."
                  : `Load more editions (${totalEditions - allEditions.length} remaining)`}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
