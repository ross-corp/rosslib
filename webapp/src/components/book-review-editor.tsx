"use client";

import { useState, useRef, useEffect } from "react";
import StarRatingInput from "@/components/star-rating-input";
import ReviewText from "@/components/review-text";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";
import { useToast } from "@/components/toast";

type BookSuggestion = {
  key: string;
  title: string;
  cover_url: string | null;
  authors: string[] | null;
  publish_year: number | null;
};

type Props = {
  openLibraryId: string;
  initialRating: number | null;
  initialReviewText: string | null;
  initialSpoiler: boolean;
  initialDateRead: string | null;
  initialDateDnf: string | null;
  initialDateStarted: string | null;
  statusSlug: string | null;
};

export default function BookReviewEditor({
  openLibraryId,
  initialRating,
  initialReviewText,
  initialSpoiler,
  initialDateRead,
  initialDateDnf,
  initialDateStarted,
  statusSlug,
}: Props) {
  const [rating, setRating] = useState<number | null>(initialRating);
  const [reviewText, setReviewText] = useState(initialReviewText ?? "");
  const [spoiler, setSpoiler] = useState(initialSpoiler);
  const [dateRead, setDateRead] = useState(
    initialDateRead ? initialDateRead.slice(0, 10) : ""
  );
  const [dateDnf, setDateDnf] = useState(
    initialDateDnf ? initialDateDnf.slice(0, 10) : ""
  );
  const [dateStarted, setDateStarted] = useState(
    initialDateStarted ? initialDateStarted.slice(0, 10) : ""
  );
  const [saving, setSaving] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  // Autocomplete state
  const [suggestions, setSuggestions] = useState<BookSuggestion[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestionQuery, setSuggestionQuery] = useState("");
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const hasExisting = initialRating != null || (initialReviewText != null && initialReviewText !== "");
  const hasChanges =
    rating !== initialRating ||
    reviewText !== (initialReviewText ?? "") ||
    spoiler !== initialSpoiler ||
    dateRead !== (initialDateRead ? initialDateRead.slice(0, 10) : "") ||
    dateDnf !== (initialDateDnf ? initialDateDnf.slice(0, 10) : "") ||
    dateStarted !== (initialDateStarted ? initialDateStarted.slice(0, 10) : "");

  // savedRating tracks what's persisted so we know if a star click is a change
  const [savedRating, setSavedRating] = useState<number | null>(initialRating);
  const toast = useToast();

  async function patchFields(body: Record<string, unknown>) {
    setSaving(true);
    setMessage(null);

    const res = await fetch(
      `/api/me/books/${openLibraryId}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }
    );

    setSaving(false);
    return res.ok;
  }

  // Clean up timeout on unmount
  useEffect(() => {
    return () => {
      if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
    };
  }, []);

  async function handleReviewChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const val = e.target.value;
    setReviewText(val);

    const cursor = e.target.selectionEnd;
    const lastOpen = val.lastIndexOf("[[", cursor);

    // If we found [[ and it's not closed/interrupted before the cursor
    if (lastOpen !== -1) {
      const textAfter = val.slice(lastOpen + 2, cursor);
      if (!textAfter.includes("]]") && !textAfter.includes("\n") && !textAfter.includes("|")) {
        setSuggestionQuery(textAfter);
        setShowSuggestions(true);

        if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
        searchTimeoutRef.current = setTimeout(() => {
          fetchSuggestions(textAfter);
        }, 300);
        return;
      }
    }
    setShowSuggestions(false);
  }

  async function fetchSuggestions(q: string) {
    if (q.length < 2) {
      setSuggestions([]);
      return;
    }
    try {
      const res = await fetch(`/api/books/search?q=${encodeURIComponent(q)}`);
      if (res.ok) {
        const data = await res.json();
        // Filter out the current book to avoid self-reference if desired,
        // but allowing it is also fine.
        setSuggestions(data.results.slice(0, 5));
      }
    } catch (e) {
      console.error(e);
    }
  }

  function selectSuggestion(book: BookSuggestion) {
    if (!textareaRef.current) return;

    const val = reviewText;
    const cursor = textareaRef.current.selectionEnd;
    const lastOpen = val.lastIndexOf("[[", cursor);

    if (lastOpen !== -1) {
      const bareOLID = book.key.replace("/works/", "");
      // Insert [Title](/books/OLID)
      const link = `[${book.title}](/books/${bareOLID})`;

      const newVal = val.slice(0, lastOpen) + link + val.slice(cursor);
      setReviewText(newVal);
      setShowSuggestions(false);
      setSuggestions([]);

      // Restore focus and move cursor after the link
      setTimeout(() => {
        if (textareaRef.current) {
            textareaRef.current.focus();
            const newCursor = lastOpen + link.length;
            textareaRef.current.setSelectionRange(newCursor, newCursor);
        }
      }, 0);
    }
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
    if (dateDnf !== (initialDateDnf ? initialDateDnf.slice(0, 10) : "")) {
      body.date_dnf = dateDnf || null;
    }
    if (dateStarted !== (initialDateStarted ? initialDateStarted.slice(0, 10) : "")) {
      body.date_started = dateStarted || null;
    }

    const ok = await patchFields(body);
    if (ok) {
      setSavedRating(rating);
      setMessage("Saved");
      setTimeout(() => setMessage(null), 2000);
      toast.success("Review saved");
    } else {
      setMessage("Failed to save");
      toast.error("Failed to save review");
    }
  }

  async function clearReview() {
    if (!confirm("Clear your rating and review for this book?")) return;
    setSaving(true);
    setMessage(null);

    const res = await fetch(
      `/api/me/books/${openLibraryId}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          rating: null,
          review_text: null,
          spoiler: false,
          date_read: null,
          date_dnf: null,
          date_started: null,
        }),
      }
    );

    setSaving(false);
    if (res.ok) {
      setRating(null);
      setReviewText("");
      setSpoiler(false);
      setDateRead("");
      setDateDnf("");
      setDateStarted("");
      setExpanded(false);
      setMessage("Review cleared");
      setTimeout(() => setMessage(null), 2000);
    } else {
      setMessage("Failed to clear");
    }
  }

  const showDateStarted = statusSlug === "currently-reading" || statusSlug === "finished";
  const showDateRead = statusSlug === "finished";
  const showDateDnf = statusSlug === "dnf";

  return (
    <div>
      {/* Inline star rating — always visible */}
      <div className="flex items-center gap-3">
        <span className="text-xs text-text-primary">Your rating</span>
        <StarRatingInput value={rating} onChange={handleRatingClick} disabled={saving} />
        {!expanded && !hasExisting && (
          <button
            type="button"
            onClick={() => setExpanded(true)}
            className="text-xs text-text-primary hover:text-text-primary transition-colors"
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
                <div className="text-sm text-text-primary leading-relaxed line-clamp-3">
                  {initialSpoiler ? (
                    <span className="text-text-primary italic">Contains spoilers — </span>
                  ) : null}
                  <ReviewText text={initialReviewText} />
                </div>
              )}
              {initialDateStarted && (
                <p className="text-xs text-text-primary mt-1">
                  Started {new Date(initialDateStarted).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </p>
              )}
              {initialDateRead && (
                <p className="text-xs text-text-primary mt-1">
                  Read {new Date(initialDateRead).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </p>
              )}
              {initialDateDnf && (
                <p className="text-xs text-text-primary mt-1">
                  Stopped {new Date(initialDateDnf).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </p>
              )}
              <button
                type="button"
                onClick={() => setExpanded(true)}
                className="text-xs text-text-primary hover:text-text-primary transition-colors mt-2"
              >
                Edit review
              </button>
            </div>
          ) : (
            /* Edit form */
            <div className="relative">
              <textarea
                ref={textareaRef}
                value={reviewText}
                onChange={handleReviewChange}
                disabled={saving}
                placeholder="Write your review (optional). Type [[ to link a book."
                rows={4}
                className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
              />

              {showSuggestions && suggestions.length > 0 && (
                <div className="absolute left-0 right-0 z-10 bg-surface-0 border border-border rounded shadow-lg max-h-60 overflow-y-auto mt-1">
                  {suggestions.map((book) => (
                    <button
                      key={book.key}
                      type="button"
                      onClick={() => selectSuggestion(book)}
                      className="w-full text-left px-3 py-2 hover:bg-surface-2 flex items-center gap-3 border-b border-border last:border-0"
                    >
                       {book.cover_url ? (
                         <img src={book.cover_url} alt="" className="w-8 h-12 object-cover rounded bg-surface-2" />
                       ) : (
                         <BookCoverPlaceholder title={book.title} className="w-8 h-12 flex-shrink-0" />
                       )}
                       <div>
                         <div className="text-sm font-medium text-text-primary line-clamp-1">{book.title}</div>
                         <div className="text-xs text-text-primary line-clamp-1">
                           {book.authors?.join(", ")}
                           {book.publish_year && ` · ${book.publish_year}`}
                         </div>
                       </div>
                    </button>
                  ))}
                </div>
              )}

              <div className="flex flex-wrap items-center gap-4 text-xs mt-3">
                <label className="flex items-center gap-1.5 text-text-primary">
                  <input
                    type="checkbox"
                    checked={spoiler}
                    onChange={(e) => setSpoiler(e.target.checked)}
                    disabled={saving}
                    className="rounded border-border"
                  />
                  Contains spoilers
                </label>

                {showDateStarted && (
                  <label className="flex items-center gap-1.5 text-text-primary">
                    Date started
                    <input
                      type="date"
                      value={dateStarted}
                      onChange={(e) => setDateStarted(e.target.value)}
                      disabled={saving}
                      className="border border-border rounded px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
                    />
                  </label>
                )}

                {showDateRead && (
                  <label className="flex items-center gap-1.5 text-text-primary">
                    Date read
                    <input
                      type="date"
                      value={dateRead}
                      onChange={(e) => setDateRead(e.target.value)}
                      disabled={saving}
                      className="border border-border rounded px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
                    />
                  </label>
                )}

                {showDateDnf && (
                  <label className="flex items-center gap-1.5 text-text-primary">
                    Date stopped
                    <input
                      type="date"
                      value={dateDnf}
                      onChange={(e) => setDateDnf(e.target.value)}
                      disabled={saving}
                      className="border border-border rounded px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
                    />
                  </label>
                )}
              </div>

              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={save}
                  disabled={saving || !hasChanges}
                  className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
                    setDateDnf(initialDateDnf ? initialDateDnf.slice(0, 10) : "");
                    setDateStarted(initialDateStarted ? initialDateStarted.slice(0, 10) : "");
                    setExpanded(false);
                  }}
                  disabled={saving}
                  className="text-xs text-text-primary hover:text-text-primary transition-colors disabled:opacity-50"
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
                  <span className="text-xs text-text-primary ml-auto">{message}</span>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Save feedback when not expanded */}
      {!expanded && message && (
        <p className="text-xs text-text-primary mt-1">{message}</p>
      )}
    </div>
  );
}
