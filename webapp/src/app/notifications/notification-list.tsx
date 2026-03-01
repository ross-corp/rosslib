"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import NotificationCard from "./notification-card";

type Notification = {
  id: string;
  notif_type: string;
  title: string;
  body: string | null;
  metadata: Record<string, string> | null;
  read: boolean;
  created_at: string;
};

export default function NotificationList({
  notifications: initialNotifications,
  nextCursor,
}: {
  notifications: Notification[];
  nextCursor?: string;
}) {
  const [notifications, setNotifications] = useState(initialNotifications);
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const hasUnread = notifications.some((n) => !n.read);

  async function markAllRead() {
    setLoading(true);
    try {
      await fetch("/api/me/notifications/read-all", { method: "POST" });
      router.refresh();
    } finally {
      setLoading(false);
    }
  }

  function handleDelete(id: string) {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  }

  return (
    <>
      {hasUnread && (
        <div className="flex justify-end mb-4">
          <button
            onClick={markAllRead}
            disabled={loading}
            className="text-sm text-text-primary hover:text-text-primary border border-border px-3 py-1.5 rounded transition-colors disabled:opacity-50"
          >
            {loading ? "..." : "Mark all read"}
          </button>
        </div>
      )}

      {notifications.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-text-primary text-sm">
            No notifications yet. Follow authors to get notified about
            new publications, or follow books to hear about new discussions,
            links, and reviews.
          </p>
          <Link
            href="/feed"
            className="inline-block mt-4 text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
          >
            Go to feed
          </Link>
        </div>
      ) : (
        <>
          <div className="divide-y divide-border">
            {notifications.map((notif) => (
              <NotificationCard
                key={notif.id}
                notif={notif}
                onDelete={handleDelete}
              />
            ))}
          </div>

          {nextCursor && (
            <div className="mt-8 text-center">
              <Link
                href={`/notifications?cursor=${encodeURIComponent(nextCursor)}`}
                className="text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
              >
                Load more
              </Link>
            </div>
          )}
        </>
      )}
    </>
  );
}
