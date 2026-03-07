import { notFound } from "next/navigation";
import Link from "next/link";
import Pagination from "@/components/pagination";
import ReviewText from "@/components/review-text";
import ReviewSortDropdown from "@/components/review-sort-dropdown";

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
  date_dnf: string | null;
  date_added: string;
  like_count?: number;
};

type ReviewsResponse = {
  reviews: ReviewItem[];
  total: number;
  page: number;
};

// ── Data fetchers ───────────────────────────────────────────────────────────────

async function fetchReviews(
  username: string,
  page: number,
  sort: string
): Promise<ReviewsResponse | null> {
  const sortParam = sort && sort !== "newest" ? `&sort=${sort}` : "";
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/reviews?page=${page}&limit=20${sortParam}`,
    { cache: "no-store" }
  );
  if (res.status === 404) return null;
  if (!res.ok) return { reviews: [], total: 0, page: 1 };
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
  searchParams,
}: {
  params: Promise<{ username: string }>;
  searchParams: Promise<{ page?: string; sort?: string }>;
}) {
  const { username } = await params;
  const sp = await searchParams;
  const page = Math.max(1, parseInt(sp.page || "1", 10) || 1);
  const sort = sp.sort || "newest";
  const data = await fetchReviews(username, page, sort);

  if (data === null) notFound();

  const { reviews, total } = data;
  const totalPages = Math.ceil(total / 20);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <div className="mb-8">
          <Link
            href={`/${username}`}
            className="text-sm text-text-primary hover:text-text-primary transition-colors"
          >
            &larr; {username}
          </Link>
          <div className="flex items-center justify-between mt-2">
            <h1 className="text-2xl font-bold text-text-primary">
              Reviews{total > 0 && <span className="text-text-tertiary font-normal text-lg ml-2">({total})</span>}
            </h1>
            {total > 1 && (
              <ReviewSortDropdown username={username} currentSort={sort} />
            )}
          </div>
        </div>

        {reviews.length === 0 ? (
          <p className="text-text-primary text-sm">No reviews yet.</p>
        ) : (
          <div className="space-y-8">
            {reviews.map((review, i) => (
              <article key={`${review.open_library_id}-${i}`} className="flex gap-4">
                {/* Cover */}
                <Link href={`/books/${review.open_library_id}`} className="shrink-0">
                  {review.cover_url ? (
                    <img
                      src={review.cover_url}
                      alt={review.title}
                      className="w-16 rounded shadow-sm object-cover"
                    />
                  ) : (
                    <div className="w-16 h-24 bg-surface-2 rounded" />
                  )}
                </Link>

                <div className="flex-1 min-w-0">
                  {/* Book info */}
                  <Link
                    href={`/books/${review.open_library_id}`}
                    className="font-semibold text-text-primary hover:underline leading-snug block"
                  >
                    {review.title}
                  </Link>
                  {review.authors && (
                    <p className="text-xs text-text-primary mt-0.5">{review.authors}</p>
                  )}

                  {/* Rating + date */}
                  <div className="flex items-center gap-3 mt-1.5">
                    {review.rating != null && (
                      <span className="text-sm tracking-tight text-amber-500">
                        {renderStars(review.rating)}
                      </span>
                    )}
                    {review.date_read && (
                      <span className="text-xs text-text-primary">
                        Read {formatDate(review.date_read)}
                      </span>
                    )}
                    {review.date_dnf && (
                      <span className="text-xs text-text-primary">
                        Stopped {formatDate(review.date_dnf)}
                      </span>
                    )}
                  </div>

                  {/* Review text */}
                  <div className="mt-2">
                    {review.spoiler ? (
                      <details className="group">
                        <summary className="text-xs text-text-primary cursor-pointer select-none hover:text-text-primary transition-colors">
                          Show review (contains spoilers)
                        </summary>
                        <div className="mt-2 text-sm text-text-primary leading-relaxed">
                          <ReviewText text={review.review_text} />
                        </div>
                      </details>
                    ) : (
                      <div className="text-sm text-text-primary leading-relaxed">
                        <ReviewText text={review.review_text} />
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-3 mt-2">
                    <p className="text-xs text-text-primary">
                      {formatDate(review.date_added)}
                    </p>
                    {(review.like_count ?? 0) > 0 && (
                      <span className="inline-flex items-center gap-1 text-xs text-text-tertiary">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
                        </svg>
                        {review.like_count}
                      </span>
                    )}
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}

        {/* Pagination */}
        <Pagination
          prevHref={page > 1 ? `/${username}/reviews?page=${page - 1}${sort !== "newest" ? `&sort=${sort}` : ""}` : null}
          nextHref={page < totalPages ? `/${username}/reviews?page=${page + 1}${sort !== "newest" ? `&sort=${sort}` : ""}` : null}
          label={totalPages > 1 ? `Page ${page} of ${totalPages}` : undefined}
        />
      </main>
    </div>
  );
}
