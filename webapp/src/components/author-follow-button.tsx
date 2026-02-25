"use client";

import { useState } from "react";

export default function AuthorFollowButton({
  authorKey,
  authorName,
  initialFollowing,
}: {
  authorKey: string;
  authorName: string;
  initialFollowing: boolean;
}) {
  const [following, setFollowing] = useState(initialFollowing);
  const [loading, setLoading] = useState(false);

  async function toggle() {
    setLoading(true);
    const res = await fetch(`/api/authors/${authorKey}/follow`, {
      method: following ? "DELETE" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ author_name: authorName }),
    });
    setLoading(false);
    if (res.ok) {
      setFollowing(!following);
    }
  }

  return (
    <button
      onClick={toggle}
      disabled={loading}
      className={`text-sm px-3 py-1.5 rounded border transition-colors disabled:opacity-50 ${
        following
          ? "border-border text-text-primary hover:border-border hover:text-text-primary"
          : "border-accent bg-accent text-white hover:bg-surface-3"
      }`}
    >
      {loading ? "..." : following ? "Following" : "Follow"}
    </button>
  );
}
