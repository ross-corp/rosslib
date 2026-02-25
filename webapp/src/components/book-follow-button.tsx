"use client";

import { useState } from "react";

export default function BookFollowButton({
  workId,
  initialFollowing,
}: {
  workId: string;
  initialFollowing: boolean;
}) {
  const [following, setFollowing] = useState(initialFollowing);
  const [loading, setLoading] = useState(false);

  async function toggle() {
    setLoading(true);
    const res = await fetch(`/api/books/${workId}/follow`, {
      method: following ? "DELETE" : "POST",
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
          ? "border-stone-300 text-stone-600 hover:border-stone-400 hover:text-stone-900"
          : "border-stone-300 text-stone-500 hover:border-stone-400 hover:text-stone-700"
      }`}
    >
      {loading ? "..." : following ? "Following book" : "Follow book"}
    </button>
  );
}
