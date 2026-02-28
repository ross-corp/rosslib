import { notFound } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
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

type SeriesDetail = {
  id: string;
  name: string;
  description: string;
  books: SeriesBook[];
};

const STATUS_LABELS: Record<string, string> = {
  "want-to-read": "Want to Read",
  "currently-reading": "Reading",
  finished: "Finished",
  dnf: "Did Not Finish",
  "owned-to-read": "Owned",
};

const STATUS_COLORS: Record<string, string> = {
  finished: "bg-emerald-900/40 text-emerald-400 border-emerald-800",
  "currently-reading": "bg-blue-900/40 text-blue-400 border-blue-800",
  "want-to-read": "bg-amber-900/40 text-amber-400 border-amber-800",
  dnf: "bg-red-900/40 text-red-400 border-red-800",
  "owned-to-read": "bg-purple-900/40 text-purple-400 border-purple-800",
};

export default async function SeriesPage({
  params,
}: {
  params: Promise<{ seriesId: string }>;
}) {
  const { seriesId } = await params;
  const [currentUser, token] = await Promise.all([getUser(), getToken()]);

  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(
    `${process.env.API_URL}/series/${seriesId}`,
    { headers, cache: "no-store" }
  );
  if (!res.ok) notFound();
  const series: SeriesDetail = await res.json();

  const totalBooks = series.books.length;
  const readCount = series.books.filter(
    (b) => b.viewer_status === "finished"
  ).length;

  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        <nav className="flex items-center gap-2 text-xs text-text-tertiary mb-8">
          <Link href="/" className="hover:text-text-primary transition-colors">
            Home
          </Link>
          <span>/</span>
          <span className="text-text-secondary">{series.name}</span>
        </nav>

        <h1 className="text-2xl font-bold text-text-primary mb-2">
          {series.name}
        </h1>

        {series.description && (
          <p className="text-sm text-text-secondary mb-6 leading-relaxed">
            {series.description}
          </p>
        )}

        <div className="flex items-center gap-4 mb-8 text-sm text-text-tertiary">
          <span>
            {totalBooks} {totalBooks === 1 ? "book" : "books"}
          </span>
          {currentUser && readCount > 0 && (
            <span>
              {readCount} of {totalBooks} read
            </span>
          )}
        </div>

        {/* Progress bar for logged-in users */}
        {currentUser && totalBooks > 0 && (
          <div className="mb-8">
            <div className="h-1.5 bg-surface-2 rounded-full overflow-hidden">
              <div
                className="h-full bg-text-tertiary rounded-full transition-all"
                style={{
                  width: `${(readCount / totalBooks) * 100}%`,
                }}
              />
            </div>
          </div>
        )}

        {/* Book list */}
        <div className="space-y-4">
          {series.books.map((book) => (
            <Link
              key={book.book_id}
              href={`/books/${book.open_library_id}`}
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

              {/* Cover */}
              <div className="shrink-0">
                {book.cover_url ? (
                  <img
                    src={book.cover_url}
                    alt={book.title}
                    className="w-12 h-[72px] object-cover rounded shadow-sm bg-surface-2"
                  />
                ) : (
                  <BookCoverPlaceholder title={book.title} author={book.authors} className="w-12 h-[72px]" />
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

              {/* Status badge */}
              {book.viewer_status && (
                <div className="shrink-0 py-1">
                  <span
                    className={`text-[10px] font-medium border rounded px-1.5 py-0.5 leading-none ${STATUS_COLORS[book.viewer_status] ?? "bg-surface-2 text-text-tertiary border-border"}`}
                  >
                    {STATUS_LABELS[book.viewer_status] ?? book.viewer_status}
                  </span>
                </div>
              )}
            </Link>
          ))}
        </div>

        {series.books.length === 0 && (
          <p className="text-sm text-text-tertiary">
            No books in this series yet.
          </p>
        )}
      </main>
    </div>
  );
}
