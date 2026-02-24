"use client";

import { useState } from "react";
import Link from "next/link";

type ShelfBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
};

type Shelf = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  collection_type: string;
  item_count: number;
  books?: ShelfBook[];
};

export default function ShelfBrowser({
  shelves,
  username,
}: {
  shelves: Shelf[];
  username: string;
}) {
  const [activeSlug, setActiveSlug] = useState(shelves[0]?.slug ?? "");
  const [bookCache, setBookCache] = useState<Record<string, ShelfBook[]>>(() => {
    const initial: Record<string, ShelfBook[]> = {};
    for (const s of shelves) {
      if (s.books && s.books.length > 0) {
        initial[s.slug] = s.books;
      }
    }
    return initial;
  });
  const [loading, setLoading] = useState(false);

  if (shelves.length === 0) return null;

  const activeShelf = shelves.find((s) => s.slug === activeSlug) ?? shelves[0];
  const books = bookCache[activeShelf.slug] ?? [];

  async function loadShelf(slug: string) {
    setActiveSlug(slug);
    if (bookCache[slug]) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/users/${username}/shelves/${slug}`);
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
        {shelves.map((shelf) => (
          <button
            key={shelf.id}
            onClick={() => loadShelf(shelf.slug)}
            className={`shrink-0 text-sm px-3 py-1.5 rounded-full border transition-colors ${
              activeShelf.slug === shelf.slug
                ? "border-stone-900 bg-stone-900 text-white"
                : "border-stone-200 text-stone-600 hover:border-stone-400"
            }`}
          >
            {shelf.name}
            <span className="ml-1.5 text-xs opacity-60">{shelf.item_count}</span>
          </button>
        ))}
      </div>

      <div className="mt-4 min-h-[120px]">
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="w-5 h-5 border-2 border-stone-300 border-t-stone-600 rounded-full animate-spin" />
          </div>
        ) : books.length === 0 ? (
          <p className="text-sm text-stone-400 py-8 text-center">
            No books in this shelf yet.
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
                    <div className="w-full aspect-[2/3] rounded bg-stone-200 flex items-center justify-center">
                      <span className="text-[10px] text-stone-400 text-center px-1 line-clamp-3">
                        {book.title}
                      </span>
                    </div>
                  )}
                  <p className="mt-1 text-xs text-stone-600 truncate group-hover:text-stone-900">
                    {book.title}
                  </p>
                </Link>
              ))}
            </div>
            {activeShelf.item_count > books.length && (
              <Link
                href={`/${username}/shelves/${activeShelf.slug}`}
                className="block text-center text-xs text-stone-400 hover:text-stone-700 transition-colors pt-3"
              >
                View all {activeShelf.item_count} books &rarr;
              </Link>
            )}
          </>
        )}
      </div>
    </div>
  );
}
