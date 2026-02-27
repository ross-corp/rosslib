"use client";

import { useState } from "react";
import Link from "next/link";

type FollowedBook = {
  open_library_id: string;
  title: string;
  authors: string[] | null;
  cover_url: string | null;
};

export default function FollowedBooksList({
  initialBooks,
}: {
  initialBooks: FollowedBook[];
}) {
  const [books, setBooks] = useState(initialBooks);
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  async function unfollow(workId: string) {
    setLoading((prev) => ({ ...prev, [workId]: true }));
    const res = await fetch(`/api/books/${workId}/follow`, {
      method: "DELETE",
    });
    if (res.ok) {
      setBooks((prev) => prev.filter((b) => b.open_library_id !== workId));
    }
    setLoading((prev) => ({ ...prev, [workId]: false }));
  }

  if (books.length === 0) {
    return (
      <p className="text-sm text-text-primary">
        You aren&apos;t following any books yet. Follow a book from its detail page to get notified about new threads.
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {books.map((book) => (
        <div
          key={book.open_library_id}
          className="flex items-center justify-between py-3 border-b border-border"
        >
          <div className="flex items-center gap-3">
            <Link href={`/books/${book.open_library_id}`} className="shrink-0">
              {book.cover_url ? (
                <img
                  src={book.cover_url}
                  alt={book.title}
                  className="w-10 h-14 object-cover rounded shadow-sm bg-surface-2"
                />
              ) : (
                <div className="w-10 h-14 rounded bg-surface-2 flex items-center justify-center">
                  <span className="text-text-primary text-xs select-none">No cover</span>
                </div>
              )}
            </Link>
            <div className="min-w-0">
              <Link
                href={`/books/${book.open_library_id}`}
                className="text-sm font-medium text-text-primary hover:underline line-clamp-1"
              >
                {book.title}
              </Link>
              {book.authors && book.authors.length > 0 && (
                <p className="text-xs text-text-primary line-clamp-1">
                  {book.authors.join(", ")}
                </p>
              )}
            </div>
          </div>
          <button
            onClick={() => unfollow(book.open_library_id)}
            disabled={loading[book.open_library_id]}
            className="shrink-0 text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
          >
            Unfollow
          </button>
        </div>
      ))}
    </div>
  );
}
