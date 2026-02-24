import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import FollowButton from "@/components/follow-button";
import { getUser, getToken } from "@/lib/auth";

type UserProfile = {
  user_id: string;
  username: string;
  display_name: string | null;
  bio: string | null;
  avatar_url: string | null;
  is_private: boolean;
  member_since: string;
  is_following: boolean;
  followers_count: number;
  following_count: number;
  friends_count: number;
  books_read: number;
  reviews_count: number;
};

type UserShelf = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  collection_type: string;
  item_count: number;
};

async function fetchProfile(
  username: string,
  token?: string
): Promise<UserProfile | null> {
  const headers: HeadersInit = token ? { Authorization: `Bearer ${token}` } : {};
  const res = await fetch(`${process.env.API_URL}/users/${username}`, {
    cache: "no-store",
    headers,
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchUserShelves(username: string): Promise<UserShelf[]> {
  const res = await fetch(`${process.env.API_URL}/users/${username}/shelves`, {
    cache: "no-store",
  });
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
  const [profile, shelves] = await Promise.all([
    fetchProfile(username, token ?? undefined),
    fetchUserShelves(username),
  ]);

  if (!profile) notFound();

  const isOwnProfile = currentUser?.user_id === profile.user_id;
  const memberSince = new Date(profile.member_since).toLocaleDateString(
    "en-US",
    { month: "long", year: "numeric" }
  );

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <div>
          <div className="flex items-start justify-between mb-6">
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
                initialFollowing={profile.is_following}
              />
            ) : null}
          </div>

          {profile.bio && (
            <p className="text-stone-700 text-sm leading-relaxed mb-6">
              {profile.bio}
            </p>
          )}

          <div className="flex items-center gap-4 mt-1">
            <span className="text-sm text-stone-700">
              <span className="font-semibold">{profile.books_read}</span>{" "}
              <span className="text-stone-400">books read</span>
            </span>
            <Link
              href={`/${profile.username}/reviews`}
              className="text-sm text-stone-700 hover:text-stone-900 transition-colors"
            >
              <span className="font-semibold">{profile.reviews_count}</span>{" "}
              <span className="text-stone-400">reviews</span>
            </Link>
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
          <p className="text-xs text-stone-400 mt-1">Member since {memberSince}</p>
        </div>

        {(() => {
          const defaultShelves = shelves.filter(s => s.exclusive_group === "read_status");
          const customShelves = shelves.filter(s => s.exclusive_group !== "read_status" && s.collection_type === "shelf");
          const tags = shelves.filter(s => s.collection_type === "tag");
          return (
            <>
              {defaultShelves.length > 0 && (
                <div className="mt-10">
                  <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-4">
                    Shelves
                  </h2>
                  <div className="flex flex-wrap gap-2">
                    {defaultShelves.map((shelf) => (
                      <Link
                        key={shelf.id}
                        href={`/${profile.username}/shelves/${shelf.slug}`}
                        className="inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full border border-stone-200 text-stone-600 hover:border-stone-400 hover:text-stone-900 transition-colors"
                      >
                        {shelf.name}
                        <span className="text-xs text-stone-400">{shelf.item_count}</span>
                      </Link>
                    ))}
                  </div>
                </div>
              )}
              {customShelves.length > 0 && (
                <div className="mt-8">
                  <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-4">
                    Custom Shelves
                  </h2>
                  <div className="flex flex-wrap gap-2">
                    {customShelves.map((shelf) => (
                      <Link
                        key={shelf.id}
                        href={`/${profile.username}/shelves/${shelf.slug}`}
                        className="inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full border border-stone-200 text-stone-600 hover:border-stone-400 hover:text-stone-900 transition-colors"
                      >
                        {shelf.name}
                        <span className="text-xs text-stone-400">{shelf.item_count}</span>
                      </Link>
                    ))}
                  </div>
                </div>
              )}
              {(tags.length > 0 || isOwnProfile) && (
                <div className="mt-8">
                  <div className="flex items-center justify-between mb-4">
                    <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider">
                      Tags
                    </h2>
                    {isOwnProfile && (
                      <Link
                        href="/settings/tags"
                        className="text-xs text-stone-400 hover:text-stone-700 transition-colors"
                      >
                        Manage labels
                      </Link>
                    )}
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {tags.map((tag) => (
                      <Link
                        key={tag.id}
                        href={`/${profile.username}/tags/${tag.slug}`}
                        className="inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full border border-stone-200 text-stone-600 hover:border-stone-400 hover:text-stone-900 transition-colors"
                      >
                        {tag.name}
                        <span className="text-xs text-stone-400">{tag.item_count}</span>
                      </Link>
                    ))}
                  </div>
                </div>
              )}
            </>
          );
        })()}
      </main>
    </div>
  );
}
