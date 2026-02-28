"use client";

import { useState } from "react";
import Link from "next/link";

type FollowRequest = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  created_at: string;
};

export default function FollowRequestsList({
  initialRequests,
}: {
  initialRequests: FollowRequest[];
}) {
  const [requests, setRequests] = useState(initialRequests);
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  async function accept(userId: string) {
    setLoading((prev) => ({ ...prev, [userId]: true }));
    const res = await fetch(`/api/me/follow-requests/${userId}/accept`, {
      method: "POST",
    });
    if (res.ok) {
      setRequests((prev) => prev.filter((r) => r.user_id !== userId));
    }
    setLoading((prev) => ({ ...prev, [userId]: false }));
  }

  async function reject(userId: string) {
    setLoading((prev) => ({ ...prev, [userId]: true }));
    const res = await fetch(`/api/me/follow-requests/${userId}/reject`, {
      method: "DELETE",
    });
    if (res.ok) {
      setRequests((prev) => prev.filter((r) => r.user_id !== userId));
    }
    setLoading((prev) => ({ ...prev, [userId]: false }));
  }

  if (requests.length === 0) {
    return (
      <p className="text-sm text-text-primary">No pending follow requests.</p>
    );
  }

  return (
    <div className="space-y-3">
      {requests.map((req) => (
        <div
          key={req.user_id}
          className="flex items-center justify-between py-3 border-b border-border"
        >
          <div className="flex items-center gap-3">
            {req.avatar_url ? (
              <img
                src={req.avatar_url}
                alt=""
                className="w-10 h-10 rounded-full object-cover bg-surface-2"
              />
            ) : (
              <div className="w-10 h-10 rounded-full bg-surface-2 flex items-center justify-center">
                <span className="text-text-primary text-sm font-medium select-none">
                  {req.username[0].toUpperCase()}
                </span>
              </div>
            )}
            <div>
              <Link
                href={`/${req.username}`}
                className="text-sm font-medium text-text-primary hover:underline"
              >
                {req.display_name || req.username}
              </Link>
              {req.display_name && (
                <p className="text-xs text-text-primary">@{req.username}</p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => accept(req.user_id)}
              disabled={loading[req.user_id]}
              className="text-sm px-3 py-1.5 rounded border border-accent bg-accent text-text-inverted hover:bg-accent-hover transition-colors disabled:opacity-50"
            >
              Accept
            </button>
            <button
              onClick={() => reject(req.user_id)}
              disabled={loading[req.user_id]}
              className="text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
            >
              Reject
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
