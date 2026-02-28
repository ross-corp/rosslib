"use client";

import { useState, useEffect, useRef } from "react";
import Link from "next/link";

type Thread = {
  id: string;
  book_id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  title: string;
  body: string;
  spoiler: boolean;
  created_at: string;
  comment_count: number;
};

type SimilarThread = Thread & { similarity: number };

type Props = {
  workId: string;
  initialThreads: Thread[];
  isLoggedIn: boolean;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function ThreadList({ workId, initialThreads, isLoggedIn }: Props) {
  const [threads, setThreads] = useState<Thread[]>(initialThreads);
  const [showForm, setShowForm] = useState(false);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [spoiler, setSpoiler] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [similarThreads, setSimilarThreads] = useState<SimilarThread[]>([]);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Debounced search for similar threads as user types title
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);

    if (title.trim().length < 5) {
      setSimilarThreads([]);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      try {
        const res = await fetch(
          `/api/books/${workId}/similar-threads?title=${encodeURIComponent(title.trim())}`
        );
        if (res.ok) {
          const data: SimilarThread[] = await res.json();
          setSimilarThreads(data);
        }
      } catch {
        // Ignore network errors for suggestions
      }
    }, 400);

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [title, workId]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim() || !body.trim()) return;

    setSubmitting(true);
    setError(null);

    const res = await fetch(`/api/books/${workId}/threads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: title.trim(), body: body.trim(), spoiler }),
    });

    setSubmitting(false);

    if (!res.ok) {
      setError("Failed to create thread");
      return;
    }

    // Refetch threads to get the full thread data.
    const listRes = await fetch(`/api/books/${workId}/threads`);
    if (listRes.ok) {
      setThreads(await listRes.json());
    }

    setTitle("");
    setBody("");
    setSpoiler(false);
    setSimilarThreads([]);
    setShowForm(false);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
          {threads.length > 0
            ? `Discussions (${threads.length})`
            : "Discussions"}
        </h2>
        {isLoggedIn && !showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover transition-colors"
          >
            New thread
          </button>
        )}
      </div>

      {/* New thread form */}
      {showForm && (
        <form onSubmit={handleSubmit} className="mb-8 space-y-3">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Thread title"
            disabled={submitting}
            className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
          />

          {/* Similar thread suggestions */}
          {similarThreads.length > 0 && (
            <div className="border border-amber-200 bg-amber-50 rounded-lg p-3">
              <p className="text-xs font-medium text-amber-700 mb-2">
                Similar discussions already exist:
              </p>
              <div className="space-y-2">
                {similarThreads.map((st) => (
                  <Link
                    key={st.id}
                    href={`/books/${workId}/threads/${st.id}`}
                    className="block text-sm text-amber-900 hover:text-amber-700 transition-colors"
                  >
                    <span className="font-medium">{st.title}</span>
                    <span className="text-xs text-amber-500 ml-2">
                      {st.comment_count} {st.comment_count === 1 ? "reply" : "replies"}
                    </span>
                  </Link>
                ))}
              </div>
            </div>
          )}

          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="What do you want to discuss?"
            rows={4}
            disabled={submitting}
            className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
          />
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-1.5 text-xs text-text-primary">
              <input
                type="checkbox"
                checked={spoiler}
                onChange={(e) => setSpoiler(e.target.checked)}
                disabled={submitting}
                className="rounded border-border"
              />
              Contains spoilers
            </label>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !title.trim() || !body.trim()}
              className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "Posting..." : "Post thread"}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowForm(false);
                setTitle("");
                setBody("");
                setSpoiler(false);
                setError(null);
                setSimilarThreads([]);
              }}
              disabled={submitting}
              className="text-xs text-text-primary hover:text-text-primary transition-colors"
            >
              Cancel
            </button>
            {error && <span className="text-xs text-semantic-error">{error}</span>}
          </div>
        </form>
      )}

      {/* Thread list */}
      {threads.length === 0 ? (
        <p className="text-text-primary text-sm">No discussions yet.</p>
      ) : (
        <div className="space-y-4">
          {threads.map((thread) => (
            <Link
              key={thread.id}
              href={`/books/${workId}/threads/${thread.id}`}
              className="block border border-border rounded-lg p-4 hover:border-border transition-colors"
            >
              <div className="flex items-start gap-3">
                {thread.avatar_url ? (
                  <img
                    src={thread.avatar_url}
                    alt={thread.display_name ?? thread.username}
                    className="w-6 h-6 rounded-full object-cover shrink-0 mt-0.5"
                  />
                ) : (
                  <div className="w-6 h-6 rounded-full bg-surface-2 shrink-0 mt-0.5" />
                )}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-medium text-text-primary truncate">
                      {thread.spoiler && (
                        <span className="text-[10px] font-medium text-amber-600 border border-amber-200 rounded px-1 py-0.5 mr-2 leading-none">
                          Spoiler
                        </span>
                      )}
                      {thread.title}
                    </h3>
                  </div>
                  <div className="flex items-center gap-2 mt-1 text-xs text-text-primary">
                    <span>{thread.display_name ?? thread.username}</span>
                    <span>&middot;</span>
                    <span>{formatDate(thread.created_at)}</span>
                    <span>&middot;</span>
                    <span>
                      {thread.comment_count}{" "}
                      {thread.comment_count === 1 ? "reply" : "replies"}
                    </span>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
