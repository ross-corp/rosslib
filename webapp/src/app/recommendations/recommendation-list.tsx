"use client";

import { useState } from "react";
import Link from "next/link";
import { formatTime } from "@/components/activity";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";

type Recommendation = {
  id: string;
  note: string | null;
  status: string;
  created_at: string;
  sender: {
    user_id: string;
    username: string;
    display_name: string | null;
    avatar_url: string | null;
  };
  book: {
    open_library_id: string;
    title: string;
    cover_url: string | null;
    authors: string | null;
  };
};

export default function RecommendationList({
  recommendations: initial,
}: {
  recommendations: Recommendation[];
}) {
  const [recommendations, setRecommendations] =
    useState<Recommendation[]>(initial);
  const [updating, setUpdating] = useState<string | null>(null);

  async function updateStatus(id: string, status: string) {
    setUpdating(id);
    try {
      const res = await fetch(`/api/me/recommendations/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status }),
      });
      if (res.ok) {
        setRecommendations((prev) =>
          prev.map((r) => (r.id === id ? { ...r, status } : r))
        );
      }
    } finally {
      setUpdating(null);
    }
  }

  return (
    <div className="space-y-4">
      {recommendations.map((rec) => (
        <div
          key={rec.id}
          className={`flex gap-4 p-4 border border-border rounded-lg ${
            rec.status === "dismissed" ? "opacity-60" : ""
          }`}
        >
          {/* Book cover */}
          <Link
            href={`/books/${rec.book.open_library_id}`}
            className="shrink-0"
          >
            {rec.book.cover_url ? (
              <img
                src={rec.book.cover_url}
                alt={rec.book.title}
                className="w-16 h-24 rounded object-cover bg-surface-2"
              />
            ) : (
              <BookCoverPlaceholder title={rec.book.title} author={rec.book.authors} className="w-16 h-24" />
            )}
          </Link>

          <div className="flex-1 min-w-0">
            {/* Book title */}
            <Link
              href={`/books/${rec.book.open_library_id}`}
              className="text-sm font-medium text-text-primary hover:underline"
            >
              {rec.book.title}
            </Link>
            {rec.book.authors && (
              <p className="text-xs text-text-secondary mt-0.5">
                {rec.book.authors}
              </p>
            )}

            {/* Sender info */}
            <div className="flex items-center gap-2 mt-2">
              <Link href={`/${rec.sender.username}`} className="shrink-0">
                {rec.sender.avatar_url ? (
                  <img
                    src={rec.sender.avatar_url}
                    alt={rec.sender.display_name ?? rec.sender.username}
                    className="w-5 h-5 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-5 h-5 rounded-full bg-surface-2" />
                )}
              </Link>
              <Link
                href={`/${rec.sender.username}`}
                className="text-xs text-text-secondary hover:text-text-primary hover:underline"
              >
                {rec.sender.display_name ?? rec.sender.username}
              </Link>
              <span className="text-xs text-text-secondary">
                {formatTime(rec.created_at)}
              </span>
            </div>

            {/* Note */}
            {rec.note && (
              <p className="text-sm text-text-primary mt-2 italic">
                &ldquo;{rec.note}&rdquo;
              </p>
            )}

            {/* Actions */}
            {rec.status === "pending" && (
              <div className="flex items-center gap-2 mt-3">
                <Link
                  href={`/books/${rec.book.open_library_id}`}
                  className="text-xs font-medium text-surface-0 bg-text-primary rounded px-3 py-1.5 hover:opacity-90 transition-opacity"
                >
                  View book
                </Link>
                <button
                  onClick={() => updateStatus(rec.id, "seen")}
                  disabled={updating === rec.id}
                  className="text-xs text-text-secondary hover:text-text-primary border border-border rounded px-3 py-1.5 transition-colors disabled:opacity-50"
                >
                  Mark seen
                </button>
                <button
                  onClick={() => updateStatus(rec.id, "dismissed")}
                  disabled={updating === rec.id}
                  className="text-xs text-text-secondary hover:text-text-primary border border-border rounded px-3 py-1.5 transition-colors disabled:opacity-50"
                >
                  Dismiss
                </button>
              </div>
            )}

            {rec.status !== "pending" && (
              <span className="inline-block mt-2 text-[10px] font-medium text-text-secondary border border-border rounded px-1.5 py-0.5 leading-none">
                {rec.status.charAt(0).toUpperCase() + rec.status.slice(1)}
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
