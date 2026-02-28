"use client";

import { useEffect } from "react";
import { addToRecentlyViewed } from "@/lib/recently-viewed";

export default function RecordRecentlyViewed({
  workId,
  title,
  coverUrl,
}: {
  workId: string;
  title: string;
  coverUrl: string | null;
}) {
  useEffect(() => {
    addToRecentlyViewed({ workId, title, coverUrl });
  }, [workId, title, coverUrl]);

  return null;
}
