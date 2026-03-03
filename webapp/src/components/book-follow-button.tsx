"use client";

import { useState } from "react";

export default function BookFollowButton({
  workId,
  initialFollowing,
  initialFollowerCount,
}: {
  workId: string;
  initialFollowing: boolean;
  initialFollowerCount: number;
}) {
  const [following, setFollowing] = useState(initialFollowing);
  const [followerCount, setFollowerCount] = useState(initialFollowerCount);
  const [loading, setLoading] = useState(false);

  async function toggle() {
    setLoading(true);
    const willFollow = !following;
    // Optimistic update
    setFollowing(willFollow);
    setFollowerCount((c) => c + (willFollow ? 1 : -1));

    const res = await fetch(`/api/books/${workId}/follow`, {
      method: willFollow ? "POST" : "DELETE",
    });
    setLoading(false);
    if (!res.ok) {
      // Revert on failure
      setFollowing(!willFollow);
      setFollowerCount((c) => c + (willFollow ? -1 : 1));
    }
  }

  return (
    <button
      onClick={toggle}
      disabled={loading}
      className={`text-sm px-3 py-1.5 rounded border transition-colors disabled:opacity-50 ${
        following
          ? "border-border text-text-primary hover:border-border hover:text-text-primary"
          : "border-border text-text-primary hover:border-border hover:text-text-primary"
      }`}
    >
      {loading
        ? "..."
        : following
          ? `Following${followerCount > 0 ? ` · ${followerCount}` : ""}`
          : `Follow${followerCount > 0 ? ` · ${followerCount}` : ""}`}
    </button>
  );
}
