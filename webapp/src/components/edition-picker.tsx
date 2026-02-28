"use client";

import { useState } from "react";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";

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
  ara: "Arabic",
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
  editions,
  totalEditions,
  currentEditionKey,
}: {
  openLibraryId: string;
  workId: string;
  editions: Edition[];
  totalEditions: number;
  currentEditionKey: string | null;
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
        if (data.entries) {
          setAllEditions((prev) => [...prev, ...parseEditions(data.entries)]);
        }
      }
    } finally {
      setLoading(false);
    }
  }

  async function selectEdition(edition: Edition | null) {
    const edKey = edition?.key ?? "";
    const coverUrl = edition?.cover_url ?? "";
    setSaving(edKey || "__reset");
    try {
      const res = await fetch(`/api/me/books/${openLibraryId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          selected_edition_key: edKey,
          selected_edition_cover_url: coverUrl,
        }),
      });
      if (res.ok) {
        setSelectedKey(edKey || null);
        setOpen(false);
        window.location.reload();
      }
    } finally {
      setSaving(null);
    }
  }

  const hasMore = totalEditions > allEditions.length;

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="text-xs text-text-secondary hover:text-text-primary transition-colors"
      >
        {selectedKey ? "Change edition" : "Select edition"}
      </button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          onClick={() => setOpen(false)}
        >
          <div
            className="bg-surface-0 border border-border rounded-lg shadow-lg w-full max-w-lg max-h-[80vh] flex flex-col mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="text-sm font-semibold text-text-primary">
                Select edition
              </h3>
              <button
                onClick={() => setOpen(false)}
                className="text-text-secondary hover:text-text-primary text-sm"
              >
                Close
              </button>
            </div>

            <div className="overflow-y-auto flex-1 p-4 space-y-2">
              {/* Reset to default option */}
              {selectedKey && (
                <button
                  onClick={() => selectEdition(null)}
                  disabled={saving !== null}
                  className="w-full text-left p-3 rounded-lg border border-border hover:bg-surface-2/50 transition-colors disabled:opacity-50 text-xs text-text-secondary"
                >
                  {saving === "__reset" ? "Resetting..." : "Use default work cover"}
                </button>
              )}

              {allEditions.map((ed) => {
                const isSelected = selectedKey === ed.key;
                return (
                  <button
                    key={ed.key}
                    onClick={() => selectEdition(ed)}
                    disabled={saving !== null}
                    className={`w-full text-left flex gap-3 p-3 rounded-lg border transition-colors disabled:opacity-50 ${
                      isSelected
                        ? "border-text-primary bg-surface-2/80"
                        : "border-border hover:bg-surface-2/50"
                    }`}
                  >
                    {ed.cover_url ? (
                      <img
                        src={ed.cover_url}
                        alt={ed.title}
                        className="w-12 h-[72px] shrink-0 rounded object-cover bg-surface-2"
                      />
                    ) : (
                      <BookCoverPlaceholder title={ed.title || "Untitled"} className="w-12 h-[72px] shrink-0" />
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
                          <span className="text-[10px] font-medium text-text-primary border border-border rounded px-1.5 py-0.5 leading-none shrink-0">
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
                          .join(" \u00b7 ")}
                      </p>
                      {ed.isbn && (
                        <p className="text-[11px] text-text-secondary mt-0.5 font-mono">
                          ISBN {ed.isbn}
                        </p>
                      )}
                      {saving === ed.key && (
                        <p className="text-[11px] text-text-secondary mt-0.5">
                          Saving...
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
                  className="w-full text-center py-2 text-xs text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50"
                >
                  {loading
                    ? "Loading..."
                    : `Load more editions (${totalEditions - allEditions.length} remaining)`}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}

function parseEditions(entries: Record<string, unknown>[]): Edition[] {
  return entries.map((e) => {
    const isbns = (e.isbn_13 as string[]) ?? (e.isbn_10 as string[]) ?? [];
    const coverId = e.covers
      ? ((e.covers as number[])[0] ?? null)
      : null;
    const langs = (e.languages as { key: string }[]) ?? [];
    const langCode = langs.length > 0 ? langs[0].key.replace("/languages/", "") : "";
    const key = ((e.key as string) ?? "").replace("/books/", "");
    return {
      key,
      title: (e.title as string) ?? "",
      publisher: ((e.publishers as string[]) ?? [])[0] ?? null,
      publish_date: (e.publish_date as string) ?? "",
      page_count: (e.number_of_pages as number) ?? null,
      isbn: isbns[0] ?? null,
      cover_url: coverId
        ? `https://covers.openlibrary.org/b/id/${coverId}-M.jpg`
        : null,
      format: (e.physical_format as string) ?? "",
      language: langCode,
    };
  });
}
