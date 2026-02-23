import { notFound } from "next/navigation";
import Nav from "@/components/nav";
import StarRating from "@/components/star-rating";

// ── Types ──────────────────────────────────────────────────────────────────────

type BookDetail = {
  key: string;
  title: string;
  authors: string[] | null;
  description: string | null;
  cover_url: string | null;
  average_rating: number | null;
  rating_count: number;
};

// ── Data fetcher ───────────────────────────────────────────────────────────────

async function fetchBook(workId: string): Promise<BookDetail | null> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function BookPage({
  params,
}: {
  params: Promise<{ workId: string }>;
}) {
  const { workId } = await params;
  const book = await fetchBook(workId);

  if (!book) notFound();

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        <div className="flex gap-8 items-start max-w-3xl">
          {/* Cover */}
          {book.cover_url ? (
            <img
              src={book.cover_url}
              alt={book.title}
              className="w-32 shrink-0 rounded shadow-sm object-cover"
            />
          ) : (
            <div className="w-32 h-48 shrink-0 bg-stone-100 rounded" />
          )}

          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-stone-900 mb-1">
              {book.title}
            </h1>

            {book.authors && book.authors.length > 0 && (
              <p className="text-stone-500 text-sm mb-4">
                {book.authors.join(", ")}
              </p>
            )}

            {book.average_rating != null && (
              <div className="mb-4 text-sm">
                <StarRating
                  rating={book.average_rating}
                  count={book.rating_count}
                />
              </div>
            )}

            {book.description && (
              <p className="text-stone-700 text-sm leading-relaxed">
                {book.description}
              </p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
