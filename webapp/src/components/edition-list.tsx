"use client";

import { useState } from "react";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";
import { type Edition, langName, formatLabel } from "@/lib/constants";

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
            className="flex gap-3 p-3 rounded-lg border border-border bg-surface-2/50"
          >
            {/* Cover thumbnail */}
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
              {/* Title / format badge */}
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-text-primary truncate">
                  {ed.title || "Untitled"}
                </span>
                {ed.format && (
                  <span className="text-[10px] font-medium text-text-primary border border-border rounded px-1.5 py-0.5 leading-none shrink-0">
                    {formatLabel(ed.format)}
                  </span>
                )}
              </div>

              {/* Metadata line */}
              <p className="text-xs text-text-primary mt-0.5">
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
                <p className="text-[11px] text-text-primary mt-0.5 font-mono">
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
            className="text-xs text-text-primary hover:text-text-primary transition-colors"
          >
            Show all {allEditions.length} loaded editions
          </button>
        )}
        {expanded && allEditions.length > 5 && (
          <button
            onClick={() => setExpanded(false)}
            className="text-xs text-text-primary hover:text-text-primary transition-colors"
          >
            Show fewer
          </button>
        )}
        {hasMore && expanded && (
          <button
            onClick={loadMore}
            disabled={loading}
            className="text-xs text-text-primary hover:text-text-primary transition-colors disabled:opacity-50"
          >
            {loading ? "Loading..." : `Load more (${totalEditions - allEditions.length} remaining)`}
          </button>
        )}
      </div>
    </div>
  );
}
