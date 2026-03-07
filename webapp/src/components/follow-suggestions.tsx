"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

type SuggestedUser = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  books_in_common: number;
};

export function FollowSuggestions() {
  const [suggestions, setSuggestions] = useState<SuggestedUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [followedIds, setFollowedIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    fetch("/api/me/suggested-follows?limit=5")
      .then((r) => (r.ok ? r.json() : []))
      .then((data: SuggestedUser[]) => setSuggestions(data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  async function handleFollow(username: string, userId: string) {
    setFollowedIds((prev) => new Set(prev).add(userId));
    try {
      await fetch(`/api/users/${username}/follow`, { method: "POST" });
    } catch {
      setFollowedIds((prev) => {
        const next = new Set(prev);
        next.delete(userId);
        return next;
      });
    }
  }

  if (loading) return null;
  if (suggestions.length === 0) return null;

  return (
    <div className="border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-text-primary mb-3">
        People you might like
      </h3>
      <div className="space-y-3">
        {suggestions.map((user) => {
          const isFollowed = followedIds.has(user.user_id);
          const displayName = user.display_name || user.username;
          return (
            <div key={user.user_id} className="flex items-center gap-3">
              <Link href={`/${user.username}`} className="shrink-0">
                {user.avatar_url ? (
                  <img
                    src={user.avatar_url}
                    alt=""
                    className="w-9 h-9 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-9 h-9 rounded-full bg-surface-2 flex items-center justify-center text-text-tertiary text-xs font-medium">
                    {displayName.charAt(0).toUpperCase()}
                  </div>
                )}
              </Link>
              <div className="flex-1 min-w-0">
                <Link
                  href={`/${user.username}`}
                  className="text-sm font-medium text-text-primary hover:underline truncate block"
                >
                  {displayName}
                </Link>
                <p className="text-xs text-text-tertiary">
                  {user.books_in_common} book{user.books_in_common !== 1 ? "s" : ""} in common
                </p>
              </div>
              <button
                onClick={() => handleFollow(user.username, user.user_id)}
                disabled={isFollowed}
                className={`shrink-0 text-xs px-3 py-1 rounded border transition-colors ${
                  isFollowed
                    ? "border-border text-text-tertiary cursor-default"
                    : "border-border text-text-primary hover:bg-surface-2 cursor-pointer"
                }`}
              >
                {isFollowed ? "Following" : "Follow"}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}
