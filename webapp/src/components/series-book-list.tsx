"use client";

import Link from "next/link";
import StatusPicker, { type StatusValue } from "@/components/shelf-picker";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";

type SeriesBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  authors: string | null;
  position: number | null;
  viewer_status?: string;
};

export default function SeriesBookList({
  books,
  statusValues,
  statusKeyId,
  bookStatusMap,
}: {
  books: SeriesBook[];
  statusValues: StatusValue[] | null;
  statusKeyId: string | null;
  bookStatusMap: Record<string, string> | null;
}) {
  if (books.length === 0) {
    return (
      <p className="text-sm text-text-tertiary">
        No books in this series yet.
      </p>
    );
  }

  return (
    <div className="space-y-4">
      {books.map((book) => (
        <div
          key={book.book_id}
          className="flex gap-4 items-start p-3 -mx-3 rounded-lg hover:bg-surface-2 transition-colors group"
        >
          {/* Position */}
          <div className="shrink-0 w-8 text-center">
            {book.position != null ? (
              <span className="text-sm font-mono text-text-tertiary">
                #{book.position}
              </span>
            ) : (
              <span className="text-sm font-mono text-text-tertiary">
                —
              </span>
            )}
          </div>

          {/* Cover + Info (linked) */}
          <Link
            href={`/books/${book.open_library_id}`}
            className="flex gap-4 flex-1 min-w-0"
          >
            {/* Cover */}
            <div className="shrink-0">
              {book.cover_url ? (
                <img
                  src={book.cover_url}
                  alt={book.title}
                  className="w-12 h-[72px] object-cover rounded shadow-sm bg-surface-2"
                />
              ) : (
                <BookCoverPlaceholder
                  title={book.title}
                  author={book.authors ?? undefined}
                  className="w-12 h-[72px]"
                />
              )}
            </div>

            {/* Info */}
            <div className="flex-1 min-w-0 py-1">
              <p className="text-sm font-medium text-text-primary group-hover:text-text-primary line-clamp-1">
                {book.title}
              </p>
              {book.authors && (
                <p className="text-xs text-text-tertiary mt-0.5">
                  {book.authors}
                </p>
              )}
            </div>
          </Link>

          {/* Status picker */}
          <div className="shrink-0 py-1">
            {statusValues && statusKeyId ? (
              <StatusPicker
                openLibraryId={book.open_library_id}
                title={book.title}
                coverUrl={book.cover_url}
                statusValues={statusValues}
                statusKeyId={statusKeyId}
                currentStatusValueId={bookStatusMap?.[book.open_library_id] ?? null}
              />
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}
