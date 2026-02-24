import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import StarRating from "@/components/star-rating";
import ShelfPicker, { type Shelf } from "@/components/shelf-picker";
import BookReviewEditor from "@/components/book-review-editor";
import { getUser, getToken } from "@/lib/auth";

// ── Types ──────────────────────────────────────────────────────────────────────

type BookDetail = {
  key: string;
  title: string;
  authors: string[] | null;
  description: string | null;
  cover_url: string | null;
  average_rating: number | null;
  rating_count: number;
  local_reads_count: number;
  local_want_to_read_count: number;
};

type BookReview = {
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  rating: number | null;
  review_text: string;
  spoiler: boolean;
  date_read: string | null;
  date_added: string;
};

type MyBookStatus = {
  shelf_id: string;
  shelf_name: string;
  shelf_slug: string;
  rating: number | null;
  review_text: string | null;
  spoiler: boolean;
  date_read: string | null;
};

type MyShelf = Shelf & {
  exclusive_group: string;
  collection_type: string;
};

// ── Data fetchers ───────────────────────────────────────────────────────────────

async function fetchBook(workId: string): Promise<BookDetail | null> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchBookReviews(workId: string): Promise<BookReview[]> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}/reviews`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchMyShelves(token: string): Promise<MyShelf[]> {
  const res = await fetch(`${process.env.API_URL}/me/shelves`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchMyBookStatus(
  token: string,
  workId: string
): Promise<MyBookStatus | null> {
  const res = await fetch(`${process.env.API_URL}/me/books/${workId}/status`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (res.status === 404) return null;
  if (!res.ok) return null;
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function renderStars(rating: number): string {
  return Array.from({ length: 5 }, (_, i) => (i < rating ? "★" : "☆")).join(
    ""
  );
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });
}

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function BookPage({
  params,
}: {
  params: Promise<{ workId: string }>;
}) {
  const { workId } = await params;
  const [currentUser, token] = await Promise.all([getUser(), getToken()]);

  const [book, reviews, myShelves, myStatus] = await Promise.all([
    fetchBook(workId),
    fetchBookReviews(workId),
    currentUser && token ? fetchMyShelves(token) : Promise.resolve(null),
    currentUser && token
      ? fetchMyBookStatus(token, workId)
      : Promise.resolve(null),
  ]);

  if (!book) notFound();

  const shelves: Shelf[] | null = myShelves
    ? myShelves.map(({ id, name, slug }) => ({ id, name, slug }))
    : null;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* ── Book header ── */}
        <div className="flex gap-8 items-start mb-10">
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
              <p className="text-stone-500 text-sm mb-3">
                {book.authors.join(", ")}
              </p>
            )}

            {book.average_rating != null && (
              <div className="mb-3 text-sm">
                <StarRating
                  rating={book.average_rating}
                  count={book.rating_count}
                />
              </div>
            )}

            {(book.local_reads_count > 0 || book.local_want_to_read_count > 0) && (
              <div className="flex items-center gap-3 mb-3 text-xs text-stone-400">
                {book.local_reads_count > 0 && (
                  <span>
                    <span className="font-medium text-stone-600">{book.local_reads_count}</span>{" "}
                    {book.local_reads_count === 1 ? "reader" : "readers"}
                  </span>
                )}
                {book.local_want_to_read_count > 0 && (
                  <span>
                    <span className="font-medium text-stone-600">{book.local_want_to_read_count}</span>{" "}
                    want to read
                  </span>
                )}
              </div>
            )}

            {/* Shelf picker */}
            {shelves && (
              <div className="mb-4">
                <ShelfPicker
                  openLibraryId={workId}
                  title={book.title}
                  coverUrl={book.cover_url}
                  shelves={shelves}
                  initialShelfId={myStatus?.shelf_id ?? null}
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

        {/* ── User's review (rate / write / edit / delete) ── */}
        {myStatus && (
          <section className="mb-10 border-t border-stone-100 pt-8">
            <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-4">
              Your review
            </h2>
            <BookReviewEditor
              shelfId={myStatus.shelf_id}
              openLibraryId={workId}
              initialRating={myStatus.rating}
              initialReviewText={myStatus.review_text}
              initialSpoiler={myStatus.spoiler}
              initialDateRead={myStatus.date_read}
            />
          </section>
        )}

        {/* ── Community reviews ── */}
        <section className="border-t border-stone-100 pt-8">
          <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-6">
            {reviews.length > 0
              ? `Reviews (${reviews.length})`
              : "Reviews"}
          </h2>

          {reviews.length === 0 ? (
            <p className="text-stone-400 text-sm">No reviews yet.</p>
          ) : (
            <div className="space-y-8">
              {reviews.map((review) => (
                <article key={review.username} className="flex gap-4">
                  {/* Avatar */}
                  <Link
                    href={`/${review.username}`}
                    className="shrink-0"
                  >
                    {review.avatar_url ? (
                      <img
                        src={review.avatar_url}
                        alt={review.display_name ?? review.username}
                        className="w-8 h-8 rounded-full object-cover"
                      />
                    ) : (
                      <div className="w-8 h-8 rounded-full bg-stone-200" />
                    )}
                  </Link>

                  <div className="flex-1 min-w-0">
                    {/* Reviewer */}
                    <Link
                      href={`/${review.username}`}
                      className="text-sm font-medium text-stone-900 hover:underline"
                    >
                      {review.display_name ?? review.username}
                    </Link>

                    {/* Rating + date */}
                    <div className="flex items-center gap-3 mt-0.5">
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
                        <details>
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
        </section>
      </main>
    </div>
  );
}
