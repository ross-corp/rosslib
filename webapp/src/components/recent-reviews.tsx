"use client";

import { useState } from "react";
import Link from "next/link";
import { Stars } from "@/components/activity";

export type ReviewItem = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  authors: string | null;
  rating: number | null;
  review_text: string;
  spoiler: boolean;
  date_read: string | null;
  date_added: string;
};

function ReviewCard({ review }: { review: ReviewItem }) {
  const [revealed, setRevealed] = useState(false);
  const snippet =
    review.review_text.length > 150
      ? review.review_text.slice(0, 150) + "..."
      : review.review_text;

  return (
    <div className="flex gap-3 py-4 border-b border-stone-100 last:border-0">
      <Link href={`/books/${review.open_library_id}`} className="shrink-0">
        {review.cover_url ? (
          <img
            src={review.cover_url}
            alt=""
            className="w-10 h-14 object-cover rounded shadow-sm"
          />
        ) : (
          <div className="w-10 h-14 rounded bg-stone-200" />
        )}
      </Link>
      <div className="flex-1 min-w-0">
        <Link
          href={`/books/${review.open_library_id}`}
          className="text-sm font-medium text-stone-900 hover:underline line-clamp-1"
        >
          {review.title}
        </Link>
        {review.authors && (
          <p className="text-xs text-stone-400">{review.authors}</p>
        )}
        {review.rating && (
          <div className="mt-0.5">
            <Stars rating={review.rating} />
          </div>
        )}
        {review.spoiler && !revealed ? (
          <button
            onClick={() => setRevealed(true)}
            className="mt-1 text-xs text-stone-400 hover:text-stone-600 italic"
          >
            Spoiler — click to reveal
          </button>
        ) : (
          <p className="mt-1 text-sm text-stone-600 leading-snug">{snippet}</p>
        )}
      </div>
    </div>
  );
}

export default function RecentReviews({
  reviews,
  username,
}: {
  reviews: ReviewItem[];
  username: string;
}) {
  if (reviews.length === 0) return null;

  return (
    <div>
      {reviews.map((r) => (
        <ReviewCard key={r.book_id} review={r} />
      ))}
      <Link
        href={`/${username}/reviews`}
        className="block text-center text-xs text-stone-400 hover:text-stone-700 transition-colors pt-2"
      >
        See all reviews &rarr;
      </Link>
    </div>
  );
}
