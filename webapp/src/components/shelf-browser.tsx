"use client";

import { useState } from "react";
import Link from "next/link";

type StatusBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  rating: number | null;
  added_at: string;
};

type StatusGroup = {
  name: string;
  slug: string;
  count: number;
  books: StatusBook[];
};

export default function ShelfBrowser({
  statuses,
  username,
}: {
  statuses: StatusGroup[];
  username: string;
}) {
  const [activeSlug, setActiveSlug] = useState(
    () => statuses.find((s) => s.count > 0)?.slug ?? statuses[0]?.slug ?? ""
  );
  const [bookCache, setBookCache] = useState<Record<string, StatusBook[]>>(() => {
    const initial: Record<string, StatusBook[]> = {};
    for (const s of statuses) {
      if (s.books && s.books.length > 0) {
        initial[s.slug] = s.books;
      }
    }
    return initial;
  });
  const [loading, setLoading] = useState(false);

  if (statuses.length === 0) return null;

  const activeStatus = statuses.find((s) => s.slug === activeSlug) ?? statuses[0];
  const books = bookCache[activeStatus.slug] ?? [];

  async function loadStatus(slug: string) {
    setActiveSlug(slug);
    if (bookCache[slug]) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/users/${username}/books?status=${slug}`);
      if (res.ok) {
        const data = await res.json();
        setBookCache((prev) => ({ ...prev, [slug]: data.books ?? [] }));
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <div className="flex gap-1.5 overflow-x-auto pb-2 scrollbar-hide">
        {statuses.map((status) => (
          <button
            key={status.slug}
            onClick={() => loadStatus(status.slug)}
            className={`shrink-0 text-sm px-3 py-1.5 rounded-full border transition-colors ${
              activeStatus.slug === status.slug
                ? "tag-pill-active"
                : "tag-pill"
            }`}
          >
            {status.name}
            <span className="ml-1.5 text-xs opacity-60">{status.count}</span>
          </button>
        ))}
      </div>

      <div className="mt-4 min-h-[120px]">
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="w-5 h-5 border-2 border-border border-t-text-secondary rounded-full animate-spin" />
          </div>
        ) : books.length === 0 ? (
          <p className="text-sm text-text-tertiary py-8 text-center">
            No books with this status yet.
          </p>
        ) : (
          <>
            <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 gap-3">
              {books.map((book) => (
                <Link
                  key={book.book_id}
                  href={`/books/${book.open_library_id}`}
                  className="group"
                >
                  {book.cover_url ? (
                    <img
                      src={book.cover_url}
                      alt={book.title}
                      className="w-full aspect-[2/3] object-cover rounded shadow-sm group-hover:shadow-md transition-shadow"
                    />
                  ) : (
                    <div className="w-full aspect-[2/3] rounded bg-surface-2 flex items-center justify-center">
                      <span className="text-[10px] text-text-tertiary text-center px-1 line-clamp-3">
                        {book.title}
                      </span>
                    </div>
                  )}
                  <p className="mt-1 text-xs text-text-secondary truncate group-hover:text-text-primary">
                    {book.title}
                  </p>
                </Link>
              ))}
            </div>
            {activeStatus.count > books.length && (
              <Link
                href={`/${username}/library/${activeStatus.slug}`}
                className="block text-center text-xs text-text-tertiary hover:text-text-secondary transition-colors pt-3"
              >
                View all {activeStatus.count} books &rarr;
              </Link>
            )}
          </>
        )}
      </div>
    </div>
  );
}
