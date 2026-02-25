"use client";

import { useState, useEffect } from "react";
import Link from "next/link";

type SimilarThread = {
  id: string;
  title: string;
  username: string;
  display_name: string | null;
  spoiler: boolean;
  created_at: string;
  comment_count: number;
  similarity: number;
};

type Props = {
  threadId: string;
  workId: string;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function SimilarThreads({ threadId, workId }: Props) {
  const [similar, setSimilar] = useState<SimilarThread[]>([]);

  useEffect(() => {
    fetch(`/api/threads/${threadId}/similar`)
      .then((res) => (res.ok ? res.json() : []))
      .then((data: SimilarThread[]) => setSimilar(data))
      .catch(() => {});
  }, [threadId]);

  if (similar.length === 0) return null;

  return (
    <aside className="border-t border-border pt-6 mt-8">
      <h3 className="text-xs font-semibold text-text-primary uppercase tracking-wider mb-3">
        Similar Discussions
      </h3>
      <div className="space-y-2">
        {similar.map((st) => (
          <Link
            key={st.id}
            href={`/books/${workId}/threads/${st.id}`}
            className="block border border-border rounded-lg px-3 py-2 hover:border-border transition-colors"
          >
            <p className="text-sm font-medium text-text-primary truncate">
              {st.spoiler && (
                <span className="text-[10px] font-medium text-amber-600 border border-amber-200 rounded px-1 py-0.5 mr-1 leading-none">
                  Spoiler
                </span>
              )}
              {st.title}
            </p>
            <div className="flex items-center gap-2 mt-0.5 text-xs text-text-primary">
              <span>{st.display_name ?? st.username}</span>
              <span>&middot;</span>
              <span>{formatDate(st.created_at)}</span>
              <span>&middot;</span>
              <span>
                {st.comment_count} {st.comment_count === 1 ? "reply" : "replies"}
              </span>
            </div>
          </Link>
        ))}
      </div>
    </aside>
  );
}
