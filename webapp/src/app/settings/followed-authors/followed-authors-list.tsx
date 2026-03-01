"use client";

import { useState } from "react";
import Link from "next/link";

type FollowedAuthor = {
  author_key: string;
  author_name: string;
};

export default function FollowedAuthorsList({
  initialAuthors,
}: {
  initialAuthors: FollowedAuthor[];
}) {
  const [authors, setAuthors] = useState(initialAuthors);
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  async function unfollow(authorKey: string) {
    setLoading((prev) => ({ ...prev, [authorKey]: true }));
    const res = await fetch(`/api/authors/${authorKey}/follow`, {
      method: "DELETE",
    });
    if (res.ok) {
      setAuthors((prev) => prev.filter((a) => a.author_key !== authorKey));
    }
    setLoading((prev) => ({ ...prev, [authorKey]: false }));
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
    </div>
  );
}
