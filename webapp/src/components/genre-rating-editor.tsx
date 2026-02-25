"use client";

import { useState, useEffect } from "react";

const GENRES = [
  "Fiction",
  "Non-fiction",
  "Fantasy",
  "Science fiction",
  "Mystery",
  "Romance",
  "Horror",
  "Thriller",
  "Biography",
  "History",
  "Poetry",
  "Children",
];

type AggregateRating = {
  genre: string;
  average: number;
  rater_count: number;
};

type MyRating = {
  genre: string;
  rating: number;
};

type Props = {
  openLibraryId: string;
  isLoggedIn: boolean;
  initialAggregateRatings: AggregateRating[];
  initialMyRatings: MyRating[];
};

export default function GenreRatingEditor({
  openLibraryId,
  isLoggedIn,
  initialAggregateRatings,
  initialMyRatings,
}: Props) {
  const [aggregateRatings, setAggregateRatings] =
    useState<AggregateRating[]>(initialAggregateRatings);
  const [myRatings, setMyRatings] = useState<Record<string, number>>(() => {
    const map: Record<string, number> = {};
    for (const r of initialMyRatings) {
      map[r.genre] = r.rating;
    }
    return map;
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [expanded, setExpanded] = useState(false);

  // Track saved state to detect changes
  const [savedRatings, setSavedRatings] = useState<Record<string, number>>(
    () => {
      const map: Record<string, number> = {};
      for (const r of initialMyRatings) {
        map[r.genre] = r.rating;
      }
      return map;
    }
  );

  const hasChanges =
    JSON.stringify(myRatings) !== JSON.stringify(savedRatings);
  const hasAnyRating = Object.keys(myRatings).length > 0;

  function handleSliderChange(genre: string, value: number) {
    setMyRatings((prev) => {
      const next = { ...prev };
      if (value === 0) {
        delete next[genre];
      } else {
        next[genre] = value;
      }
      return next;
    });
  }

  async function save() {
    setSaving(true);
    setMessage(null);

    const payload = GENRES.map((genre) => ({
      genre,
      rating: myRatings[genre] ?? 0,
    })).filter((r) => r.rating > 0 || savedRatings[r.genre]);

    const res = await fetch(`/api/me/books/${openLibraryId}/genre-ratings`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    setSaving(false);

    if (res.ok) {
      setSavedRatings({ ...myRatings });
      setMessage("Saved");
      setTimeout(() => setMessage(null), 2000);
      // Refresh aggregate ratings
      refreshAggregate();
    } else {
      setMessage("Failed to save");
    }
  }

  async function refreshAggregate() {
    const res = await fetch(`/api/books/${openLibraryId}/genre-ratings`);
    if (res.ok) {
      setAggregateRatings(await res.json());
    }
  }

  // Build the aggregate map for display
  const aggregateMap: Record<string, AggregateRating> = {};
  for (const r of aggregateRatings) {
    aggregateMap[r.genre] = r;
  }

  // Show aggregate ratings that have at least one rater
  const hasAggregate = aggregateRatings.length > 0;

  return (
    <div>
      {/* ── Aggregate display ── */}
      {hasAggregate && (
        <div className="space-y-2 mb-4">
          {aggregateRatings.map((r) => (
            <div key={r.genre} className="flex items-center gap-3">
              <span className="text-xs text-text-primary w-28 shrink-0 truncate">
                {r.genre}
              </span>
              <div className="flex-1 h-2 bg-surface-2 rounded-full overflow-hidden">
                <div
                  className="h-full bg-text-tertiary rounded-full transition-all"
                  style={{ width: `${(r.average / 10) * 100}%` }}
                />
              </div>
              <span className="text-xs text-text-primary w-16 text-right shrink-0">
                {r.average.toFixed(1)}/10
              </span>
              <span className="text-[10px] text-text-primary w-10 text-right shrink-0">
                ({r.rater_count})
              </span>
            </div>
          ))}
        </div>
      )}

      {!hasAggregate && !isLoggedIn && (
        <p className="text-text-primary text-sm">No genre ratings yet.</p>
      )}

      {/* ── User rating editor ── */}
      {isLoggedIn && !expanded && (
        <button
          type="button"
          onClick={() => setExpanded(true)}
          className="text-xs text-text-primary hover:text-text-primary transition-colors"
        >
          {hasAnyRating ? "Edit your genre ratings" : "Rate genres"}
        </button>
      )}

      {isLoggedIn && expanded && (
        <div className="mt-3 space-y-3 border border-border rounded-lg p-4 bg-surface-2/50">
          <p className="text-xs text-text-primary mb-2">
            How strongly does this book fit each genre? (0 = not at all, 10 =
            defining example)
          </p>

          <div className="space-y-2">
            {GENRES.map((genre) => (
              <div key={genre} className="flex items-center gap-3">
                <span className="text-xs text-text-primary w-28 shrink-0 truncate">
                  {genre}
                </span>
                <input
                  type="range"
                  min="0"
                  max="10"
                  value={myRatings[genre] ?? 0}
                  onChange={(e) =>
                    handleSliderChange(genre, parseInt(e.target.value))
                  }
                  disabled={saving}
                  className="flex-1 h-1.5 accent-neutral-400"
                />
                <span className="text-xs text-text-primary w-6 text-right tabular-nums">
                  {myRatings[genre] ?? 0}
                </span>
              </div>
            ))}
          </div>

          <div className="flex items-center gap-3 pt-2">
            <button
              type="button"
              onClick={save}
              disabled={saving || !hasChanges}
              className="text-xs px-3 py-1.5 rounded bg-accent text-white hover:bg-surface-3 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {saving ? "Saving..." : "Save"}
            </button>
            <button
              type="button"
              onClick={() => {
                setMyRatings({ ...savedRatings });
                setExpanded(false);
              }}
              disabled={saving}
              className="text-xs text-text-primary hover:text-text-primary transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            {message && (
              <span className="text-xs text-text-primary ml-auto">{message}</span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
