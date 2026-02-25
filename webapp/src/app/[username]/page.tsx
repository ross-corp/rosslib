import { notFound } from "next/navigation";
import Link from "next/link";
import FollowButton from "@/components/follow-button";
import { ActivityCard } from "@/components/activity";
import type { ActivityItem } from "@/components/activity";
import { getUser, getToken } from "@/lib/auth";
import BookCoverRow from "@/components/book-cover-row";
import ReadingStats from "@/components/reading-stats";
import RecentReviews from "@/components/recent-reviews";
import type { ReviewItem } from "@/components/recent-reviews";
import ShelfBrowser from "@/components/shelf-browser";

type StatusBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  rating: number | null;
  added_at: string;
  progress_pages?: number | null;
  progress_percent?: number | null;
  page_count?: number | null;
};

type StatusGroup = {
  name: string;
  slug: string;
  count: number;
  books: StatusBook[];
};

type UserBooksResponse = {
  statuses: StatusGroup[];
  unstatused_count: number;
};

type UserProfile = {
  user_id: string;
  username: string;
  display_name: string | null;
  bio: string | null;
  avatar_url: string | null;
  is_private: boolean;
  member_since: string;
  is_following: boolean;
  follow_status: "none" | "active" | "pending";
  followers_count: number;
  following_count: number;
  friends_count: number;
  books_read: number;
  reviews_count: number;
  books_this_year: number;
  average_rating: number | null;
  is_restricted: boolean;
  author_key: string | null;
};

type UserShelf = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  collection_type: string;
  item_count: number;
  books?: StatusBook[];
};

type TagValue = {
  id: string;
  name: string;
  slug: string;
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: TagValue[];
};

async function fetchProfile(
  username: string,
  token?: string
): Promise<UserProfile | null> {
  const headers: HeadersInit = token
    ? { Authorization: `Bearer ${token}` }
    : {};
  const res = await fetch(`${process.env.API_URL}/users/${username}`, {
    cache: "no-store",
    headers,
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchUserBooks(username: string): Promise<UserBooksResponse> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/books?limit=8`,
    { cache: "no-store" }
  );
  if (!res.ok) return { statuses: [], unstatused_count: 0 };
  return res.json();
}

async function fetchUserShelves(username: string): Promise<UserShelf[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves?include_books=8`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

async function fetchTagKeys(username: string): Promise<TagKey[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/tag-keys`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

async function fetchRecentActivity(username: string): Promise<ActivityItem[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/activity?limit=10`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  const data = await res.json();
  return data.activities || [];
}

async function fetchRecentReviews(username: string): Promise<ReviewItem[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/reviews?limit=3`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

export default async function UserPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const [currentUser, token] = await Promise.all([getUser(), getToken()]);
  const profile = await fetchProfile(username, token ?? undefined);

  if (!profile) notFound();

  const isOwnProfile = currentUser?.user_id === profile.user_id;
  const isRestricted = profile.is_restricted && !isOwnProfile;

  const [userBooks, shelves, tagKeys, recentActivity, recentReviews] = isRestricted
    ? [{ statuses: [], unstatused_count: 0 } as UserBooksResponse, [] as UserShelf[], [] as TagKey[], [] as ActivityItem[], [] as ReviewItem[]]
    : await Promise.all([
        fetchUserBooks(username),
        fetchUserShelves(username),
        fetchTagKeys(username),
        fetchRecentActivity(username),
        fetchRecentReviews(username),
      ]);

  const memberSince = new Date(profile.member_since).toLocaleDateString(
    "en-US",
    { month: "long", year: "numeric" }
  );

  const currentlyReadingStatus = userBooks.statuses.find(
    (s) => s.slug === "currently-reading"
  );
  const favorites = shelves.find(
    (s) => s.slug === "favorites" && s.collection_type === "tag"
  );

  // Tag collections (custom shelves, excluding favorites which is shown separately)
  const tagCollections = shelves.filter(
    (s) => s.collection_type === "tag" && s.slug !== "favorites"
  );

  // Label keys with values (status key is excluded by the API)
  const labelKeys = tagKeys.filter((k) => k.values.length > 0);

  return (
    <>
      {/* Header */}
      <div className="mb-10 relative">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-4">
            {profile.avatar_url ? (
              <img
                src={profile.avatar_url}
                alt=""
                className="w-14 h-14 rounded-full object-cover bg-surface-2 shrink-0"
              />
            ) : (
              <div className="w-14 h-14 rounded-full bg-surface-2 flex items-center justify-center shrink-0">
                <span className="text-text-tertiary text-lg font-medium select-none">
                  {(profile.display_name || profile.username)[0].toUpperCase()}
                </span>
              </div>
            )}
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold text-text-primary">
                  {profile.display_name || profile.username}
                </h1>
                {profile.author_key && (
                  <Link
                    href={`/authors/${profile.author_key}`}
                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs font-medium hover:bg-amber-500/20 transition-colors"
                    title="This user is a published author"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                      className="w-3 h-3"
                    >
                      <path d="M10.75 16.82A7.462 7.462 0 0115 15.5c.71 0 1.396.098 2.046.282A.75.75 0 0018 15.06v-11a.75.75 0 00-.546-.721A9.006 9.006 0 0015 3a8.963 8.963 0 00-4.25 1.065V16.82zM9.25 4.065A8.963 8.963 0 005 3c-.85 0-1.673.118-2.454.339A.75.75 0 002 4.06v11a.75.75 0 00.954.721A7.506 7.506 0 015 15.5c1.579 0 3.042.487 4.25 1.32V4.065z" />
                    </svg>
                    Author
                  </Link>
                )}
              </div>
              {profile.display_name && (
                <p className="text-text-tertiary text-sm mt-0.5">
                  @{profile.username}
                </p>
              )}
            </div>
          </div>
          {isOwnProfile ? (
            <div className="absolute top-0 right-0 flex flex-col items-end gap-1">
              <Link
                href="/settings"
                className="btn-secondary"
              >
                Edit profile
              </Link>
              <Link
                href="/settings/follow-requests"
                className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
              >
                Follow requests
              </Link>
              <Link
                href="/settings/tags"
                className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
              >
                Tag categories
              </Link>
              <Link
                href="/scan"
                className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
              >
                Scan ISBN
              </Link>
              <Link
                href="/settings/import"
                className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
              >
                Import from Goodreads
              </Link>
              <Link
                href="/library/compare"
                className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
              >
                Compare lists
              </Link>
            </div>
          ) : currentUser ? (
            <FollowButton
              username={profile.username}
              initialFollowStatus={profile.follow_status || "none"}
            />
          ) : null}
        </div>

        {profile.bio && (
          <p className="text-text-secondary text-sm leading-relaxed mb-4">
            {profile.bio}
          </p>
        )}

        <div className="flex items-center gap-4 flex-wrap">
          {!isRestricted && (
            <>
              <span className="text-sm text-text-secondary">
                <span className="font-semibold text-text-primary">{profile.books_read}</span>{" "}
                <span className="text-text-tertiary">read</span>
              </span>
              <Link
                href={`/${profile.username}/reviews`}
                className="text-sm text-text-secondary hover:text-text-primary transition-colors"
              >
                <span className="font-semibold text-text-primary">
                  {profile.reviews_count}
                </span>{" "}
                <span className="text-text-tertiary">reviews</span>
              </Link>
            </>
          )}
          <span className="text-sm text-text-secondary">
            <span className="font-semibold text-text-primary">{profile.followers_count}</span>{" "}
            <span className="text-text-tertiary">followers</span>
          </span>
          <span className="text-sm text-text-secondary">
            <span className="font-semibold text-text-primary">{profile.following_count}</span>{" "}
            <span className="text-text-tertiary">following</span>
          </span>
          <span className="text-sm text-text-secondary">
            <span className="font-semibold text-text-primary">{profile.friends_count}</span>{" "}
            <span className="text-text-tertiary">friends</span>
          </span>
        </div>
        <p className="text-xs text-text-tertiary mt-1">
          Member since {memberSince}
        </p>

      </div>

      {isRestricted && (
        <div className="text-center py-8 border border-border rounded">
          <p className="text-text-tertiary text-sm">
            This account is private
          </p>
          <p className="text-text-tertiary text-xs mt-1">
            Follow this user to see their books and activity
          </p>
        </div>
      )}

      {!isRestricted && (
        <div className="lg:grid lg:grid-cols-3 lg:gap-8">
          {/* Main content — 2/3 */}
          <div className="lg:col-span-2 space-y-10">
            {/* Currently Reading */}
            {currentlyReadingStatus &&
              currentlyReadingStatus.books.length > 0 && (
                <section>
                  <h2 className="section-heading mb-3">
                    Currently Reading
                  </h2>
                  <BookCoverRow
                    books={currentlyReadingStatus.books}
                    size="lg"
                    showProgress
                    seeAllHref={
                      currentlyReadingStatus.count >
                      currentlyReadingStatus.books.length
                        ? `/${username}/shelves/currently-reading`
                        : undefined
                    }
                  />
                </section>
              )}

            {/* Favorites */}
            {favorites && favorites.books && favorites.books.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">
                  Favorites
                </h2>
                <BookCoverRow
                  books={favorites.books}
                  size="md"
                  seeAllHref={`/${username}/tags/favorites`}
                />
              </section>
            )}

            {/* Reading Stats */}
            <section>
              <h2 className="section-heading mb-3">
                Reading Stats
              </h2>
              <ReadingStats
                booksRead={profile.books_read}
                reviewsCount={profile.reviews_count}
                booksThisYear={profile.books_this_year}
                averageRating={profile.average_rating}
              />
            </section>

            {/* Recent Reviews */}
            {recentReviews.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">
                  Recent Reviews
                </h2>
                <RecentReviews
                  reviews={recentReviews}
                  username={username}
                />
              </section>
            )}

            {/* Status Browser (replaces ShelfBrowser) */}
            {userBooks.statuses.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">
                  Library
                </h2>
                <ShelfBrowser statuses={userBooks.statuses} username={username} />
              </section>
            )}

            {/* Tags */}
            {tagCollections.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">
                  Tags
                </h2>
                <div className="flex flex-wrap gap-2">
                  {tagCollections.map((tag) => (
                    <Link
                      key={tag.id}
                      href={`/${username}/tags/${tag.slug}`}
                      className="tag-pill"
                    >
                      {tag.name || tag.slug}
                      <span className="ml-1.5 text-xs text-text-tertiary">
                        {tag.item_count}
                      </span>
                    </Link>
                  ))}
                </div>
              </section>
            )}

            {/* Labels */}
            {labelKeys.length > 0 && (
              <section>
                <h2 className="section-heading mb-3">
                  Labels
                </h2>
                <div className="space-y-4">
                  {labelKeys.map((key) => (
                    <div key={key.id}>
                      <h3 className="text-xs font-medium text-text-tertiary mb-2">
                        {key.name}
                      </h3>
                      <div className="flex flex-wrap gap-2">
                        {key.values.map((val) => (
                          <Link
                            key={val.id}
                            href={`/${username}/labels/${key.slug}/${val.slug}`}
                            className="tag-pill"
                          >
                            {val.name}
                          </Link>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>

          {/* Sidebar — 1/3 */}
          <div className="mt-10 lg:mt-0">
            <div className="lg:sticky lg:top-20">
              {recentActivity.length > 0 && (
                <div>
                  <h2 className="section-heading mb-2">
                    Recent Activity
                  </h2>
                  <div>
                    {recentActivity.map((item) => (
                      <ActivityCard
                        key={item.id}
                        item={item}
                        showUser={false}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
