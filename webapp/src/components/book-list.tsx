"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import StarRating from "@/components/star-rating";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";
import StatusPicker, { type StatusValue } from "@/components/shelf-picker";

function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

type BookResult = {
  key: string;
  title: string;
  authors: string[] | null;
  publish_year: number | null;
  cover_url: string | null;
  average_rating: number | null;
  rating_count: number;
  already_read_count: number;
  subjects: string[] | null;
};

export default function BookList({
  books,
  statusValues,
  statusKeyId,
  bookStatusMap,
}: {
  books: BookResult[];
  statusValues: StatusValue[] | null;
  statusKeyId: string | null;
  bookStatusMap: Record<string, string> | null;
}) {
  const router = useRouter();

  if (books.length === 0) return null;

  return (
    <ul className="divide-y divide-border">
      {books.map((book) => {
        const workId = book.key.replace("/works/", "");
        return (
          <li key={book.key} className="flex items-center gap-3 py-4">
            <Link
              href={`/books/${workId}`}
              className="flex gap-4 flex-1 min-w-0 hover:bg-surface-2 -mx-3 px-3 rounded transition-colors"
            >
              {book.cover_url ? (
                <img
                  src={book.cover_url}
                  alt={book.title}
                  width={48}
                  height={64}
                  className="w-12 h-16 object-cover rounded shrink-0 bg-surface-2"
                />
              ) : (
                <BookCoverPlaceholder
                  title={book.title}
                  author={book.authors?.slice(0, 1).join(", ")}
                  className="w-12 h-16 shrink-0"
                />
              )}
              <div className="flex flex-col justify-center gap-0.5 min-w-0">
                <span className="text-sm font-medium text-text-primary truncate">
                  {book.title}
                </span>
                {book.authors && book.authors.length > 0 && (
                  <span className="text-xs text-text-primary">
                    {book.authors.slice(0, 3).join(", ")}
                  </span>
                )}
                <div className="flex items-center gap-2 mt-0.5">
                  {book.publish_year && (
                    <span className="text-xs text-text-primary">
                      {book.publish_year}
                    </span>
                  )}
                  {book.average_rating != null && (
                    <StarRating
                      rating={book.average_rating}
                      className="text-xs"
                    />
                  )}
                  {book.already_read_count > 0 && (
                    <span className="text-xs text-text-primary">
                      {book.already_read_count.toLocaleString()} reads
                    </span>
                  )}
                </div>
                {book.subjects && book.subjects.length > 0 && (
                  <div className="flex items-center gap-1.5 mt-1">
                    {book.subjects.slice(0, 3).map((subject) => (
                      <span
                        key={subject}
                        role="link"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          router.push(`/genres/${slugify(subject)}`);
                        }}
                        className="text-[10px] px-1.5 py-0.5 rounded bg-surface-2 text-text-secondary hover:bg-surface-3 cursor-pointer transition-colors"
                      >
                        {subject}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </Link>
            {statusValues && statusKeyId && (
              <StatusPicker
                openLibraryId={workId}
                title={book.title}
                coverUrl={book.cover_url}
                statusValues={statusValues}
                statusKeyId={statusKeyId}
                currentStatusValueId={bookStatusMap?.[workId] ?? null}
              />
            )}
          </li>
        );
      })}
    </ul>
  );
}
