import { notFound } from "next/navigation";
import Link from "next/link";
import StarRating from "@/components/star-rating";
import StatusPicker, { type StatusValue } from "@/components/shelf-picker";
import BookReviewEditor from "@/components/book-review-editor";
import ReadingProgress from "@/components/reading-progress";
import ThreadList from "@/components/thread-list";
import ReviewText from "@/components/review-text";
import EditionList from "@/components/edition-list";
import EditionPicker from "@/components/edition-picker";
import BookLinkList from "@/components/book-link-list";
import BookFollowButton from "@/components/book-follow-button";
import GenreRatingEditor from "@/components/genre-rating-editor";
import { getUser, getToken } from "@/lib/auth";

// ── Types ──────────────────────────────────────────────────────────────────────

type BookEdition = {
  key: string;
  title: string;
  publisher: string | null;
  publish_date: string;
  page_count: number | null;
  isbn: string | null;
  cover_url: string | null;
  format: string;
  language: string;
};

type BookSeriesMembership = {
  series_id: string;
  name: string;
  position: number | null;
};

type BookDetail = {
  key: string;
  title: string;
  authors: { name: string; key: string | null }[] | null;
  description: string | null;
  cover_url: string | null;
  average_rating: number | null;
  rating_count: number;
  local_reads_count: number;
  local_want_to_read_count: number;
  publisher: string | null;
  page_count: number | null;
  first_publish_year: number | null;
  edition_count: number;
  editions: BookEdition[] | null;
  subjects: string[];
  series: BookSeriesMembership[] | null;
};

type BookReview = {
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  rating: number | null;
  review_text: string;
  spoiler: boolean;
  date_read: string | null;
  date_dnf: string | null;
  date_added: string;
  is_followed: boolean;
};

type MyBookStatus = {
  status_value_id: string | null;
  status_name: string | null;
  status_slug: string | null;
  rating: number | null;
  review_text: string | null;
  spoiler: boolean;
  date_read: string | null;
  date_dnf: string | null;
  progress_pages: number | null;
  progress_percent: number | null;
  device_total_pages: number | null;
  selected_edition_key: string | null;
  selected_edition_cover_url: string | null;
};

type BookThread = {
  id: string;
  book_id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  title: string;
  body: string;
  spoiler: boolean;
  created_at: string;
  comment_count: number;
};

type BookLinkItem = {
  id: string;
  from_book_ol_id: string;
  to_book_ol_id: string;
  to_book_title: string;
  to_book_authors: string | null;
  to_book_cover_url: string | null;
  link_type: string;
  note: string | null;
  username: string;
  display_name: string | null;
  votes: number;
  user_voted: boolean;
  created_at: string;
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

type AggregateGenreRating = {
  genre: string;
  average: number;
  rater_count: number;
};

type MyGenreRating = {
  genre: string;
  rating: number;
};

// ── Data fetchers ───────────────────────────────────────────────────────────────

async function fetchBook(workId: string): Promise<BookDetail | null> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchBookReviews(workId: string, token?: string): Promise<BookReview[]> {
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  const res = await fetch(`${process.env.API_URL}/books/${workId}/reviews`, {
    headers,
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchThreads(workId: string): Promise<BookThread[]> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}/threads`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
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

async function fetchBookLinks(workId: string, token?: string): Promise<BookLinkItem[]> {
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  const res = await fetch(`${process.env.API_URL}/books/${workId}/links`, {
    headers,
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

async function fetchBookFollowStatus(
  token: string,
  workId: string
): Promise<boolean> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}/follow`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return false;
  const data = await res.json();
  return data.following === true;
}

async function fetchAggregateGenreRatings(
  workId: string
): Promise<AggregateGenreRating[]> {
  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/genre-ratings`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

async function fetchMyGenreRatings(
  token: string,
  workId: string
): Promise<MyGenreRating[]> {
  const res = await fetch(
    `${process.env.API_URL}/me/books/${workId}/genre-ratings`,
    {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    }
  );
  if (!res.ok) return [];
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function renderStars(rating: number): string {
  return Array.from({ length: 5 }, (_, i) => (i < rating ? "\u2605" : "\u2606")).join(
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

  const [book, reviews, threads, bookLinks, tagKeys, myStatus, isFollowingBook, aggregateGenreRatings, myGenreRatings] = await Promise.all([
    fetchBook(workId),
    fetchBookReviews(workId, token ?? undefined),
    fetchThreads(workId),
    fetchBookLinks(workId, token ?? undefined),
    currentUser && token ? fetchTagKeys(token) : Promise.resolve(null),
    currentUser && token
      ? fetchMyBookStatus(token, workId)
      : Promise.resolve(null),
    currentUser && token
      ? fetchBookFollowStatus(token, workId)
      : Promise.resolve(false),
    fetchAggregateGenreRatings(workId),
    currentUser && token
      ? fetchMyGenreRatings(token, workId)
      : Promise.resolve([]),
  ]);

  if (!book) notFound();

  // Extract the Status key and its values for the StatusPicker
  const statusKey = tagKeys?.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] = statusKey?.values ?? [];
  const statusKeyId = statusKey?.id ?? null;

  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* ── Book header ── */}
        <div className="flex gap-8 items-start mb-10">
          {/* Cover */}
          <div className="shrink-0">
            {(myStatus?.selected_edition_cover_url || book.cover_url) ? (
              <img
                src={myStatus?.selected_edition_cover_url ?? book.cover_url!}
                alt={book.title}
                className="w-32 rounded shadow-sm object-cover"
              />
            ) : (
              <div className="w-32 h-48 bg-surface-2 rounded" />
            )}
            {myStatus && book.editions && book.editions.length > 0 && (
              <div className="mt-2 text-center">
                <EditionPicker
                  openLibraryId={workId}
                  workId={workId}
                  editions={book.editions}
                  totalEditions={book.edition_count}
                  currentEditionKey={myStatus.selected_edition_key}
                />
              </div>
            )}
          </div>

          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-text-primary mb-1">
              {book.title}
            </h1>

            {book.series && book.series.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-1">
                {book.series.map((s) => (
                  <Link
                    key={s.series_id}
                    href={`/series/${s.series_id}`}
                    className="text-xs text-text-tertiary hover:text-text-primary transition-colors"
                  >
                    {s.position != null
                      ? `Book ${s.position} in ${s.name}`
                      : s.name}
                  </Link>
                ))}
              </div>
            )}

            {book.authors && book.authors.length > 0 && (
              <p className="text-text-primary text-sm mb-1">
                {book.authors.map((a, i) => (
                  <span key={i}>
                    {i > 0 && ", "}
                    {a.key ? (
                      <Link
                        href={`/authors/${a.key}`}
                        className="hover:underline"
                      >
                        {a.name}
                      </Link>
                    ) : (
                      a.name
                    )}
                  </span>
                ))}
              </p>
            )}

            {(book.first_publish_year || book.publisher || book.page_count) && (
              <p className="text-text-primary text-xs mb-3">
                {[
                  book.first_publish_year && `${book.first_publish_year}`,
                  book.publisher,
                  book.page_count && `${book.page_count} pages`,
                ]
                  .filter(Boolean)
                  .join(" · ")}
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
              <div className="flex items-center gap-3 mb-3 text-xs text-text-primary">
                {book.local_reads_count > 0 && (
                  <span>
                    <span className="font-medium text-text-primary">{book.local_reads_count}</span>{" "}
                    {book.local_reads_count === 1 ? "reader" : "readers"}
                  </span>
                )}
                {book.local_want_to_read_count > 0 && (
                  <span>
                    <span className="font-medium text-text-primary">{book.local_want_to_read_count}</span>{" "}
                    want to read
                  </span>
                )}
              </div>
            )}

            {/* Status picker + follow */}
            {currentUser && (
              <div className="flex items-center gap-3 mb-4">
                {statusValues.length > 0 && statusKeyId && (
                  <StatusPicker
                    openLibraryId={workId}
                    title={book.title}
                    coverUrl={book.cover_url}
                    statusValues={statusValues}
                    statusKeyId={statusKeyId}
                    currentStatusValueId={myStatus?.status_value_id ?? null}
                  />
                )}
                <BookFollowButton
                  workId={workId}
                  initialFollowing={isFollowingBook}
                />
              </div>
            )}

            {/* Reading progress (only when Currently Reading) */}
            {myStatus?.status_slug === "currently-reading" && (
              <div className="mb-4">
                <ReadingProgress
                  openLibraryId={workId}
                  initialPages={myStatus.progress_pages}
                  initialPercent={myStatus.progress_percent}
                  pageCount={book.page_count}
                  initialDeviceTotalPages={myStatus.device_total_pages}
                />
              </div>
            )}

            {book.description && (
              <p className="text-text-primary text-sm leading-relaxed">
                {book.description}
              </p>
            )}

            {book.subjects && book.subjects.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-3">
                {book.subjects.map((subject) => (
                  <Link
                    key={subject}
                    href={`/search?q=${encodeURIComponent(subject)}`}
                    className="inline-block text-xs px-2.5 py-1 rounded-full bg-surface-2 text-text-primary hover:bg-border transition-colors"
                  >
                    {subject}
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* ── User's review (rate / write / edit / delete) ── */}
        {myStatus && (
          <section className="mb-10 border-t border-border pt-8">
            <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-4">
              Your review
            </h2>
            <BookReviewEditor
              openLibraryId={workId}
              initialRating={myStatus.rating}
              initialReviewText={myStatus.review_text}
              initialSpoiler={myStatus.spoiler}
              initialDateRead={myStatus.date_read}
              initialDateDnf={myStatus.date_dnf}
              statusSlug={myStatus.status_slug}
            />
          </section>
        )}

        {/* ── Genre ratings ── */}
        {(aggregateGenreRatings.length > 0 || currentUser) && (
          <section className="mb-10 border-t border-border pt-8">
            <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-4">
              Genre ratings
            </h2>
            <GenreRatingEditor
              openLibraryId={workId}
              isLoggedIn={!!currentUser}
              initialAggregateRatings={aggregateGenreRatings}
              initialMyRatings={myGenreRatings}
            />
          </section>
        )}

        {/* ── Community reviews ── */}
        <section className="border-t border-border pt-8">
          <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-6">
            {reviews.length > 0
              ? `Reviews (${reviews.length})`
              : "Reviews"}
          </h2>

          {reviews.length === 0 ? (
            <p className="text-text-primary text-sm">No reviews yet.</p>
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
                      <div className="w-8 h-8 rounded-full bg-surface-2" />
                    )}
                  </Link>

                  <div className="flex-1 min-w-0">
                    {/* Reviewer */}
                    <div className="flex items-center gap-2">
                      <Link
                        href={`/${review.username}`}
                        className="text-sm font-medium text-text-primary hover:underline"
                      >
                        {review.display_name ?? review.username}
                      </Link>
                      {review.is_followed && (
                        <span className="text-[10px] font-medium text-text-primary border border-border rounded px-1.5 py-0.5 leading-none">
                          Following
                        </span>
                      )}
                    </div>

                    {/* Rating + date */}
                    <div className="flex items-center gap-3 mt-0.5">
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
                        <details>
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

                    <p className="text-xs text-text-primary mt-2">
                      {formatDate(review.date_added)}
                    </p>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>

        {/* ── Editions ── */}
        {book.editions && book.editions.length > 0 && (
          <section className="border-t border-border pt-8 mt-10">
            <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-4">
              Editions{book.edition_count > 0 && ` (${book.edition_count})`}
            </h2>
            <EditionList
              editions={book.editions}
              totalEditions={book.edition_count}
              workId={workId}
            />
          </section>
        )}

        {/* ── Related books (community links) ── */}
        <section className="border-t border-border pt-8 mt-10">
          <BookLinkList
            workId={workId}
            initialLinks={bookLinks}
            isLoggedIn={!!currentUser}
            currentUsername={currentUser?.username}
            isModerator={currentUser?.is_moderator}
          />
        </section>

        {/* ── Discussion threads ── */}
        <section className="border-t border-border pt-8 mt-10">
          <ThreadList
            workId={workId}
            initialThreads={threads}
            isLoggedIn={!!currentUser}
          />
        </section>
      </main>
    </div>
  );
}
