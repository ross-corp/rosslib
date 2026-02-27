"use client";

import { useState } from "react";

export default function BlockButton({
  username,
  initialBlocked,
}: {
  username: string;
  initialBlocked: boolean;
}) {
  const [blocked, setBlocked] = useState(initialBlocked);
  const [loading, setLoading] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  async function handleBlock() {
    setLoading(true);
    const res = await fetch(`/api/users/${username}/block`, {
      method: "POST",
    });
    setLoading(false);
    setShowConfirm(false);
    if (res.ok) {
      setBlocked(true);
      window.location.reload();
    }
  }

  async function handleUnblock() {
    setLoading(true);
    const res = await fetch(`/api/users/${username}/block`, {
      method: "DELETE",
    });
    setLoading(false);
    if (res.ok) {
      setBlocked(false);
      window.location.reload();
    }
  }

  if (showConfirm) {
    return (
      <div className="flex items-center gap-2">
        <span className="text-xs text-text-tertiary">Block @{username}?</span>
        <button
          onClick={handleBlock}
          disabled={loading}
          className="text-xs px-2 py-1 rounded border border-red-500/30 text-red-400 hover:bg-red-500/10 transition-colors disabled:opacity-50"
        >
          {loading ? "..." : "Confirm"}
        </button>
        <button
          onClick={() => setShowConfirm(false)}
          disabled={loading}
          className="text-xs px-2 py-1 rounded border border-border text-text-tertiary hover:text-text-secondary transition-colors disabled:opacity-50"
        >
          Cancel
        </button>
      </div>
    );
  }

  if (blocked) {
    return (
      <button
        onClick={handleUnblock}
        disabled={loading}
        className="text-xs px-2 py-1 rounded border border-red-500/20 text-red-400/70 hover:text-red-400 hover:border-red-500/30 transition-colors disabled:opacity-50"
      >
        {loading ? "..." : "Unblock"}
      </button>
    );
  }

  return (
    <button
      onClick={() => setShowConfirm(true)}
      className="text-xs px-2 py-1 rounded border border-border text-text-tertiary hover:text-text-secondary hover:border-border-strong transition-colors"
    >
      Block
    </button>
  );
}
