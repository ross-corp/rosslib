import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import FollowButton from "@/components/follow-button";
import { ActivityCard } from "@/components/activity";
import type { ActivityItem } from "@/components/activity";
import { getUser, getToken } from "@/lib/auth";
import BookCoverRow from "@/components/book-cover-row";
import ReadingStats from "@/components/reading-stats";
import RecentReviews from "@/components/recent-reviews";
import type { ReviewItem } from "@/components/recent-reviews";
import ShelfBrowser from "@/components/shelf-browser";

type ShelfBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
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
};

type UserShelf = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  collection_type: string;
  item_count: number;
  books?: ShelfBook[];
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

async function fetchUserShelves(username: string): Promise<UserShelf[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves?include_books=8`,
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

  const [shelves, recentActivity, recentReviews] = isRestricted
    ? [[] as UserShelf[], [] as ActivityItem[], [] as ReviewItem[]]
    : await Promise.all([
        fetchUserShelves(username),
        fetchRecentActivity(username),
        fetchRecentReviews(username),
      ]);

  const memberSince = new Date(profile.member_since).toLocaleDateString(
    "en-US",
    { month: "long", year: "numeric" }
  );

  const currentlyReading = shelves.find(
    (s) => s.slug === "currently-reading"
  );
  const favorites = shelves.find(
    (s) => s.slug === "favorites" && s.collection_type === "tag"
  );
  const allShelves = shelves.filter((s) => s.collection_type === "shelf");

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-6xl mx-auto px-4 sm:px-6 py-12">
        {/* Header */}
        <div className="mb-10">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h1 className="text-2xl font-bold text-stone-900">
                {profile.display_name || profile.username}
              </h1>
              {profile.display_name && (
                <p className="text-stone-400 text-sm mt-0.5">
                  @{profile.username}
                </p>
              )}
            </div>
            {isOwnProfile ? (
              <Link
                href="/settings"
                className="text-sm text-stone-500 hover:text-stone-900 border border-stone-300 px-3 py-1.5 rounded transition-colors"
              >
                Edit profile
              </Link>
            ) : currentUser ? (
              <FollowButton
                username={profile.username}
                initialFollowStatus={profile.follow_status || "none"}
              />
            ) : null}
          </div>

          {profile.bio && (
            <p className="text-stone-700 text-sm leading-relaxed mb-4">
              {profile.bio}
            </p>
          )}

          <div className="flex items-center gap-4 flex-wrap">
            {!isRestricted && (
              <>
                <span className="text-sm text-stone-700">
                  <span className="font-semibold">{profile.books_read}</span>{" "}
                  <span className="text-stone-400">read</span>
                </span>
                <Link
                  href={`/${profile.username}/reviews`}
                  className="text-sm text-stone-700 hover:text-stone-900 transition-colors"
                >
                  <span className="font-semibold">
                    {profile.reviews_count}
                  </span>{" "}
                  <span className="text-stone-400">reviews</span>
                </Link>
              </>
            )}
            <span className="text-sm text-stone-700">
              <span className="font-semibold">{profile.followers_count}</span>{" "}
              <span className="text-stone-400">followers</span>
            </span>
            <span className="text-sm text-stone-700">
              <span className="font-semibold">{profile.following_count}</span>{" "}
              <span className="text-stone-400">following</span>
            </span>
            <span className="text-sm text-stone-700">
              <span className="font-semibold">{profile.friends_count}</span>{" "}
              <span className="text-stone-400">friends</span>
            </span>
          </div>
          <p className="text-xs text-stone-400 mt-1">
            Member since {memberSince}
          </p>
        </div>

        {isRestricted && (
          <div className="text-center py-8 border border-stone-200 rounded-lg">
            <p className="text-stone-400 text-sm">
              This account is private
            </p>
            <p className="text-stone-400 text-xs mt-1">
              Follow this user to see their books and activity
            </p>
          </div>
        )}

        {!isRestricted && (
          <div className="lg:grid lg:grid-cols-3 lg:gap-8">
            {/* Main content — 2/3 */}
            <div className="lg:col-span-2 space-y-10">
              {/* Currently Reading */}
              {currentlyReading &&
                currentlyReading.books &&
                currentlyReading.books.length > 0 && (
                  <section>
                    <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-3">
                      Currently Reading
                    </h2>
                    <BookCoverRow
                      books={currentlyReading.books}
                      size="lg"
                      seeAllHref={
                        currentlyReading.item_count >
                        currentlyReading.books.length
                          ? `/${username}/shelves/currently-reading`
                          : undefined
                      }
                    />
                  </section>
                )}

              {/* Favorites */}
              {favorites && favorites.books && favorites.books.length > 0 && (
                <section>
                  <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-3">
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
                <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-3">
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
                  <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-3">
                    Recent Reviews
                  </h2>
                  <RecentReviews
                    reviews={recentReviews}
                    username={username}
                  />
                </section>
              )}

              {/* Shelf Browser */}
              {allShelves.length > 0 && (
                <section>
                  <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-3">
                    Shelves
                  </h2>
                  <ShelfBrowser shelves={allShelves} username={username} />
                </section>
              )}
            </div>

            {/* Sidebar — 1/3 */}
            <div className="mt-10 lg:mt-0">
              <div className="lg:sticky lg:top-20">
                {recentActivity.length > 0 && (
                  <div>
                    <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-2">
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
      </main>
    </div>
  );
}
