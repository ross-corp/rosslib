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
  publisher: string | null;
  page_count: number | null;
  first_publish_year: number | null;
  edition_count: number;
  editions: BookEdition[] | null;
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

async function fetchEditions(
  workId: string
): Promise<{ entries: BookEdition[]; size: number }> {
  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/editions?limit=20`,
    { cache: "no-store" }
  );
  if (!res.ok) return { entries: [], size: 0 };
  const data = await res.json();
  // OL returns entries under the "entries" key
  const raw = (data.entries ?? []) as Record<string, unknown>[];
  const editions: BookEdition[] = raw.map((e) => {
    const key = ((e.key as string) ?? "").replace("/books/", "");
    let coverUrl: string | null = null;
    if (Array.isArray(e.covers) && (e.covers as number[]).length > 0) {
      coverUrl = `https://covers.openlibrary.org/b/id/${(e.covers as number[])[0]}-M.jpg`;
    }
    let isbn: string | null = null;
    if (Array.isArray(e.isbn_13) && (e.isbn_13 as string[]).length > 0) {
      isbn = (e.isbn_13 as string[])[0];
    } else if (Array.isArray(e.isbn_10) && (e.isbn_10 as string[]).length > 0) {
      isbn = (e.isbn_10 as string[])[0];
    }
    const pubs = e.publishers as string[] | undefined;
    const langs = e.languages as { key: string }[] | undefined;
    return {
      key,
      title: (e.title as string) ?? "",
      publisher: pubs?.[0] ?? null,
      publish_date: (e.publish_date as string) ?? "",
      page_count: (e.number_of_pages as number) ?? null,
      isbn,
      cover_url: coverUrl,
      format: (e.physical_format as string) ?? "",
      language: langs?.[0]?.key?.replace("/languages/", "") ?? "",
    };
  });
  return { entries: editions, size: data.size ?? editions.length };
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

  const [book, reviews, threads, bookLinks, tagKeys, myStatus, isFollowingBook, aggregateGenreRatings, myGenreRatings, editionsData] = await Promise.all([
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
    fetchEditions(workId),
  ]);

  if (!book) notFound();

  // Extract the Status key and its values for the StatusPicker
  const statusKey = tagKeys?.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] = statusKey?.values ?? [];
  const statusKeyId = statusKey?.id ?? null;

  // Prefer the user's selected edition cover
  const editionKey = myStatus?.selected_edition_key ?? null;
  const displayCover = editionKey
    ? `https://covers.openlibrary.org/b/olid/${editionKey}-L.jpg`
    : book.cover_url;

  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* ── Book header ── */}
        <div className="flex gap-8 items-start mb-10">
          {/* Cover */}
          <div className="shrink-0">
            {displayCover ? (
              <img
                src={displayCover}
                alt={book.title}
                className="w-32 rounded shadow-sm object-cover"
              />
            ) : (
              <div className="w-32 h-48 bg-surface-2 rounded" />
            )}
            {myStatus && editionsData.entries.length > 0 && (
              <div className="mt-2">
                <EditionPicker
                  workId={workId}
                  openLibraryId={workId}
                  initialEditions={editionsData.entries}
                  totalEditions={editionsData.size}
                  currentEditionKey={editionKey}
                />
              </div>
            )}
          </div>

          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-text-primary mb-1">
              {book.title}
            </h1>

            {book.authors && book.authors.length > 0 && (
              <p className="text-text-primary text-sm mb-1">
                {book.authors.join(", ")}
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
        {editionsData.entries.length > 0 && (
          <section className="border-t border-border pt-8 mt-10">
            <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-4">
              Editions{editionsData.size > 0 && ` (${editionsData.size})`}
            </h2>
            <EditionList
              editions={editionsData.entries}
              totalEditions={editionsData.size}
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
