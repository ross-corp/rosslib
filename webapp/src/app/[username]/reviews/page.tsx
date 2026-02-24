import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";

// ── Types ──────────────────────────────────────────────────────────────────────

type ReviewItem = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  authors: string | null;
  rating: number | null;
  review_text: string;
  spoiler: boolean;
  date_read: string | null;
  date_added: string;
};

// ── Data fetchers ───────────────────────────────────────────────────────────────

async function fetchReviews(username: string): Promise<ReviewItem[] | null> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/reviews`,
    { cache: "no-store" }
  );
  if (res.status === 404) return null;
  if (!res.ok) return [];
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function renderStars(rating: number): string {
  return Array.from({ length: 5 }, (_, i) => (i < rating ? "★" : "☆")).join("");
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });
}

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function ReviewsPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const reviews = await fetchReviews(username);

  if (reviews === null) notFound();

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <div className="mb-8">
          <Link
            href={`/${username}`}
            className="text-sm text-stone-400 hover:text-stone-700 transition-colors"
          >
            &larr; {username}
          </Link>
          <h1 className="text-2xl font-bold text-stone-900 mt-2">Reviews</h1>
        </div>

        {reviews.length === 0 ? (
          <p className="text-stone-400 text-sm">No reviews yet.</p>
        ) : (
          <div className="space-y-8">
            {reviews.map((review) => (
              <article key={review.book_id} className="flex gap-4">
                {/* Cover */}
                <Link href={`/books/${review.open_library_id}`} className="shrink-0">
                  {review.cover_url ? (
                    <img
                      src={review.cover_url}
                      alt={review.title}
                      className="w-16 rounded shadow-sm object-cover"
                    />
                  ) : (
                    <div className="w-16 h-24 bg-stone-100 rounded" />
                  )}
                </Link>

                <div className="flex-1 min-w-0">
                  {/* Book info */}
                  <Link
                    href={`/books/${review.open_library_id}`}
                    className="font-semibold text-stone-900 hover:underline leading-snug block"
                  >
                    {review.title}
                  </Link>
                  {review.authors && (
                    <p className="text-xs text-stone-400 mt-0.5">{review.authors}</p>
                  )}

                  {/* Rating + date */}
                  <div className="flex items-center gap-3 mt-1.5">
                    {review.rating != null && (
                      <span className="text-sm tracking-tight text-amber-500">
                        {renderStars(review.rating)}
                      </span>
                    )}
                    {review.date_read && (
                      <span className="text-xs text-stone-400">
                        Read {formatDate(review.date_read)}
                      </span>
                    )}
                  </div>

                  {/* Review text */}
                  <div className="mt-2">
                    {review.spoiler ? (
                      <details className="group">
                        <summary className="text-xs text-stone-400 cursor-pointer select-none hover:text-stone-600 transition-colors">
                          Show review (contains spoilers)
                        </summary>
                        <p className="mt-2 text-sm text-stone-700 leading-relaxed whitespace-pre-wrap">
                          {review.review_text}
                        </p>
                      </details>
                    ) : (
                      <p className="text-sm text-stone-700 leading-relaxed whitespace-pre-wrap">
                        {review.review_text}
                      </p>
                    )}
                  </div>

                  <p className="text-xs text-stone-400 mt-2">
                    {formatDate(review.date_added)}
                  </p>
                </div>
              </article>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
