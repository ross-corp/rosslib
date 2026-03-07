import { notFound } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";

// ── Types ──────────────────────────────────────────────────────────────────────

type YearCount = { year: number; count: number };
type MonthCount = { year: number; month: number; count: number };
type RatingBucket = { rating: number; count: number };

type StatsData = {
  books_by_year: YearCount[];
  books_by_month: MonthCount[];
  average_rating: number | null;
  rating_distribution: RatingBucket[];
  total_books: number;
  total_reviews: number;
  total_pages_read: number;
};

// ── Data fetcher ───────────────────────────────────────────────────────────────

async function fetchStats(
  username: string,
  token?: string
): Promise<StatsData | null> {
  const headers: HeadersInit = token
    ? { Authorization: `Bearer ${token}` }
    : {};
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/stats`,
    { cache: "no-store", headers }
  );
  if (res.status === 404 || res.status === 403) return null;
  if (!res.ok) return null;
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

const MONTH_NAMES = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function StatsPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const [, token] = await Promise.all([getUser(), getToken()]);
  const stats = await fetchStats(username, token ?? undefined);

  if (!stats) notFound();

  const maxByYear = Math.max(...stats.books_by_year.map((y) => y.count), 1);
  const maxByMonth = Math.max(...stats.books_by_month.map((m) => m.count), 1);
  const maxRating = Math.max(
    ...stats.rating_distribution.map((r) => r.count),
    1
  );
  const totalRated = stats.rating_distribution.reduce(
    (sum, r) => sum + r.count,
    0
  );

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
          <h1 className="text-2xl font-bold text-text-primary mt-2">
            Reading Stats
          </h1>
        </div>

        {/* Summary cards */}
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-5 mb-10">
          <div>
            <p className="text-2xl font-bold text-text-primary">
              {stats.total_books}
            </p>
            <p className="text-xs text-text-tertiary">books tracked</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-text-primary">
              {stats.total_reviews}
            </p>
            <p className="text-xs text-text-tertiary">reviews</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-text-primary">
              {stats.average_rating != null
                ? stats.average_rating.toFixed(1)
                : "—"}
            </p>
            <p className="text-xs text-text-tertiary">avg rating</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-text-primary">
              {totalRated}
            </p>
            <p className="text-xs text-text-tertiary">books rated</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-text-primary">
              {stats.total_pages_read.toLocaleString()}
            </p>
            <p className="text-xs text-text-tertiary">pages read</p>
          </div>
        </div>

        {/* Rating distribution */}
        <section className="mb-10">
          <h2 className="section-heading mb-4">Rating Distribution</h2>
          {totalRated === 0 ? (
            <p className="text-sm text-text-tertiary">No ratings yet.</p>
          ) : (
            <div className="space-y-2">
              {[...stats.rating_distribution].reverse().map((bucket) => (
                <div key={bucket.rating} className="flex items-center gap-3">
                  <span className="w-12 text-sm text-text-secondary text-right">
                    {"★".repeat(bucket.rating)}
                  </span>
                  <div className="flex-1 h-5 bg-surface-2 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-amber-500/70 rounded-full transition-all"
                      style={{
                        width: `${(bucket.count / maxRating) * 100}%`,
                      }}
                    />
                  </div>
                  <span className="w-8 text-xs text-text-tertiary text-right">
                    {bucket.count}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Books by year */}
        <section className="mb-10">
          <h2 className="section-heading mb-4">Books by Year</h2>
          {stats.books_by_year.length === 0 ? (
            <p className="text-sm text-text-tertiary">
              No finished books with dates yet.
            </p>
          ) : (
            <div className="space-y-2">
              {stats.books_by_year.map((y) => (
                <div key={y.year} className="flex items-center gap-3">
                  <span className="w-12 text-sm text-text-secondary text-right tabular-nums">
                    {y.year}
                  </span>
                  <div className="flex-1 h-5 bg-surface-2 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-accent/70 rounded-full transition-all"
                      style={{
                        width: `${(y.count / maxByYear) * 100}%`,
                      }}
                    />
                  </div>
                  <span className="w-8 text-xs text-text-tertiary text-right tabular-nums">
                    {y.count}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Books by month (current year) */}
        {stats.books_by_month.length > 0 && (
          <section className="mb-10">
            <h2 className="section-heading mb-4">
              {stats.books_by_month[0].year} by Month
            </h2>
            <div className="space-y-2">
              {stats.books_by_month.map((m) => (
                <div key={m.month} className="flex items-center gap-3">
                  <span className="w-12 text-sm text-text-secondary text-right">
                    {MONTH_NAMES[m.month - 1]}
                  </span>
                  <div className="flex-1 h-5 bg-surface-2 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-accent/70 rounded-full transition-all"
                      style={{
                        width: `${(m.count / maxByMonth) * 100}%`,
                      }}
                    />
                  </div>
                  <span className="w-8 text-xs text-text-tertiary text-right tabular-nums">
                    {m.count}
                  </span>
                </div>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
