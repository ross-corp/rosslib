"use client";

import Link from "next/link";
import { formatTime } from "@/components/activity";

type SentRecommendation = {
  id: string;
  note: string | null;
  status: string;
  created_at: string;
  recipient: {
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

export default function SentRecommendationList({
  recommendations,
}: {
  recommendations: SentRecommendation[];
}) {
  const statusColors: Record<string, string> = {
    pending: "text-amber-600 border-amber-300",
    seen: "text-green-600 border-green-300",
    dismissed: "text-text-secondary border-border",
  };

  return (
    <div className="space-y-4">
      {recommendations.map((rec) => (
        <div
          key={rec.id}
          className="flex gap-4 p-4 border border-border rounded-lg"
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
              <div className="w-16 h-24 rounded bg-surface-2" />
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

            {/* Recipient info */}
            <div className="flex items-center gap-2 mt-2">
              <span className="text-xs text-text-secondary">Sent to</span>
              <Link href={`/${rec.recipient.username}`} className="shrink-0">
                {rec.recipient.avatar_url ? (
                  <img
                    src={rec.recipient.avatar_url}
                    alt={
                      rec.recipient.display_name ?? rec.recipient.username
                    }
                    className="w-5 h-5 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-5 h-5 rounded-full bg-surface-2" />
                )}
              </Link>
              <Link
                href={`/${rec.recipient.username}`}
                className="text-xs text-text-secondary hover:text-text-primary hover:underline"
              >
                {rec.recipient.display_name ?? rec.recipient.username}
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

            {/* Status badge */}
            <span
              className={`inline-block mt-2 text-[10px] font-medium border rounded px-1.5 py-0.5 leading-none ${statusColors[rec.status] ?? "text-text-secondary border-border"}`}
            >
              {rec.status.charAt(0).toUpperCase() + rec.status.slice(1)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}
