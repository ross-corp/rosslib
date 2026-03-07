import { notFound } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import BookCoverRow from "@/components/book-cover-row";
import YearInReviewSelector from "./year-selector";

// ── Types ────────────────────────────────────────────────────────────────────────

type GenreEntry = { name: string; count: number };

type BookSummary = {
  open_library_id: string;
  title: string;
  cover_url: string | null;
  rating?: number;
  page_count?: number;
};

type MonthGroup = {
  month: number;
  count: number;
  books: {
    open_library_id: string;
    title: string;
    cover_url: string | null;
    rating: number | null;
  }[];
};

type YearInReviewData = {
  year: number;
  total_books: number;
  total_pages: number;
  average_rating: number | null;
  highest_rated: BookSummary | null;
  longest_book: BookSummary | null;
  shortest_book: BookSummary | null;
  top_genres: GenreEntry[];
  books_by_month: MonthGroup[];
  available_years: number[];
};

// ── Data fetcher ────────────────────────────────────────────────────────────────

async function fetchYearInReview(
  username: string,
  year: number,
  token?: string
): Promise<YearInReviewData | null> {
  const headers: HeadersInit = token
    ? { Authorization: `Bearer ${token}` }
    : {};
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/year-in-review?year=${year}`,
    { cache: "no-store", headers }
  );
  if (res.status === 404 || res.status === 403) return null;
  if (!res.ok) return null;
  return res.json();
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

const monthNames = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

function StarDisplay({ rating }: { rating: number }) {
  return (
    <span className="text-amber-500">
      {"★".repeat(Math.round(rating))}
      {"☆".repeat(5 - Math.round(rating))}
    </span>
  );
}

// ── Page ─────────────────────────────────────────────────────────────────────────

export default async function YearInReviewPage({
  params,
  searchParams,
}: {
  params: Promise<{ username: string }>;
  searchParams: Promise<{ year?: string }>;
}) {
  const { username } = await params;
  const sp = await searchParams;

  const currentYear = new Date().getFullYear();
  const selectedYear = sp.year ? parseInt(sp.year, 10) : currentYear;
  const year = isNaN(selectedYear) ? currentYear : selectedYear;

  const [, token] = await Promise.all([getUser(), getToken()]);
  const data = await fetchYearInReview(username, year, token ?? undefined);

  if (!data) notFound();

  const maxMonthCount = Math.max(...data.books_by_month.map((m) => m.count), 1);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <div className="mb-8">
          <Link
            href={`/${username}`}
            className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
          >
            &larr; {username}
          </Link>
          <div className="flex items-center justify-between mt-2">
            <h1 className="text-2xl font-bold text-text-primary">
              {year} Year in Review
            </h1>
            <YearInReviewSelector
              username={username}
              currentYear={year}
              availableYears={data.available_years}
            />
          </div>
        </div>

        {data.total_books === 0 ? (
          <p className="text-text-tertiary text-sm">
            No books finished in {year}.
          </p>
        ) : (
          <div className="space-y-10">
            {/* Summary cards */}
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
              <div>
                <p className="text-2xl font-bold text-text-primary">
                  {data.total_books}
                </p>
                <p className="text-xs text-text-tertiary">books finished</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-text-primary">
                  {data.total_pages > 0
                    ? data.total_pages.toLocaleString()
                    : "—"}
                </p>
                <p className="text-xs text-text-tertiary">pages read</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-text-primary">
                  {data.average_rating != null
                    ? data.average_rating.toFixed(1)
                    : "—"}
                </p>
                <p className="text-xs text-text-tertiary">avg rating</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-text-primary">
                  {data.total_pages > 0 && data.total_books > 0
                    ? Math.round(data.total_pages / data.total_books)
                    : "—"}
                </p>
                <p className="text-xs text-text-tertiary">avg pages/book</p>
              </div>
            </div>

            {/* Highest rated book */}
            {data.highest_rated && (
              <section>
                <h2 className="section-heading mb-3">Highest Rated</h2>
                <Link
                  href={`/books/${data.highest_rated.open_library_id}`}
                  className="flex items-center gap-4 p-3 rounded border border-border hover:border-accent/40 transition-colors group"
                >
                  {data.highest_rated.cover_url ? (
                    <img
                      src={data.highest_rated.cover_url}
                      alt={data.highest_rated.title}
                      className="w-12 h-[72px] object-cover rounded shadow-sm"
                    />
                  ) : (
                    <div className="w-12 h-[72px] rounded bg-surface-2 flex items-center justify-center">
                      <span className="text-text-tertiary text-xs">No cover</span>
                    </div>
                  )}
                  <div>
                    <p className="text-sm font-medium text-text-primary group-hover:text-accent transition-colors">
                      {data.highest_rated.title}
                    </p>
                    {data.highest_rated.rating != null && (
                      <p className="text-sm mt-0.5">
                        <StarDisplay rating={data.highest_rated.rating} />
                      </p>
                    )}
                  </div>
                </Link>
              </section>
            )}

            {/* Longest & shortest books */}
            {(data.longest_book || data.shortest_book) && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {data.longest_book && (
                  <section>
                    <h2 className="section-heading mb-3">Longest Book</h2>
                    <Link
                      href={`/books/${data.longest_book.open_library_id}`}
                      className="flex items-center gap-3 p-3 rounded border border-border hover:border-accent/40 transition-colors group"
                    >
                      {data.longest_book.cover_url ? (
                        <img
                          src={data.longest_book.cover_url}
                          alt={data.longest_book.title}
                          className="w-10 h-[60px] object-cover rounded shadow-sm"
                        />
                      ) : (
                        <div className="w-10 h-[60px] rounded bg-surface-2" />
                      )}
                      <div className="min-w-0">
                        <p className="text-sm font-medium text-text-primary group-hover:text-accent transition-colors truncate">
                          {data.longest_book.title}
                        </p>
                        <p className="text-xs text-text-tertiary mt-0.5">
                          {data.longest_book.page_count?.toLocaleString()} pages
                        </p>
                      </div>
                    </Link>
                  </section>
                )}
                {data.shortest_book && (
                  <section>
                    <h2 className="section-heading mb-3">Shortest Book</h2>
                    <Link
                      href={`/books/${data.shortest_book.open_library_id}`}
                      className="flex items-center gap-3 p-3 rounded border border-border hover:border-accent/40 transition-colors group"
                    >
                      {data.shortest_book.cover_url ? (
                        <img
                          src={data.shortest_book.cover_url}
                          alt={data.shortest_book.title}
                          className="w-10 h-[60px] object-cover rounded shadow-sm"
                        />
                      ) : (
                        <div className="w-10 h-[60px] rounded bg-surface-2" />
                      )}
                      <div className="min-w-0">
                        <p className="text-sm font-medium text-text-primary group-hover:text-accent transition-colors truncate">
                          {data.shortest_book.title}
                        </p>
                        <p className="text-xs text-text-tertiary mt-0.5">
                          {data.shortest_book.page_count?.toLocaleString()} pages
                        </p>
                      </div>
                    </Link>
                  </section>
                )}
              </div>
            )}

            {/* Top genres */}
            {data.top_genres.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">Top Genres</h2>
                <div className="flex flex-wrap gap-2">
                  {data.top_genres.map((genre) => (
                    <span
                      key={genre.name}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-surface-2 text-sm text-text-secondary border border-border"
                    >
                      {genre.name}
                      <span className="text-xs text-text-tertiary">
                        ({genre.count})
                      </span>
                    </span>
                  ))}
                </div>
              </section>
            )}

            {/* Books by month chart */}
            <section>
              <h2 className="section-heading mb-4">Books by Month</h2>
              <div className="space-y-2">
                {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => {
                  const group = data.books_by_month.find((g) => g.month === m);
                  const count = group?.count ?? 0;
                  return (
                    <div key={m} className="flex items-center gap-3">
                      <span className="w-12 text-sm text-text-secondary text-right">
                        {monthNames[m - 1].slice(0, 3)}
                      </span>
                      <div className="flex-1 h-5 bg-surface-2 rounded-full overflow-hidden">
                        {count > 0 && (
                          <div
                            className="h-full bg-accent/70 rounded-full transition-all"
                            style={{
                              width: `${(count / maxMonthCount) * 100}%`,
                            }}
                          />
                        )}
                      </div>
                      <span className="w-8 text-xs text-text-tertiary text-right tabular-nums">
                        {count}
                      </span>
                    </div>
                  );
                })}
              </div>
            </section>

            {/* Month-by-month book covers */}
            {data.books_by_month.length > 0 && (
              <section>
                <h2 className="section-heading mb-4">Reading Journey</h2>
                <div className="space-y-8">
                  {data.books_by_month.map((group) => (
                    <div key={group.month}>
                      <h3 className="text-sm font-medium text-text-secondary mb-2">
                        {monthNames[group.month - 1]}
                        <span className="ml-2 text-text-tertiary font-normal">
                          ({group.count})
                        </span>
                      </h3>
                      <BookCoverRow
                        books={group.books.map((b) => ({
                          book_id: b.open_library_id,
                          open_library_id: b.open_library_id,
                          title: b.title,
                          cover_url: b.cover_url,
                          rating: b.rating,
                        }))}
                        size="lg"
                        showTitle
                      />
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
