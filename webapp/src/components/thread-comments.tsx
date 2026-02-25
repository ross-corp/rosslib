"use client";

import { useState } from "react";
import Link from "next/link";

type Comment = {
  id: string;
  thread_id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  parent_id: string | null;
  body: string;
  created_at: string;
};

type Props = {
  threadId: string;
  initialComments: Comment[];
  isLoggedIn: boolean;
  currentUserId: string | null;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function CommentItem({
  comment,
  replies,
  threadId,
  isLoggedIn,
  currentUserId,
  onReply,
  onDelete,
}: {
  comment: Comment;
  replies: Comment[];
  threadId: string;
  isLoggedIn: boolean;
  currentUserId: string | null;
  onReply: (parentId: string, body: string) => Promise<void>;
  onDelete: (commentId: string) => Promise<void>;
}) {
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [replyBody, setReplyBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const isOwner = currentUserId === comment.user_id;

  async function handleReply(e: React.FormEvent) {
    e.preventDefault();
    if (!replyBody.trim()) return;
    setSubmitting(true);
    await onReply(comment.id, replyBody.trim());
    setReplyBody("");
    setShowReplyForm(false);
    setSubmitting(false);
  }

  return (
    <div>
      <div className="flex gap-3">
        <Link href={`/${comment.username}`} className="shrink-0">
          {comment.avatar_url ? (
            <img
              src={comment.avatar_url}
              alt={comment.display_name ?? comment.username}
              className="w-7 h-7 rounded-full object-cover"
            />
          ) : (
            <div className="w-7 h-7 rounded-full bg-surface-2" />
          )}
        </Link>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Link
              href={`/${comment.username}`}
              className="text-sm font-medium text-text-primary hover:underline"
            >
              {comment.display_name ?? comment.username}
            </Link>
            <span className="text-xs text-text-primary">
              {formatDate(comment.created_at)}
            </span>
          </div>
          <p className="text-sm text-text-primary leading-relaxed whitespace-pre-wrap mt-1">
            {comment.body}
          </p>
          <div className="flex items-center gap-3 mt-1.5">
            {isLoggedIn && !comment.parent_id && (
              <button
                type="button"
                onClick={() => setShowReplyForm(!showReplyForm)}
                className="text-xs text-text-primary hover:text-text-primary transition-colors"
              >
                Reply
              </button>
            )}
            {isOwner && (
              <button
                type="button"
                onClick={() => onDelete(comment.id)}
                className="text-xs text-red-400 hover:text-red-600 transition-colors"
              >
                Delete
              </button>
            )}
          </div>

          {/* Reply form */}
          {showReplyForm && (
            <form onSubmit={handleReply} className="mt-3 space-y-2">
              <textarea
                value={replyBody}
                onChange={(e) => setReplyBody(e.target.value)}
                placeholder="Write a reply..."
                rows={2}
                disabled={submitting}
                className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
              />
              <div className="flex items-center gap-3">
                <button
                  type="submit"
                  disabled={submitting || !replyBody.trim()}
                  className="text-xs px-3 py-1.5 rounded bg-accent text-white hover:bg-surface-3 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {submitting ? "Posting..." : "Reply"}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowReplyForm(false);
                    setReplyBody("");
                  }}
                  disabled={submitting}
                  className="text-xs text-text-primary hover:text-text-primary transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          )}
        </div>
      </div>

      {/* Nested replies */}
      {replies.length > 0 && (
        <div className="ml-10 mt-4 space-y-4 border-l border-border pl-4">
          {replies.map((reply) => (
            <div key={reply.id} className="flex gap-3">
              <Link href={`/${reply.username}`} className="shrink-0">
                {reply.avatar_url ? (
                  <img
                    src={reply.avatar_url}
                    alt={reply.display_name ?? reply.username}
                    className="w-6 h-6 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-6 h-6 rounded-full bg-surface-2" />
                )}
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Link
                    href={`/${reply.username}`}
                    className="text-sm font-medium text-text-primary hover:underline"
                  >
                    {reply.display_name ?? reply.username}
                  </Link>
                  <span className="text-xs text-text-primary">
                    {formatDate(reply.created_at)}
                  </span>
                </div>
                <p className="text-sm text-text-primary leading-relaxed whitespace-pre-wrap mt-1">
                  {reply.body}
                </p>
                {currentUserId === reply.user_id && (
                  <button
                    type="button"
                    onClick={() => onDelete(reply.id)}
                    className="text-xs text-red-400 hover:text-red-600 transition-colors mt-1"
                  >
                    Delete
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function ThreadComments({
  threadId,
  initialComments,
  isLoggedIn,
  currentUserId,
}: Props) {
  const [comments, setComments] = useState<Comment[]>(initialComments);
  const [newComment, setNewComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Group top-level comments and their replies.
  const topLevel = comments.filter((c) => !c.parent_id);
  const repliesByParent = new Map<string, Comment[]>();
  for (const c of comments) {
    if (c.parent_id) {
      const arr = repliesByParent.get(c.parent_id) ?? [];
      arr.push(c);
      repliesByParent.set(c.parent_id, arr);
    }
  }

  async function refreshComments() {
    const res = await fetch(`/api/threads/${threadId}`);
    if (res.ok) {
      const data = await res.json();
      setComments(data.comments);
    }
  }

  async function handleSubmitComment(e: React.FormEvent) {
    e.preventDefault();
    if (!newComment.trim()) return;

    setSubmitting(true);
    setError(null);

    const res = await fetch(`/api/threads/${threadId}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ body: newComment.trim() }),
    });

    setSubmitting(false);

    if (!res.ok) {
      setError("Failed to post comment");
      return;
    }

    setNewComment("");
    await refreshComments();
  }

  async function handleReply(parentId: string, body: string) {
    const res = await fetch(`/api/threads/${threadId}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ body, parent_id: parentId }),
    });

    if (res.ok) {
      await refreshComments();
    }
  }

  async function handleDelete(commentId: string) {
    if (!confirm("Delete this comment?")) return;

    const res = await fetch(
      `/api/threads/${threadId}/comments/${commentId}`,
      { method: "DELETE" }
    );

    if (res.ok) {
      await refreshComments();
    }
  }

  return (
    <div>
      <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-6">
        {comments.length > 0
          ? `Comments (${comments.length})`
          : "Comments"}
      </h2>

      {/* New comment form */}
      {isLoggedIn && (
        <form onSubmit={handleSubmitComment} className="mb-8 space-y-3">
          <textarea
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder="Add a comment..."
            rows={3}
            disabled={submitting}
            className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
          />
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !newComment.trim()}
              className="text-xs px-3 py-1.5 rounded bg-accent text-white hover:bg-surface-3 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "Posting..." : "Comment"}
            </button>
            {error && <span className="text-xs text-red-500">{error}</span>}
          </div>
        </form>
      )}

      {/* Comment list */}
      {topLevel.length === 0 ? (
        <p className="text-text-primary text-sm">No comments yet.</p>
      ) : (
        <div className="space-y-6">
          {topLevel.map((comment) => (
            <CommentItem
              key={comment.id}
              comment={comment}
              replies={repliesByParent.get(comment.id) ?? []}
              threadId={threadId}
              isLoggedIn={isLoggedIn}
              currentUserId={currentUserId}
              onReply={handleReply}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
