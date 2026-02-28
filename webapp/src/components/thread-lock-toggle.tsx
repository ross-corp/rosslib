"use client";

import { useState } from "react";

type Props = {
  threadId: string;
  initialLockedAt: string | null;
};

export default function ThreadLockToggle({ threadId, initialLockedAt }: Props) {
  const [lockedAt, setLockedAt] = useState<string | null>(initialLockedAt);
  const [loading, setLoading] = useState(false);

  async function handleToggle() {
    setLoading(true);
    const endpoint = lockedAt ? "unlock" : "lock";
    const res = await fetch(`/api/threads/${threadId}/${endpoint}`, {
      method: "POST",
    });
    if (res.ok) {
      const data = await res.json();
      setLockedAt(data.locked_at ?? null);
    }
    setLoading(false);
  }

  return (
    <button
      type="button"
      onClick={handleToggle}
      disabled={loading}
      className="shrink-0 text-xs px-3 py-1.5 rounded border border-border text-text-primary hover:bg-surface-2 disabled:opacity-50 transition-colors"
    >
      {loading ? (lockedAt ? "Unlocking..." : "Locking...") : (
        <>
          <svg className="inline w-3.5 h-3.5 mr-1 align-text-bottom" viewBox="0 0 20 20" fill="currentColor">
            {lockedAt ? (
              <path d="M10 1a4.5 4.5 0 00-4.5 4.5V9H5a2 2 0 00-2 2v6a2 2 0 002 2h10a2 2 0 002-2v-6a2 2 0 00-2-2h-.5V5.5a3 3 0 10-6 0v.5a.75.75 0 001.5 0v-.5a1.5 1.5 0 113 0V9h-8z" />
            ) : (
              <path fillRule="evenodd" d="M10 1a4.5 4.5 0 00-4.5 4.5V9H5a2 2 0 00-2 2v6a2 2 0 002 2h10a2 2 0 002-2v-6a2 2 0 00-2-2h-.5V5.5A4.5 4.5 0 0010 1zm3 8V5.5a3 3 0 10-6 0V9h6z" clipRule="evenodd" />
            )}
          </svg>
          {lockedAt ? "Unlock" : "Lock"}
        </>
      )}
    </button>
  );
}
