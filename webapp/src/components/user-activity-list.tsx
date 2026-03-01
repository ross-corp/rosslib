"use client";

import { useState } from "react";
import { ActivityCard } from "@/components/activity";
import type { ActivityItem } from "@/components/activity";

export default function UserActivityList({
  initialActivities,
  initialCursor,
  username,
  showUser = false,
}: {
  initialActivities: ActivityItem[];
  initialCursor?: string | null;
  username: string;
  showUser?: boolean;
}) {
  const [activities, setActivities] = useState(initialActivities);
  const [cursor, setCursor] = useState<string | null>(initialCursor ?? null);
  const [loading, setLoading] = useState(false);

  async function loadMore() {
    if (!cursor || loading) return;
    setLoading(true);
    try {
      const res = await fetch(
        `/api/users/${username}/activity?cursor=${encodeURIComponent(cursor)}`
      );
      if (!res.ok) return;
      const data = await res.json();
      const newActivities: ActivityItem[] = data.activities || [];
      setActivities((prev) => [...prev, ...newActivities]);
      setCursor(data.next_cursor ?? null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      {activities.map((item) => (
        <ActivityCard key={item.id} item={item} showUser={showUser} />
      ))}
      {cursor && (
        <div className="mt-2 text-center">
          <button
            onClick={loadMore}
            disabled={loading}
            className="text-sm text-text-tertiary hover:text-text-primary transition-colors"
          >
            {loading ? "Loading..." : "Load more"}
          </button>
        </div>
      )}
    </div>
  );
}
