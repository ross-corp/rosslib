"use client";

import { useState } from "react";
import Link from "next/link";

type FollowedAuthor = {
  author_key: string;
  author_name: string;
};

export default function FollowedAuthorsList({
  initialAuthors,
  initialTotal,
}: {
  initialAuthors: FollowedAuthor[];
  initialTotal: number;
}) {
  const [authors, setAuthors] = useState(initialAuthors);
  const [total, setTotal] = useState(initialTotal);
  const [loading, setLoading] = useState<Record<string, boolean>>({});
  const [loadingMore, setLoadingMore] = useState(false);

  async function unfollow(authorKey: string) {
    setLoading((prev) => ({ ...prev, [authorKey]: true }));
    const res = await fetch(`/api/authors/${authorKey}/follow`, {
      method: "DELETE",
    });
    if (res.ok) {
      setAuthors((prev) => prev.filter((a) => a.author_key !== authorKey));
      setTotal((prev) => prev - 1);
    }
    setLoading((prev) => ({ ...prev, [authorKey]: false }));
  }

  async function loadMore() {
    setLoadingMore(true);
    const res = await fetch(
      `/api/me/followed-authors?limit=50&offset=${authors.length}`
    );
    if (res.ok) {
      const data = await res.json();
      setAuthors((prev) => [...prev, ...data.authors]);
      setTotal(data.total);
    }
    setLoadingMore(false);
  }

  if (authors.length === 0) {
    return (
      <div className="text-sm text-text-primary space-y-2">
        <p>
          You aren&apos;t following any authors yet. Follow authors on their profile pages to get notified about new publications.
        </p>
        <p>
          <Link
            href="/search?tab=authors"
            className="text-accent hover:underline"
          >
            Browse authors
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {authors.map((author) => (
        <div
          key={author.author_key}
          className="flex items-center justify-between py-3 border-b border-border"
        >
          <div className="min-w-0">
            <Link
              href={`/authors/${author.author_key}`}
              className="text-sm font-medium text-text-primary hover:underline line-clamp-1"
            >
              {author.author_name || author.author_key}
            </Link>
          </div>
          <button
            onClick={() => unfollow(author.author_key)}
            disabled={loading[author.author_key]}
            className="shrink-0 text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
          >
            Unfollow
          </button>
        </div>
      ))}
      {authors.length < total && (
        <div className="pt-4 text-center">
          <button
            onClick={loadMore}
            disabled={loadingMore}
            className="text-sm px-4 py-2 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
          >
            {loadingMore ? "Loading..." : "Load more"}
          </button>
        </div>
      )}
    </div>
  );
}
