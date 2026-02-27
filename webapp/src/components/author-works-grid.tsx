"use client";

import { useState } from "react";
import Link from "next/link";

type AuthorWork = {
  key: string;
  title: string;
  cover_url: string | null;
};

const PAGE_SIZE = 24;

export default function AuthorWorksGrid({
  authorKey,
  initialWorks,
  workCount,
}: {
  authorKey: string;
  initialWorks: AuthorWork[];
  workCount: number;
}) {
  const [works, setWorks] = useState<AuthorWork[]>(initialWorks);
  const [loading, setLoading] = useState(false);

  const hasMore = works.length < workCount;

  async function loadMore() {
    setLoading(true);
    try {
      const res = await fetch(
        `/api/authors/${authorKey}/works?limit=${PAGE_SIZE}&offset=${works.length}`
      );
      if (res.ok) {
        const data = await res.json();
        setWorks((prev) => [...prev, ...(data.works ?? [])]);
      }
    } finally {
      setLoading(false);
    }
  }

  if (works.length === 0) {
    return <p className="text-text-primary text-sm">No works found.</p>;
  }

  return (
    <>
      <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 gap-4">
        {works.map((work) => (
          <Link key={work.key} href={`/books/${work.key}`} className="group">
            {work.cover_url ? (
              <img
                src={work.cover_url}
                alt={work.title}
                className="w-full aspect-[2/3] object-cover rounded shadow-sm bg-surface-2 group-hover:shadow-md transition-shadow"
              />
            ) : (
              <div className="w-full aspect-[2/3] bg-surface-2 rounded flex items-center justify-center p-2">
                <span className="text-[10px] text-text-primary text-center leading-tight line-clamp-3">
                  {work.title}
                </span>
              </div>
            )}
            <p className="mt-1.5 text-xs text-text-primary leading-tight line-clamp-2 group-hover:text-text-primary transition-colors">
              {work.title}
            </p>
          </Link>
        ))}
      </div>

      {hasMore && (
        <div className="flex justify-center mt-8">
          <button
            onClick={loadMore}
            disabled={loading}
            className="px-6 py-2 text-sm font-medium text-text-primary border border-border rounded-md hover:bg-surface-2 transition-colors disabled:opacity-50"
          >
            {loading ? "Loading…" : "Show more"}
          </button>
        </div>
      )}
    </>
  );
}
