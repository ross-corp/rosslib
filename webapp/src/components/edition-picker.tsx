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
  if (lower.includes("paperback") || lower.includes("softcover") || lower === "mass market")
    return "Paperback";
  if (lower.includes("ebook") || lower.includes("e-book") || lower.includes("kindle"))
    return "eBook";
  if (lower.includes("audio")) return "Audiobook";
  return format.charAt(0).toUpperCase() + format.slice(1);
}

export default function EditionPicker({
  workId,
  openLibraryId,
  initialEditions,
  totalEditions,
  currentEditionKey,
}: {
  workId: string;
  openLibraryId: string;
  initialEditions: Edition[];
  totalEditions: number;
  currentEditionKey: string | null;
}) {
  const [open, setOpen] = useState(false);
  const [editions, setEditions] = useState<Edition[]>(initialEditions);
  const [selected, setSelected] = useState<string | null>(currentEditionKey);
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(false);

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
          const parsed: Edition[] = (data.entries as Record<string, unknown>[]).map(
            (e) => {
              const key = ((e.key as string) ?? "").replace("/books/", "");
              let coverUrl: string | null = null;
              if (Array.isArray(e.covers) && (e.covers as number[]).length > 0) {
                coverUrl = `https://covers.openlibrary.org/b/id/${(e.covers as number[])[0]}-M.jpg`;
              }
              let isbn: string | null = null;
              if (Array.isArray(e.isbn_13) && (e.isbn_13 as string[]).length > 0) {
                isbn = (e.isbn_13 as string[])[0];
              } else if (Array.isArray(e.isbn_10) && (e.isbn_10 as string[]).length > 0) {
                isbn = (e.isbn_10 as string[])[0];
              }
              const pubs = e.publishers as string[] | undefined;
              const langs = e.languages as { key: string }[] | undefined;
              return {
                key,
                title: (e.title as string) ?? "",
                publisher: pubs?.[0] ?? null,
                publish_date: (e.publish_date as string) ?? "",
                page_count: (e.number_of_pages as number) ?? null,
                isbn,
                cover_url: coverUrl,
                format: (e.physical_format as string) ?? "",
                language: langs?.[0]?.key?.replace("/languages/", "") ?? "",
              };
            }
          );
          setEditions((prev) => [...prev, ...parsed]);
        }
      }
    } finally {
      setLoading(false);
    }
  }

  async function saveEdition(editionKey: string | null) {
    setSaving(true);
    try {
      const res = await fetch(`/api/me/books/${openLibraryId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ selected_edition_key: editionKey ?? "" }),
      });
      if (res.ok) {
        setSelected(editionKey);
        setOpen(false);
        window.location.reload();
      }
    } finally {
      setSaving(false);
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
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setOpen(false)}
          />

          {/* Modal */}
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
              {/* Reset to default option */}
              {selected && (
                <button
                  onClick={() => saveEdition(null)}
                  disabled={saving}
                  className="w-full text-left p-3 rounded-lg border border-border hover:border-text-secondary transition-colors text-xs text-text-secondary disabled:opacity-50"
                >
                  Reset to default cover
                </button>
              )}

              {editions.map((ed) => {
                const isSelected = selected === ed.key;
                return (
                  <button
                    key={ed.key}
                    onClick={() => saveEdition(ed.key)}
                    disabled={saving}
                    className={`w-full text-left flex gap-3 p-3 rounded-lg border transition-colors disabled:opacity-50 ${
                      isSelected
                        ? "border-text-primary bg-surface-2"
                        : "border-border hover:border-text-secondary"
                    }`}
                  >
                    {/* Cover thumbnail */}
                    {ed.cover_url ? (
                      <img
                        src={ed.cover_url}
                        alt={ed.title}
                        className="w-12 h-[72px] shrink-0 rounded object-cover bg-surface-2"
                      />
                    ) : (
                      <div className="w-12 h-[72px] shrink-0 rounded bg-surface-2" />
                    )}

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-text-primary truncate">
                          {ed.title || "Untitled"}
                        </span>
                        {ed.format && (
                          <span className="text-[10px] font-medium text-text-primary border border-border rounded px-1.5 py-0.5 leading-none shrink-0">
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
                  </button>
                );
              })}

              {hasMore && (
                <button
                  onClick={loadMore}
                  disabled={loading}
                  className="w-full text-center text-xs text-text-secondary hover:text-text-primary transition-colors py-2 disabled:opacity-50"
                >
                  {loading
                    ? "Loading..."
                    : `Load more (${totalEditions - editions.length} remaining)`}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
