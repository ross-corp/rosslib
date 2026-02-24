"use client";

import Link from "next/link";
import { useState } from "react";
import BookTagPicker, { TagKey } from "@/components/book-tag-picker";

type Book = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
};

export default function ShelfBookGrid({
  shelfId,
  initialBooks,
  isOwner,
  tagKeys = [],
}: {
  shelfId: string;
  initialBooks: Book[];
  isOwner: boolean;
  tagKeys?: TagKey[];
}) {
  const [books, setBooks] = useState(initialBooks);
  const [removing, setRemoving] = useState<string | null>(null);

  async function removeBook(olId: string) {
    setRemoving(olId);
    const res = await fetch(`/api/me/books/${olId}`, {
      method: "DELETE",
    });
    setRemoving(null);
    if (res.ok) {
      setBooks((prev) => prev.filter((b) => b.open_library_id !== olId));
    }
  }

  if (books.length === 0) {
    return (
      <p className="text-sm text-stone-400">No books on this shelf yet.</p>
    );
  }

  return (
    <ul className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-5">
      {books.map((book) => (
        <li key={book.book_id} className="group relative flex flex-col gap-2">
          <Link href={`/books/${book.open_library_id}`} className="block">
            {book.cover_url ? (
              <img
                src={book.cover_url}
                alt={book.title}
                className="w-full aspect-[2/3] object-cover rounded shadow-sm bg-stone-100 group-hover:shadow-md transition-shadow"
              />
            ) : (
              <div className="w-full aspect-[2/3] bg-stone-100 rounded shadow-sm flex items-end p-2 group-hover:shadow-md transition-shadow">
                <span className="text-xs text-stone-400 leading-tight line-clamp-3">
                  {book.title}
                </span>
              </div>
            )}
          </Link>
          <div className="min-w-0">
            <Link
              href={`/books/${book.open_library_id}`}
              className="text-xs font-medium text-stone-800 hover:text-stone-900 line-clamp-2 leading-snug"
            >
              {book.title}
            </Link>
            {book.rating != null && book.rating > 0 && (
              <div className="flex gap-px mt-1" aria-label={`${book.rating} out of 5 stars`}>
                {[1, 2, 3, 4, 5].map((n) => (
                  <span
                    key={n}
                    className={`text-[10px] leading-none ${n <= book.rating! ? "text-amber-500" : "text-stone-200"}`}
                  >
                    ★
                  </span>
                ))}
              </div>
            )}
            {isOwner && tagKeys.length > 0 && (
              <div className="mt-1">
                <BookTagPicker
                  openLibraryId={book.open_library_id}
                  tagKeys={tagKeys}
                />
              </div>
            )}
          </div>
          {isOwner && (
            <button
              onClick={() => removeBook(book.open_library_id)}
              disabled={removing === book.open_library_id}
              className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity bg-white border border-stone-200 rounded px-1.5 py-0.5 text-xs text-stone-400 hover:text-stone-700 hover:border-stone-400 disabled:opacity-50"
            >
              {removing === book.open_library_id ? "..." : "✕"}
            </button>
          )}
        </li>
      ))}
    </ul>
  );
}
