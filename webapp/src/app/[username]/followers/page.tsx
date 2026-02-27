import { notFound } from "next/navigation";
import Link from "next/link";
import FollowButton from "@/components/follow-button";
import { getUser, getToken } from "@/lib/auth";

type FollowUser = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

type UserProfile = {
  user_id: string;
  username: string;
  display_name: string | null;
  is_restricted: boolean;
};

const PER_PAGE = 50;

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

async function fetchFollowers(
  username: string,
  page: number
): Promise<FollowUser[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/followers?page=${page}&limit=${PER_PAGE}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

export default async function FollowersPage({
  params,
  searchParams,
}: {
  params: Promise<{ username: string }>;
  searchParams: Promise<{ page?: string }>;
}) {
  const { username } = await params;
  const { page: pageParam = "1" } = await searchParams;
  const page = Math.max(1, parseInt(pageParam, 10) || 1);

  const [currentUser, token] = await Promise.all([getUser(), getToken()]);
  const profile = await fetchProfile(username, token ?? undefined);
  if (!profile) notFound();

  const isOwnProfile = currentUser?.user_id === profile.user_id;
  if (profile.is_restricted && !isOwnProfile) {
    return (
      <div className="text-center py-12">
        <p className="text-text-tertiary text-sm">This account is private</p>
      </div>
    );
  }

  const followers = await fetchFollowers(username, page);
  const hasNext = followers.length >= PER_PAGE;

  return (
    <>
      <div className="flex items-center gap-2 mb-6">
        <Link
          href={`/${username}`}
          className="text-text-tertiary hover:text-text-primary transition-colors text-sm"
        >
          {profile.display_name || profile.username}
        </Link>
        <span className="text-text-tertiary text-sm">/</span>
        <h1 className="text-xl font-bold text-text-primary">Followers</h1>
      </div>

      {followers.length === 0 ? (
        <p className="text-sm text-text-tertiary">No followers yet.</p>
      ) : (
        <ul className="divide-y divide-border max-w-md">
          {followers.map((user) => (
            <li key={user.user_id} className="flex items-center justify-between py-3">
              <Link
                href={`/${user.username}`}
                className="flex items-center gap-3 min-w-0 hover:opacity-80 transition-opacity"
              >
                {user.avatar_url ? (
                  <img
                    src={user.avatar_url}
                    alt=""
                    className="w-9 h-9 rounded-full object-cover bg-surface-2 shrink-0"
                  />
                ) : (
                  <div className="w-9 h-9 rounded-full bg-surface-2 flex items-center justify-center shrink-0">
                    <span className="text-text-tertiary text-sm font-medium select-none">
                      {(user.display_name || user.username)[0].toUpperCase()}
                    </span>
                  </div>
                )}
                <div className="flex flex-col min-w-0">
                  <span className="text-sm font-medium text-text-primary truncate">
                    {user.display_name || user.username}
                  </span>
                  {user.display_name && (
                    <span className="text-xs text-text-tertiary mt-0.5">
                      @{user.username}
                    </span>
                  )}
                </div>
              </Link>
              {currentUser && currentUser.user_id !== user.user_id && (
                <FollowButton
                  username={user.username}
                  initialFollowStatus="none"
                />
              )}
            </li>
          ))}
        </ul>
      )}

      {(page > 1 || hasNext) && (
        <div className="flex items-center gap-4 mt-8">
          {page > 1 ? (
            <Link
              href={`/${username}/followers?page=${page - 1}`}
              className="text-sm text-text-primary hover:text-text-primary transition-colors"
            >
              &larr; Previous
            </Link>
          ) : (
            <span />
          )}
          {hasNext && (
            <Link
              href={`/${username}/followers?page=${page + 1}`}
              className="text-sm text-text-primary hover:text-text-primary transition-colors ml-auto"
            >
              Next &rarr;
            </Link>
          )}
        </div>
      )}
    </>
  );
}
