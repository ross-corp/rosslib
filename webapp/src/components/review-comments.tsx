"use client";

import { useState } from "react";
import Link from "next/link";

type ReviewComment = {
  id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  body: string;
  created_at: string;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function ReviewComments({
  workId,
  reviewUserId,
  initialCount,
  isLoggedIn,
  currentUserId,
}: {
  workId: string;
  reviewUserId: string;
  initialCount: number;
  isLoggedIn: boolean;
  currentUserId?: string;
}) {
  const [expanded, setExpanded] = useState(false);
  const [comments, setComments] = useState<ReviewComment[]>([]);
  const [count, setCount] = useState(initialCount);
  const [loading, setLoading] = useState(false);
  const [body, setBody] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function loadComments() {
    if (loading) return;
    setLoading(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/reviews/${reviewUserId}/comments`
      );
      if (res.ok) {
        const data = await res.json();
        setComments(data);
        setCount(data.length);
      }
    } finally {
      setLoading(false);
    }
  }

  async function toggleExpanded() {
    if (!expanded) {
      await loadComments();
    }
    setExpanded(!expanded);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!body.trim() || submitting) return;
    setSubmitting(true);
    try {
      const res = await fetch(
        `/api/books/${workId}/reviews/${reviewUserId}/comments`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ body: body.trim() }),
        }
      );
      if (res.ok) {
        setBody("");
        await loadComments();
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(commentId: string) {
    const res = await fetch(`/api/review-comments/${commentId}`, {
      method: "DELETE",
    });
    if (res.ok || res.status === 204) {
      setComments((prev) => prev.filter((c) => c.id !== commentId));
      setCount((prev) => Math.max(0, prev - 1));
    }
  }

  return (
    <div className="mt-1">
      <button
        onClick={toggleExpanded}
        className="inline-flex items-center gap-1 text-xs text-text-tertiary hover:text-text-primary transition-colors"
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        </svg>
        {count > 0
          ? `${count} ${count === 1 ? "comment" : "comments"}`
          : "Comment"}
      </button>

      {expanded && (
        <div className="mt-3 ml-0 pl-4 border-l-2 border-border">
          {loading && comments.length === 0 ? (
            <p className="text-xs text-text-tertiary">Loading...</p>
          ) : (
            <>
              {comments.length > 0 && (
                <div className="space-y-3 mb-3">
                  {comments.map((comment) => (
                    <div key={comment.id} className="flex gap-2">
                      <Link
                        href={`/${comment.username}`}
                        className="shrink-0"
                      >
                        {comment.avatar_url ? (
                          <img
                            src={comment.avatar_url}
                            alt={comment.display_name ?? comment.username}
                            className="w-6 h-6 rounded-full object-cover"
                          />
                        ) : (
                          <div className="w-6 h-6 rounded-full bg-surface-2" />
                        )}
                      </Link>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <Link
                            href={`/${comment.username}`}
                            className="text-xs font-medium text-text-primary hover:underline"
                          >
                            {comment.display_name ?? comment.username}
                          </Link>
                          <span className="text-[10px] text-text-tertiary">
                            {formatDate(comment.created_at)}
                          </span>
                          {currentUserId === comment.user_id && (
                            <button
                              onClick={() => handleDelete(comment.id)}
                              className="text-[10px] text-text-tertiary hover:text-red-500 transition-colors"
                              title="Delete comment"
                            >
                              Delete
                            </button>
                          )}
                        </div>
                        <p className="text-xs text-text-primary leading-relaxed mt-0.5 whitespace-pre-wrap">
                          {comment.body}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {isLoggedIn && (
                <form onSubmit={handleSubmit} className="flex gap-2">
                  <input
                    type="text"
                    value={body}
                    onChange={(e) => setBody(e.target.value)}
                    placeholder="Add a comment..."
                    maxLength={2000}
                    className="flex-1 text-xs px-3 py-1.5 rounded bg-surface-2 border border-border text-text-primary placeholder:text-text-tertiary focus:outline-none focus:border-text-tertiary"
                  />
                  <button
                    type="submit"
                    disabled={!body.trim() || submitting}
                    className="text-xs px-3 py-1.5 rounded bg-surface-2 border border-border text-text-primary hover:bg-border transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {submitting ? "..." : "Post"}
                  </button>
                </form>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
