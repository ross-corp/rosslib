import { notFound } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import SeriesDescription from "@/components/series-description";
import SeriesBookList from "@/components/series-book-list";
import { type StatusValue } from "@/components/shelf-picker";

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

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

const STATUS_COLORS: Record<string, string> = {
  finished: "bg-semantic-success-bg text-semantic-success border-semantic-success-border",
  "currently-reading": "bg-semantic-info-bg text-semantic-info border-semantic-info-border",
  "want-to-read": "bg-semantic-warning-bg text-semantic-warning border-semantic-warning-border",
  dnf: "bg-semantic-error-bg text-semantic-error border-semantic-error-border",
  "owned-to-read": "bg-purple-900/40 text-purple-400 border-purple-800",
};

async function fetchTagKeys(token: string): Promise<TagKey[]> {
  const res = await fetch(`${process.env.API_URL}/me/tag-keys`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchStatusMap(token: string): Promise<Record<string, string>> {
  const res = await fetch(`${process.env.API_URL}/me/books/status-map`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return {};
  return res.json();
}

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
  if (res.status === 404) notFound();
  if (!res.ok) {
    throw new Error("Failed to load series");
  }
  const series: SeriesDetail = await res.json();

  const [tagKeys, statusMap] = await Promise.all([
    currentUser && token ? fetchTagKeys(token) : Promise.resolve(null),
    currentUser && token ? fetchStatusMap(token) : Promise.resolve(null),
  ]);

  const statusKey = tagKeys?.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] | null = statusKey ? statusKey.values : null;
  const statusKeyId: string | null = statusKey?.id ?? null;
  const bookStatusMap: Record<string, string> | null = statusMap;

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

        <SeriesDescription
          seriesId={series.id}
          initialName={series.name}
          initialDescription={series.description ?? ""}
          isLoggedIn={!!currentUser}
        />

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
        <SeriesBookList
          books={series.books}
          statusValues={statusValues}
          statusKeyId={statusKeyId}
          bookStatusMap={bookStatusMap}
        />
      </main>
    </div>
  );
}
