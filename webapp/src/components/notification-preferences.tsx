"use client";

import { useState, useEffect } from "react";

type Preferences = {
  new_publication: boolean;
  book_new_thread: boolean;
  book_new_link: boolean;
  book_new_review: boolean;
  review_liked: boolean;
  thread_mention: boolean;
  book_recommendation: boolean;
};

const PREF_LABELS: { key: keyof Preferences; label: string; description: string }[] = [
  { key: "new_publication", label: "New publications", description: "When an author you follow publishes a new work" },
  { key: "book_new_thread", label: "New threads", description: "When a new discussion is started on a book you follow" },
  { key: "book_new_link", label: "New links", description: "When a new community link is added to a book you follow" },
  { key: "book_new_review", label: "New reviews", description: "When a new review is posted on a book you follow" },
  { key: "review_liked", label: "Review likes", description: "When someone likes your review" },
  { key: "thread_mention", label: "Mentions", description: "When someone @mentions you in a thread comment" },
  { key: "book_recommendation", label: "Recommendations", description: "When someone recommends a book to you" },
];

const ALL_TRUE: Preferences = {
  new_publication: true,
  book_new_thread: true,
  book_new_link: true,
  book_new_review: true,
  review_liked: true,
  thread_mention: true,
  book_recommendation: true,
};

export default function NotificationPreferences() {
  const [prefs, setPrefs] = useState<Preferences>(ALL_TRUE);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      const res = await fetch("/api/me/notification-preferences");
      if (res.ok) {
        setPrefs(await res.json());
      }
      setLoading(false);
    }
    load();
  }, []);

  async function toggle(key: keyof Preferences) {
    const newValue = !prefs[key];
    setSaving(key);

    // Optimistic update
    setPrefs((prev) => ({ ...prev, [key]: newValue }));

    const res = await fetch("/api/me/notification-preferences", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ [key]: newValue }),
    });

    if (res.ok) {
      const updated = await res.json();
      setPrefs(updated);
    } else {
      // Revert on failure
      setPrefs((prev) => ({ ...prev, [key]: !newValue }));
    }

    setSaving(null);
  }

  if (loading) {
    return (
      <div className="mt-10 border-t border-border pt-8">
        <h2 className="text-lg font-semibold text-text-primary mb-4">
          Notification preferences
        </h2>
        <p className="text-sm text-text-primary">Loading...</p>
      </div>
    );
  }

  return (
    <div className="mt-10 border-t border-border pt-8">
      <h2 className="text-lg font-semibold text-text-primary mb-2">
        Notification preferences
      </h2>
      <p className="text-sm text-text-primary mb-6">
        Choose which notifications you want to receive.
      </p>

      <div className="space-y-4">
        {PREF_LABELS.map(({ key, label, description }) => (
          <div
            key={key}
            className="flex items-center justify-between gap-4"
          >
            <div className="min-w-0">
              <p className="text-sm font-medium text-text-primary">{label}</p>
              <p className="text-xs text-text-primary">{description}</p>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={prefs[key]}
              onClick={() => toggle(key)}
              disabled={saving === key}
              className={`relative shrink-0 inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 focus:ring-offset-surface-0 disabled:opacity-50 ${
                prefs[key] ? "bg-accent" : "bg-surface-3"
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  prefs[key] ? "translate-x-6" : "translate-x-1"
                }`}
              />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
