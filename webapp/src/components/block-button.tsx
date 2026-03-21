"use client";

import { useState } from "react";
import { useToast } from "@/components/toast";
import ConfirmDialog from "@/components/confirm-dialog";

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
  const toast = useToast();

  async function handleBlock() {
    setLoading(true);
    const res = await fetch(`/api/users/${username}/block`, {
      method: "POST",
    });
    setLoading(false);
    setShowConfirm(false);
    if (res.ok) {
      setBlocked(true);
      toast.success(`Blocked ${username}`);
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
      toast.success(`Unblocked ${username}`);
      window.location.reload();
    }
  }

  if (blocked) {
    return (
      <button
        onClick={handleUnblock}
        disabled={loading}
        className="text-xs px-2 py-1 rounded border border-danger-bg text-danger hover:text-danger hover:border-semantic-error-border transition-colors disabled:opacity-50"
      >
        {loading ? "..." : "Unblock"}
      </button>
    );
  }

  return (
    <>
      <button
        onClick={() => setShowConfirm(true)}
        className="text-xs px-2 py-1 rounded border border-border text-text-tertiary hover:text-text-secondary hover:border-border-strong transition-colors"
      >
        Block
      </button>
      {showConfirm && (
        <ConfirmDialog
          title={`Block @${username}?`}
          message="They won't be able to see your profile, reviews, or activity. Any existing follow relationship will be removed."
          confirmLabel="Block"
          onConfirm={handleBlock}
          onCancel={() => setShowConfirm(false)}
        />
      )}
    </>
  );
}
