"use client";

import Link from "next/link";
import { useState } from "react";
import BookTagPicker, { TagKey } from "@/components/book-tag-picker";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";
import QuickAddButton from "@/components/quick-add-button";
import type { StatusValue } from "@/components/shelf-picker";

type Book = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
  series_position?: number | null;
};

export default function ShelfBookGrid({
  shelfId,
  initialBooks,
  isOwner,
  tagKeys = [],
  statusValues,
  statusKeyId,
}: {
  shelfId: string;
  initialBooks: Book[];
  isOwner: boolean;
  tagKeys?: TagKey[];
  statusValues?: StatusValue[];
  statusKeyId?: string;
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
      <p className="text-sm text-text-primary">No books on this shelf yet.</p>
    );
  }

  return (
    <ul className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-5">
      {books.map((book) => (
        <li key={book.book_id} className="group relative flex flex-col gap-2">
          <Link href={`/books/${book.open_library_id}`} className="block relative">
            {book.cover_url ? (
              <img
                src={book.cover_url}
                alt={book.title}
                className="w-full aspect-[2/3] object-cover rounded shadow-sm bg-surface-2 group-hover:shadow-md transition-shadow"
              />
            ) : (
              <BookCoverPlaceholder
                title={book.title}
                className="w-full aspect-[2/3] group-hover:shadow-md transition-shadow"
              />
            )}
            {book.series_position != null && (
              <span className="absolute top-1 left-1 bg-surface-0/80 backdrop-blur-sm text-[10px] font-mono font-medium text-text-secondary border border-border rounded px-1 py-0.5 leading-none">
                #{book.series_position}
              </span>
            )}
          </Link>
          <div className="min-w-0">
            <Link
              href={`/books/${book.open_library_id}`}
              className="text-xs font-medium text-text-primary hover:text-text-primary line-clamp-2 leading-snug"
            >
              {book.title}
            </Link>
            {book.rating != null && book.rating > 0 && (
              <div className="flex gap-px mt-1" aria-label={`${book.rating} out of 5 stars`}>
                {[1, 2, 3, 4, 5].map((n) => (
                  <span
                    key={n}
                    className={`text-[10px] leading-none ${n <= book.rating! ? "text-amber-500" : "text-text-primary"}`}
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
              className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity bg-surface-0 border border-border rounded px-1.5 py-0.5 text-xs text-text-primary hover:text-text-primary hover:border-border disabled:opacity-50"
            >
              {removing === book.open_library_id ? "..." : "✕"}
            </button>
          )}
          {!isOwner && statusValues && statusKeyId && (
            <QuickAddButton
              openLibraryId={book.open_library_id}
              title={book.title}
              coverUrl={book.cover_url}
              statusValues={statusValues}
              statusKeyId={statusKeyId}
            />
          )}
        </li>
      ))}
    </ul>
  );
}
