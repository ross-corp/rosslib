"use client";

import { useState } from "react";
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
    setShowForm(false);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider">
          {threads.length > 0
            ? `Discussions (${threads.length})`
            : "Discussions"}
        </h2>
        {isLoggedIn && !showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded bg-stone-900 text-white hover:bg-stone-700 transition-colors"
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
            className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
          />
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="What do you want to discuss?"
            rows={4}
            disabled={submitting}
            className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-400 resize-y disabled:opacity-50"
          />
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-1.5 text-xs text-stone-500">
              <input
                type="checkbox"
                checked={spoiler}
                onChange={(e) => setSpoiler(e.target.checked)}
                disabled={submitting}
                className="rounded border-stone-300"
              />
              Contains spoilers
            </label>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !title.trim() || !body.trim()}
              className="text-xs px-3 py-1.5 rounded bg-stone-900 text-white hover:bg-stone-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
              }}
              disabled={submitting}
              className="text-xs text-stone-400 hover:text-stone-600 transition-colors"
            >
              Cancel
            </button>
            {error && <span className="text-xs text-red-500">{error}</span>}
          </div>
        </form>
      )}

      {/* Thread list */}
      {threads.length === 0 ? (
        <p className="text-stone-400 text-sm">No discussions yet.</p>
      ) : (
        <div className="space-y-4">
          {threads.map((thread) => (
            <Link
              key={thread.id}
              href={`/books/${workId}/threads/${thread.id}`}
              className="block border border-stone-100 rounded-lg p-4 hover:border-stone-300 transition-colors"
            >
              <div className="flex items-start gap-3">
                {thread.avatar_url ? (
                  <img
                    src={thread.avatar_url}
                    alt={thread.display_name ?? thread.username}
                    className="w-6 h-6 rounded-full object-cover shrink-0 mt-0.5"
                  />
                ) : (
                  <div className="w-6 h-6 rounded-full bg-stone-200 shrink-0 mt-0.5" />
                )}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-medium text-stone-900 truncate">
                      {thread.spoiler && (
                        <span className="text-[10px] font-medium text-amber-600 border border-amber-200 rounded px-1 py-0.5 mr-2 leading-none">
                          Spoiler
                        </span>
                      )}
                      {thread.title}
                    </h3>
                  </div>
                  <div className="flex items-center gap-2 mt-1 text-xs text-stone-400">
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
