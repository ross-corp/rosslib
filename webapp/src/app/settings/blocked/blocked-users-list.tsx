"use client";

import { useState } from "react";
import Link from "next/link";

type BlockedUser = {
  id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

export default function BlockedUsersList({
  initialUsers,
}: {
  initialUsers: BlockedUser[];
}) {
  const [users, setUsers] = useState(initialUsers);
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  async function unblock(username: string) {
    setLoading((prev) => ({ ...prev, [username]: true }));
    const res = await fetch(`/api/users/${username}/block`, {
      method: "DELETE",
    });
    if (res.ok) {
      setUsers((prev) => prev.filter((u) => u.username !== username));
    }
    setLoading((prev) => ({ ...prev, [username]: false }));
  }

  if (users.length === 0) {
    return (
      <p className="text-sm text-text-primary">
        You haven&apos;t blocked anyone.
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {users.map((user) => (
        <div
          key={user.id}
          className="flex items-center justify-between py-3 border-b border-border"
        >
          <div className="flex items-center gap-3">
            {user.avatar_url ? (
              <img
                src={user.avatar_url}
                alt=""
                className="w-10 h-10 rounded-full object-cover bg-surface-2"
              />
            ) : (
              <div className="w-10 h-10 rounded-full bg-surface-2 flex items-center justify-center">
                <span className="text-text-primary text-sm font-medium select-none">
                  {user.username[0].toUpperCase()}
                </span>
              </div>
            )}
            <div>
              <Link
                href={`/${user.username}`}
                className="text-sm font-medium text-text-primary hover:underline"
              >
                {user.display_name || user.username}
              </Link>
              {user.display_name && (
                <p className="text-xs text-text-primary">@{user.username}</p>
              )}
            </div>
          </div>
          <button
            onClick={() => unblock(user.username)}
            disabled={loading[user.username]}
            className="text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
          >
            Unblock
          </button>
        </div>
      ))}
    </div>
  );
}
