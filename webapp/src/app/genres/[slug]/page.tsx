import { notFound } from "next/navigation";
import Link from "next/link";
import BookList from "@/components/book-list";
import { type StatusValue } from "@/components/shelf-picker";
import { getToken, getUser } from "@/lib/auth";

// ── Types ──────────────────────────────────────────────────────────────────────

type BookResult = {
  key: string;
  title: string;
  authors: string[] | null;
  publish_year: number | null;
  isbn: string[] | null;
  cover_url: string | null;
  edition_count: number;
  average_rating: number | null;
  rating_count: number;
  already_read_count: number;
  subjects: string[] | null;
};

type GenreBooksResponse = {
  genre: string;
  total: number;
  page: number;
  results: BookResult[];
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

// ── Data fetchers ──────────────────────────────────────────────────────────────

async function fetchGenreBooks(
  slug: string,
  page: number,
): Promise<GenreBooksResponse | null> {
  const res = await fetch(
    `${process.env.API_URL}/genres/${slug}/books?page=${page}&limit=20`,
    { cache: "no-store" },
  );
  if (res.status === 404) return null;
  if (!res.ok) return null;
  return res.json();
}

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

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function GenreDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ slug: string }>;
  searchParams: Promise<{ page?: string }>;
}) {
  const { slug } = await params;
  const { page: pageParam = "1" } = await searchParams;
  const page = Math.max(1, parseInt(pageParam, 10) || 1);

  const [currentUser, token] = await Promise.all([getUser(), getToken()]);

  const [data, tagKeys, statusMap] = await Promise.all([
    fetchGenreBooks(slug, page),
    currentUser && token ? fetchTagKeys(token) : Promise.resolve(null),
    currentUser && token ? fetchStatusMap(token) : Promise.resolve(null),
  ]);

  if (!data) notFound();

  const statusKey = tagKeys?.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] | null = statusKey ? statusKey.values : null;
  const statusKeyId: string | null = statusKey?.id ?? null;
  const bookStatusMap: Record<string, string> | null = statusMap;

  const totalPages = Math.ceil(data.total / 20);
  const hasNext = page < totalPages;

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-2">
          <Link
            href="/genres"
            className="text-xs text-text-primary hover:text-text-primary transition-colors"
          >
            Genres
          </Link>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-1">
          {data.genre}
        </h1>
        <p className="text-sm text-text-primary mb-8">
          {data.total.toLocaleString()} book{data.total === 1 ? "" : "s"}
        </p>

        {data.results.length === 0 ? (
          <p className="text-sm text-text-primary">
            No books found in this genre yet.
          </p>
        ) : (
          <BookList
            books={data.results}
            statusValues={statusValues}
            statusKeyId={statusKeyId}
            bookStatusMap={bookStatusMap}
          />
        )}

        {(page > 1 || hasNext) && (
          <div className="flex items-center gap-4 mt-8">
            {page > 1 ? (
              <Link
                href={`/genres/${slug}?page=${page - 1}`}
                className="text-sm text-text-primary hover:text-text-primary transition-colors"
              >
                &larr; Previous
              </Link>
            ) : (
              <span />
            )}
            <span className="text-xs text-text-primary">
              Page {page} of {totalPages}
            </span>
            {hasNext && (
              <Link
                href={`/genres/${slug}?page=${page + 1}`}
                className="text-sm text-text-primary hover:text-text-primary transition-colors ml-auto"
              >
                Next &rarr;
              </Link>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
