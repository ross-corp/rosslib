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
  if (lower.includes("hardcover") || lower.includes("hardback") || lower === "capa dura")
    return "Hardcover";
  if (lower.includes("paperback") || lower.includes("softcover") || lower === "mass market")
    return "Paperback";
  if (lower.includes("ebook") || lower.includes("e-book") || lower.includes("kindle"))
    return "eBook";
  if (lower.includes("audio")) return "Audiobook";
  return format.charAt(0).toUpperCase() + format.slice(1);
}

export default function EditionPicker({
  openLibraryId,
  workId,
  initialEditions,
  totalEditions,
  currentEditionKey,
  currentEditionCoverUrl,
  onEditionChanged,
}: {
  openLibraryId: string;
  workId: string;
  initialEditions: Edition[];
  totalEditions: number;
  currentEditionKey: string | null;
  currentEditionCoverUrl: string | null;
  onEditionChanged?: (key: string | null, coverUrl: string | null) => void;
}) {
  const [open, setOpen] = useState(false);
  const [editions, setEditions] = useState<Edition[]>(initialEditions);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState<string | null>(null);
  const [selectedKey, setSelectedKey] = useState(currentEditionKey);
  const [selectedCoverUrl, setSelectedCoverUrl] = useState(currentEditionCoverUrl);

  const hasMore = totalEditions > editions.length;

  async function loadMore() {
    setLoading(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/editions?limit=50&offset=${editions.length}`
      );
      if (res.ok) {
        const data = await res.json();
        if (data.entries) {
          setEditions((prev) => [...prev, ...data.entries]);
        }
      }
    } finally {
      setLoading(false);
    }
  }

  async function selectEdition(editionKey: string | null, coverUrl: string | null) {
    setSaving(editionKey ?? "reset");
    try {
      const res = await fetch(`/api/me/books/${openLibraryId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          selected_edition_key: editionKey ?? "",
          selected_edition_cover_url: coverUrl ?? "",
        }),
      });
      if (res.ok) {
        setSelectedKey(editionKey);
        setSelectedCoverUrl(coverUrl);
        onEditionChanged?.(editionKey, coverUrl);
        setOpen(false);
      }
    } finally {
      setSaving(null);
    }
  }

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="text-xs text-text-secondary hover:text-text-primary transition-colors underline underline-offset-2"
      >
        Change edition
      </button>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setOpen(false)}
          />
          <div className="relative bg-surface-0 border border-border rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[80vh] flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-border">
              <h3 className="text-sm font-semibold text-text-primary">
                Choose edition
              </h3>
              <button
                onClick={() => setOpen(false)}
                className="text-text-secondary hover:text-text-primary text-lg leading-none"
              >
                &times;
              </button>
            </div>

            {/* Edition list */}
            <div className="overflow-y-auto flex-1 p-4 space-y-2">
              {/* Reset to default */}
              {selectedKey && (
                <button
                  onClick={() => selectEdition(null, null)}
                  disabled={saving !== null}
                  className="w-full flex items-center gap-3 p-3 rounded-lg border border-border hover:border-text-secondary transition-colors text-left disabled:opacity-50"
                >
                  <div className="w-10 h-[60px] shrink-0 rounded bg-surface-2 flex items-center justify-center">
                    <span className="text-[10px] text-text-tertiary">Default</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <span className="text-sm font-medium text-text-primary">
                      Use default cover
                    </span>
                    <p className="text-xs text-text-secondary mt-0.5">
                      Reset to the work&apos;s original cover
                    </p>
                  </div>
                  {saving === "reset" && (
                    <span className="text-xs text-text-secondary">Saving...</span>
                  )}
                </button>
              )}

              {editions.map((ed) => {
                const isSelected = selectedKey === ed.key;
                return (
                  <button
                    key={ed.key}
                    onClick={() => selectEdition(ed.key, ed.cover_url ?? null)}
                    disabled={saving !== null}
                    className={`w-full flex items-center gap-3 p-3 rounded-lg border transition-colors text-left disabled:opacity-50 ${
                      isSelected
                        ? "border-text-primary bg-surface-2/50"
                        : "border-border hover:border-text-secondary"
                    }`}
                  >
                    {/* Cover thumbnail */}
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
                          <span className="text-[10px] font-medium text-text-primary border border-text-primary rounded px-1.5 py-0.5 leading-none shrink-0">
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
                        <p className="text-[11px] text-text-tertiary mt-0.5 font-mono">
                          ISBN {ed.isbn}
                        </p>
                      )}
                    </div>

                    {saving === ed.key && (
                      <span className="text-xs text-text-secondary shrink-0">Saving...</span>
                    )}
                  </button>
                );
              })}

              {/* Load more */}
              {hasMore && (
                <button
                  onClick={loadMore}
                  disabled={loading}
                  className="w-full py-2 text-xs text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50"
                >
                  {loading
                    ? "Loading..."
                    : `Load more editions (${totalEditions - editions.length} remaining)`}
                </button>
              )}

              {editions.length === 0 && (
                <p className="text-sm text-text-secondary text-center py-4">
                  No editions available for this work.
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
