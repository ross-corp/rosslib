"use client";

import { useState } from "react";

export default function FollowButton({
  username,
  initialFollowStatus,
}: {
  username: string;
  initialFollowStatus: "none" | "active" | "pending";
}) {
  const [status, setStatus] = useState(initialFollowStatus);
  const [loading, setLoading] = useState(false);

  async function toggle() {
    setLoading(true);
    const isUnfollow = status === "active" || status === "pending";
    const res = await fetch(`/api/users/${username}/follow`, {
      method: isUnfollow ? "DELETE" : "POST",
    });
    setLoading(false);
    if (res.ok) {
      if (isUnfollow) {
        setStatus("none");
      } else {
        const data = await res.json();
        setStatus(data.status === "pending" ? "pending" : "active");
      }
    }
  }

  const label =
    status === "active"
      ? "Following"
      : status === "pending"
        ? "Requested"
        : "Follow";

  return (
    <button
      onClick={toggle}
      disabled={loading}
      className={`text-sm px-3 py-1.5 rounded border transition-colors disabled:opacity-50 ${
        status === "active"
          ? "border-stone-300 text-stone-600 hover:border-stone-400 hover:text-stone-900"
          : status === "pending"
            ? "border-stone-200 text-stone-400 hover:border-stone-300 hover:text-stone-600"
            : "border-stone-900 bg-stone-900 text-white hover:bg-stone-700"
      }`}
    >
      {loading ? "..." : label}
    </button>
  );
}
