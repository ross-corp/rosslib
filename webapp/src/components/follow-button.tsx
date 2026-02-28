"use client";

import { useState } from "react";
import { useToast } from "@/components/toast";

export default function FollowButton({
  username,
  initialFollowStatus,
}: {
  username: string;
  initialFollowStatus: "none" | "active" | "pending";
}) {
  const [status, setStatus] = useState(initialFollowStatus);
  const [loading, setLoading] = useState(false);
  const toast = useToast();

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
        toast.success(`Unfollowed ${username}`);
      } else {
        const data = await res.json();
        const newStatus = data.status === "pending" ? "pending" : "active";
        setStatus(newStatus);
        toast.success(newStatus === "pending" ? `Follow request sent to ${username}` : `Following ${username}`);
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
          ? "border-border text-text-secondary hover:border-border-strong hover:text-text-primary"
          : status === "pending"
            ? "border-border text-text-tertiary hover:border-border-strong hover:text-text-secondary"
            : "border-accent bg-accent text-text-inverted hover:bg-accent-hover"
      }`}
    >
      {loading ? "..." : label}
    </button>
  );
}
