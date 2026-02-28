"use client";

import { useState } from "react";

export default function ReviewLikeButton({
  workId,
  reviewUserId,
  initialLiked,
  initialCount,
  disabled,
}: {
  workId: string;
  reviewUserId: string;
  initialLiked: boolean;
  initialCount: number;
  disabled?: boolean;
}) {
  const [liked, setLiked] = useState(initialLiked);
  const [count, setCount] = useState(initialCount);
  const [loading, setLoading] = useState(false);

  async function toggle() {
    if (loading || disabled) return;
    setLoading(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/reviews/${reviewUserId}/like`,
        { method: "POST" }
      );
      if (res.ok) {
        const data = await res.json();
        setLiked(data.liked);
        setCount((prev) => (data.liked ? prev + 1 : Math.max(0, prev - 1)));
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <button
      onClick={toggle}
      disabled={loading || disabled}
      className={`inline-flex items-center gap-1 text-xs transition-colors ${
        disabled
          ? "text-text-tertiary cursor-default"
          : liked
            ? "text-like hover:text-like-hover"
            : "text-text-tertiary hover:text-like-hover"
      }`}
      title={disabled ? "You can't like your own review" : liked ? "Unlike" : "Like"}
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill={liked ? "currentColor" : "none"}
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
      </svg>
      {count > 0 && <span>{count}</span>}
    </button>
  );
}
