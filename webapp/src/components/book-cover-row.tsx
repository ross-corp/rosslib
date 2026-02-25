import Link from "next/link";

type Book = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  rating?: number | null;
  progress_pages?: number | null;
  progress_percent?: number | null;
  page_count?: number | null;
  device_total_pages?: number | null;
};

const sizeClasses = {
  sm: "w-12 h-[72px]",
  md: "w-16 h-24",
  lg: "w-20 h-[120px]",
} as const;

function getProgressPercent(book: Book): number | null {
  if (book.progress_percent != null) return book.progress_percent;
  const total = book.device_total_pages ?? book.page_count;
  if (book.progress_pages != null && total) {
    return Math.min(100, Math.round((book.progress_pages / total) * 100));
  }
  return null;
}

export default function BookCoverRow({
  books,
  size = "md",
  showTitle = false,
  showProgress = false,
  seeAllHref,
  seeAllLabel = "See all",
}: {
  books: Book[];
  size?: "sm" | "md" | "lg";
  showTitle?: boolean;
  showProgress?: boolean;
  seeAllHref?: string;
  seeAllLabel?: string;
}) {
  if (books.length === 0) return null;

  return (
    <div className="flex items-end gap-3 overflow-x-auto pb-2 scrollbar-hide">
      {books.map((book) => {
        const pct = showProgress ? getProgressPercent(book) : null;
        return (
          <Link
            key={book.book_id}
            href={`/books/${book.open_library_id}`}
            className="shrink-0 group"
          >
            {book.cover_url ? (
              <img
                src={book.cover_url}
                alt={book.title}
                className={`${sizeClasses[size]} object-cover rounded shadow-sm group-hover:shadow-md transition-shadow`}
              />
            ) : (
              <div
                className={`${sizeClasses[size]} rounded bg-stone-200 flex items-center justify-center`}
              >
                <span className="text-[10px] text-stone-400 text-center px-1 line-clamp-3">
                  {book.title}
                </span>
              </div>
            )}
            {pct != null && (
              <div className="mt-1">
                <div className="w-full h-1 bg-stone-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-stone-400 rounded-full"
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <p className="text-[10px] text-stone-400 mt-0.5 text-center">
                  {pct}%
                </p>
              </div>
            )}
            {showTitle && (
              <p className="mt-1 text-xs text-stone-600 truncate max-w-[80px] group-hover:text-stone-900">
                {book.title}
              </p>
            )}
          </Link>
        );
      })}
      {seeAllHref && (
        <Link
          href={seeAllHref}
          className="shrink-0 text-xs text-stone-400 hover:text-stone-700 transition-colors self-center pl-1"
        >
          {seeAllLabel} &rarr;
        </Link>
      )}
    </div>
  );
}
