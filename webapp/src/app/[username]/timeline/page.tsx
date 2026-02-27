import { notFound } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import BookCoverRow from "@/components/book-cover-row";
import TimelineYearSelector from "./year-selector";

// ── Types ───────────────────────────────────────────────────────────────────────

type TimelineBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  rating: number | null;
  date_read: string;
};

type MonthGroup = {
  month: number;
  books: TimelineBook[];
};

type TimelineResponse = {
  year: number;
  months: MonthGroup[];
};

// ── Data fetcher ────────────────────────────────────────────────────────────────

async function fetchTimeline(
  username: string,
  year: number,
  token?: string
): Promise<TimelineResponse | null> {
  const headers: HeadersInit = token
    ? { Authorization: `Bearer ${token}` }
    : {};
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/timeline?year=${year}`,
    { cache: "no-store", headers }
  );
  if (res.status === 404) return null;
  if (res.status === 403) return null;
  if (!res.ok) return { year, months: [] };
  return res.json();
}

// ── Helpers ─────────────────────────────────────────────────────────────────────

const monthNames = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

// ── Page ────────────────────────────────────────────────────────────────────────

export default async function TimelinePage({
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
  const timeline = await fetchTimeline(username, year, token ?? undefined);

  if (timeline === null) notFound();

  const totalBooks = timeline.months.reduce((sum, m) => sum + m.books.length, 0);

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
              Reading Timeline
            </h1>
            <TimelineYearSelector
              username={username}
              currentYear={year}
            />
          </div>
          <p className="text-sm text-text-tertiary mt-1">
            {totalBooks} {totalBooks === 1 ? "book" : "books"} read in {year}
          </p>
        </div>

        {timeline.months.length === 0 ? (
          <p className="text-text-tertiary text-sm">
            No books finished in {year}.
          </p>
        ) : (
          <div className="space-y-8">
            {timeline.months.map((group) => (
              <section key={group.month}>
                <h2 className="section-heading mb-3">
                  {monthNames[group.month - 1]}
                  <span className="ml-2 text-text-tertiary font-normal text-sm">
                    ({group.books.length})
                  </span>
                </h2>
                <BookCoverRow books={group.books} size="lg" showTitle />
              </section>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
