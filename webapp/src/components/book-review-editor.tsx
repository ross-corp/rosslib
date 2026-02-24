"use client";

import { useState } from "react";
import StarRatingInput from "@/components/star-rating-input";

type Props = {
  shelfId: string;
  openLibraryId: string;
  initialRating: number | null;
  initialReviewText: string | null;
  initialSpoiler: boolean;
  initialDateRead: string | null;
};

export default function BookReviewEditor({
  shelfId,
  openLibraryId,
  initialRating,
  initialReviewText,
  initialSpoiler,
  initialDateRead,
}: Props) {
  const [rating, setRating] = useState<number | null>(initialRating);
  const [reviewText, setReviewText] = useState(initialReviewText ?? "");
  const [spoiler, setSpoiler] = useState(initialSpoiler);
  const [dateRead, setDateRead] = useState(
    initialDateRead ? initialDateRead.slice(0, 10) : ""
  );
  const [saving, setSaving] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const hasExisting = initialRating != null || (initialReviewText != null && initialReviewText !== "");
  const hasChanges =
    rating !== initialRating ||
    reviewText !== (initialReviewText ?? "") ||
    spoiler !== initialSpoiler ||
    dateRead !== (initialDateRead ? initialDateRead.slice(0, 10) : "");

  // savedRating tracks what's persisted so we know if a star click is a change
  const [savedRating, setSavedRating] = useState<number | null>(initialRating);

  async function patchFields(body: Record<string, unknown>) {
    setSaving(true);
    setMessage(null);

    const res = await fetch(
      `/api/shelves/${shelfId}/books/${openLibraryId}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }
    );

    setSaving(false);
    return res.ok;
  }

  async function handleRatingClick(newRating: number | null) {
    setRating(newRating);
    if (newRating === savedRating) return;
    const ok = await patchFields({ rating: newRating });
    if (ok) {
      setSavedRating(newRating);
      setMessage("Saved");
      setTimeout(() => setMessage(null), 2000);
    } else {
      setRating(savedRating);
      setMessage("Failed to save");
    }
  }

  async function save() {
    const body: Record<string, unknown> = {};
    if (rating !== savedRating) body.rating = rating;
    if (reviewText !== (initialReviewText ?? "")) {
      body.review_text = reviewText || null;
    }
    if (spoiler !== initialSpoiler) body.spoiler = spoiler;
    if (dateRead !== (initialDateRead ? initialDateRead.slice(0, 10) : "")) {
      body.date_read = dateRead || null;
    }

    const ok = await patchFields(body);
    if (ok) {
      setSavedRating(rating);
      setMessage("Saved");
      setTimeout(() => setMessage(null), 2000);
    } else {
      setMessage("Failed to save");
    }
  }

  async function clearReview() {
    if (!confirm("Clear your rating and review for this book?")) return;
    setSaving(true);
    setMessage(null);

    const res = await fetch(
      `/api/shelves/${shelfId}/books/${openLibraryId}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          rating: null,
          review_text: null,
          spoiler: false,
          date_read: null,
        }),
      }
    );

    setSaving(false);
    if (res.ok) {
      setRating(null);
      setReviewText("");
      setSpoiler(false);
      setDateRead("");
      setExpanded(false);
      setMessage("Review cleared");
      setTimeout(() => setMessage(null), 2000);
    } else {
      setMessage("Failed to clear");
    }
  }

  return (
    <div>
      {/* Inline star rating — always visible */}
      <div className="flex items-center gap-3">
        <span className="text-xs text-stone-500">Your rating</span>
        <StarRatingInput value={rating} onChange={handleRatingClick} disabled={saving} />
        {!expanded && !hasExisting && (
          <button
            type="button"
            onClick={() => setExpanded(true)}
            className="text-xs text-stone-400 hover:text-stone-600 transition-colors"
          >
            Write a review
          </button>
        )}
      </div>

      {/* Expanded review form */}
      {(expanded || hasExisting) && (
        <div className="mt-4 space-y-3">
          {!expanded && hasExisting ? (
            /* Collapsed view of existing review */
            <div>
              {initialReviewText && (
                <p className="text-sm text-stone-700 leading-relaxed whitespace-pre-wrap line-clamp-3">
                  {initialSpoiler ? (
                    <span className="text-stone-400 italic">Contains spoilers — </span>
                  ) : null}
                  {initialReviewText}
                </p>
              )}
              {initialDateRead && (
                <p className="text-xs text-stone-400 mt-1">
                  Read {new Date(initialDateRead).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </p>
              )}
              <button
                type="button"
                onClick={() => setExpanded(true)}
                className="text-xs text-stone-400 hover:text-stone-600 transition-colors mt-2"
              >
                Edit review
              </button>
            </div>
          ) : (
            /* Edit form */
            <>
              <textarea
                value={reviewText}
                onChange={(e) => setReviewText(e.target.value)}
                disabled={saving}
                placeholder="Write your review (optional)"
                rows={4}
                className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-400 resize-y disabled:opacity-50"
              />

              <div className="flex flex-wrap items-center gap-4 text-xs">
                <label className="flex items-center gap-1.5 text-stone-500">
                  <input
                    type="checkbox"
                    checked={spoiler}
                    onChange={(e) => setSpoiler(e.target.checked)}
                    disabled={saving}
                    className="rounded border-stone-300"
                  />
                  Contains spoilers
                </label>

                <label className="flex items-center gap-1.5 text-stone-500">
                  Date read
                  <input
                    type="date"
                    value={dateRead}
                    onChange={(e) => setDateRead(e.target.value)}
                    disabled={saving}
                    className="border border-stone-200 rounded px-2 py-1 text-xs text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
                  />
                </label>
              </div>

              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={save}
                  disabled={saving || !hasChanges}
                  className="text-xs px-3 py-1.5 rounded bg-stone-900 text-white hover:bg-stone-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {saving ? "Saving..." : "Save"}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setRating(initialRating);
                    setReviewText(initialReviewText ?? "");
                    setSpoiler(initialSpoiler);
                    setDateRead(initialDateRead ? initialDateRead.slice(0, 10) : "");
                    setExpanded(false);
                  }}
                  disabled={saving}
                  className="text-xs text-stone-400 hover:text-stone-600 transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
                {hasExisting && (
                  <button
                    type="button"
                    onClick={clearReview}
                    disabled={saving}
                    className="text-xs text-red-400 hover:text-red-600 transition-colors disabled:opacity-50 ml-auto"
                  >
                    Clear review
                  </button>
                )}
                {message && (
                  <span className="text-xs text-stone-500 ml-auto">{message}</span>
                )}
              </div>
            </>
          )}
        </div>
      )}

      {/* Save feedback when not expanded */}
      {!expanded && message && (
        <p className="text-xs text-stone-500 mt-1">{message}</p>
      )}
    </div>
  );
}
