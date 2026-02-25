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
  ara: "Arabic",
  hin: "Hindi",
  pol: "Polish",
  swe: "Swedish",
  nor: "Norwegian",
  dan: "Danish",
  fin: "Finnish",
  tur: "Turkish",
  heb: "Hebrew",
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
  // Capitalize first letter for anything else
  return format.charAt(0).toUpperCase() + format.slice(1);
}

export default function EditionList({
  editions,
  totalEditions,
  workId,
}: {
  editions: Edition[];
  totalEditions: number;
  workId: string;
}) {
  const [allEditions, setAllEditions] = useState<Edition[]>(editions);
  const [loading, setLoading] = useState(false);
  const [expanded, setExpanded] = useState(false);

  if (allEditions.length === 0) return null;

  const displayed = expanded ? allEditions : allEditions.slice(0, 5);
  const hasMore = totalEditions > allEditions.length;

  async function loadMore() {
    setLoading(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/editions?limit=50&offset=${allEditions.length}`
      );
      if (res.ok) {
        const data = await res.json();
        setAllEditions((prev) => [...prev, ...data.editions]);
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <div className="space-y-3">
        {displayed.map((ed) => (
          <div
            key={ed.key}
            className="flex gap-3 p-3 rounded-lg border border-stone-100 bg-stone-50/50"
          >
            {/* Cover thumbnail */}
            {ed.cover_url ? (
              <img
                src={ed.cover_url}
                alt={ed.title}
                className="w-12 h-[72px] shrink-0 rounded object-cover bg-stone-100"
              />
            ) : (
              <div className="w-12 h-[72px] shrink-0 rounded bg-stone-200" />
            )}

            <div className="flex-1 min-w-0">
              {/* Title / format badge */}
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-stone-800 truncate">
                  {ed.title || "Untitled"}
                </span>
                {ed.format && (
                  <span className="text-[10px] font-medium text-stone-500 border border-stone-200 rounded px-1.5 py-0.5 leading-none shrink-0">
                    {formatLabel(ed.format)}
                  </span>
                )}
              </div>

              {/* Metadata line */}
              <p className="text-xs text-stone-500 mt-0.5">
                {[
                  ed.publisher,
                  ed.publish_date,
                  ed.page_count ? `${ed.page_count} pp` : null,
                  ed.language ? langName(ed.language) : null,
                ]
                  .filter(Boolean)
                  .join(" · ")}
              </p>

              {/* ISBN */}
              {ed.isbn && (
                <p className="text-[11px] text-stone-400 mt-0.5 font-mono">
                  ISBN {ed.isbn}
                </p>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Show more / Load more */}
      <div className="mt-3 flex gap-3">
        {!expanded && allEditions.length > 5 && (
          <button
            onClick={() => setExpanded(true)}
            className="text-xs text-stone-500 hover:text-stone-700 transition-colors"
          >
            Show all {allEditions.length} loaded editions
          </button>
        )}
        {expanded && allEditions.length > 5 && (
          <button
            onClick={() => setExpanded(false)}
            className="text-xs text-stone-500 hover:text-stone-700 transition-colors"
          >
            Show fewer
          </button>
        )}
        {hasMore && expanded && (
          <button
            onClick={loadMore}
            disabled={loading}
            className="text-xs text-stone-500 hover:text-stone-700 transition-colors disabled:opacity-50"
          >
            {loading ? "Loading..." : `Load more (${totalEditions - allEditions.length} remaining)`}
          </button>
        )}
      </div>
    </div>
  );
}
