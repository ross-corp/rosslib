"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

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
  notifications,
  nextCursor,
}: {
  notifications: Notification[];
  nextCursor?: string;
}) {
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

  if (!hasUnread) return null;

  return (
    <button
      onClick={markAllRead}
      disabled={loading}
      className="text-sm text-text-primary hover:text-text-primary border border-border px-3 py-1.5 rounded transition-colors disabled:opacity-50"
    >
      {loading ? "..." : "Mark all read"}
    </button>
  );
}
