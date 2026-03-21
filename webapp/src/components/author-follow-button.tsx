"use client";

import { useState } from "react";
import { useToast } from "@/components/toast";

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
  const toast = useToast();

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
      toast.success(following ? `Unfollowed ${authorName}` : `Following ${authorName}`);
    } else {
      toast.error("Failed to update follow");
    }
  }

  return (
    <button
      onClick={toggle}
      disabled={loading}
      aria-pressed={following}
      className={`text-sm px-3 py-1.5 rounded border transition-colors disabled:opacity-50 ${
        following
          ? "border-border text-text-primary hover:border-border hover:text-text-primary"
          : "border-accent bg-accent text-text-inverted hover:bg-accent-hover"
      }`}
    >
      {loading ? "..." : following ? "Following" : "Follow"}
    </button>
  );
}
