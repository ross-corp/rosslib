"use client";

import { useState } from "react";
import Link from "next/link";
import { Stars } from "@/components/activity";
import ReviewText from "@/components/review-text";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";

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
  date_dnf: string | null;
  date_added: string;
  like_count?: number;
};

function ReviewCard({ review }: { review: ReviewItem }) {
  const [revealed, setRevealed] = useState(false);

  return (
    <div className="flex gap-3 py-4 border-b border-border last:border-0">
      <Link href={`/books/${review.open_library_id}`} className="shrink-0">
        {review.cover_url ? (
          <img
            src={review.cover_url}
            alt=""
            className="w-10 h-14 object-cover rounded shadow-sm"
          />
        ) : (
          <BookCoverPlaceholder title={review.title} author={review.authors} className="w-10 h-14" />
        )}
      </Link>
      <div className="flex-1 min-w-0">
        <Link
          href={`/books/${review.open_library_id}`}
          className="text-sm font-medium text-text-primary hover:underline line-clamp-1"
        >
          {review.title}
        </Link>
        {review.authors && (
          <p className="text-xs text-text-tertiary">{review.authors}</p>
        )}
        {review.rating && (
          <div className="mt-0.5">
            <Stars rating={review.rating} />
          </div>
        )}
        {review.spoiler && !revealed ? (
          <button
            onClick={() => setRevealed(true)}
            className="mt-1 text-xs text-text-tertiary hover:text-text-secondary italic"
          >
            Spoiler — click to reveal
          </button>
        ) : (
          <div className="mt-1 text-sm text-text-secondary leading-snug line-clamp-3">
            <ReviewText text={review.review_text} />
          </div>
        )}
        {(review.like_count ?? 0) > 0 && (
          <span className="inline-flex items-center gap-1 mt-1 text-xs text-text-tertiary">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
            </svg>
            {review.like_count}
          </span>
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
      {reviews.map((r, i) => (
        <ReviewCard key={`${r.book_id}-${i}`} review={r} />
      ))}
      <Link
        href={`/${username}/reviews`}
        className="block text-center text-xs text-text-tertiary hover:text-text-secondary transition-colors pt-2"
      >
        See all reviews &rarr;
      </Link>
    </div>
  );
}
