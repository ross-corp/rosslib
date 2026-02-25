"use client";

import { useState } from "react";
import Link from "next/link";

type BookLink = {
  id: string;
  from_book_ol_id: string;
  to_book_ol_id: string;
  to_book_title: string;
  to_book_authors: string | null;
  to_book_cover_url: string | null;
  link_type: string;
  note: string | null;
  username: string;
  display_name: string | null;
  votes: number;
  user_voted: boolean;
  created_at: string;
};

type Props = {
  workId: string;
  initialLinks: BookLink[];
  isLoggedIn: boolean;
  currentUsername?: string;
  isModerator?: boolean;
};

const LINK_TYPES = [
  { value: "sequel", label: "Sequel" },
  { value: "prequel", label: "Prequel" },
  { value: "companion", label: "Companion" },
  { value: "similar", label: "Similar" },
  { value: "mentioned_in", label: "Mentioned in" },
  { value: "adaptation", label: "Adaptation" },
];

function linkTypeLabel(type: string): string {
  return LINK_TYPES.find((t) => t.value === type)?.label ?? type;
}

export default function BookLinkList({ workId, initialLinks, isLoggedIn, currentUsername, isModerator }: Props) {
  const [links, setLinks] = useState<BookLink[]>(initialLinks);
  const [showForm, setShowForm] = useState(false);
  const [toWorkId, setToWorkId] = useState("");
  const [linkType, setLinkType] = useState("similar");
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [votingIds, setVotingIds] = useState<Set<string>>(new Set());
  const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());

  async function refetchLinks() {
    const res = await fetch(`/api/books/${workId}/links`);
    if (res.ok) setLinks(await res.json());
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!toWorkId.trim() || !linkType) return;

    setSubmitting(true);
    setError(null);

    const res = await fetch(`/api/books/${workId}/links`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        to_work_id: toWorkId.trim(),
        link_type: linkType,
        note: note.trim() || null,
      }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => null);
      setError(data?.error ?? "Failed to add link");
      return;
    }

    await refetchLinks();
    setToWorkId("");
    setNote("");
    setLinkType("similar");
    setShowForm(false);
  }

  async function handleVote(linkId: string, currentlyVoted: boolean) {
    if (votingIds.has(linkId)) return;
    setVotingIds((prev) => new Set(prev).add(linkId));

    await fetch(`/api/links/${linkId}/vote`, {
      method: currentlyVoted ? "DELETE" : "POST",
    });

    await refetchLinks();
    setVotingIds((prev) => {
      const next = new Set(prev);
      next.delete(linkId);
      return next;
    });
  }

  async function handleDelete(linkId: string) {
    if (deletingIds.has(linkId)) return;
    setDeletingIds((prev) => new Set(prev).add(linkId));

    const res = await fetch(`/api/links/${linkId}`, { method: "DELETE" });
    if (res.ok || res.status === 204) {
      setLinks((prev) => prev.filter((l) => l.id !== linkId));
    }

    setDeletingIds((prev) => {
      const next = new Set(prev);
      next.delete(linkId);
      return next;
    });
  }

  // Group links by type.
  const grouped = links.reduce<Record<string, BookLink[]>>((acc, link) => {
    (acc[link.link_type] ??= []).push(link);
    return acc;
  }, {});

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider">
          {links.length > 0 ? `Related Books (${links.length})` : "Related Books"}
        </h2>
        {isLoggedIn && !showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded bg-stone-900 text-white hover:bg-stone-700 transition-colors"
          >
            Suggest link
          </button>
        )}
      </div>

      {/* Add link form */}
      {showForm && (
        <form onSubmit={handleSubmit} className="mb-8 space-y-3">
          <div>
            <label className="block text-xs text-stone-500 mb-1">
              Target book ID (Open Library work ID, e.g. OL82592W)
            </label>
            <input
              type="text"
              value={toWorkId}
              onChange={(e) => setToWorkId(e.target.value)}
              placeholder="OL82592W"
              disabled={submitting}
              className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            />
          </div>
          <div>
            <label className="block text-xs text-stone-500 mb-1">
              Relationship type
            </label>
            <select
              value={linkType}
              onChange={(e) => setLinkType(e.target.value)}
              disabled={submitting}
              className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            >
              {LINK_TYPES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-stone-500 mb-1">
              Note (optional)
            </label>
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Explain the connection..."
              disabled={submitting}
              className="w-full border border-stone-200 rounded px-3 py-2 text-sm text-stone-700 placeholder:text-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-400 disabled:opacity-50"
            />
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !toWorkId.trim()}
              className="text-xs px-3 py-1.5 rounded bg-stone-900 text-white hover:bg-stone-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "Adding..." : "Add link"}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowForm(false);
                setToWorkId("");
                setNote("");
                setLinkType("similar");
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

      {/* Link list grouped by type */}
      {links.length === 0 ? (
        <p className="text-stone-400 text-sm">
          No related books yet.{isLoggedIn && " Be the first to suggest a connection."}
        </p>
      ) : (
        <div className="space-y-6">
          {Object.entries(grouped).map(([type, typeLinks]) => (
            <div key={type}>
              <h3 className="text-xs font-medium text-stone-400 uppercase tracking-wider mb-3">
                {linkTypeLabel(type)}
              </h3>
              <div className="space-y-3">
                {typeLinks.map((link) => (
                  <div
                    key={link.id}
                    className="flex items-start gap-3 border border-stone-100 rounded-lg p-3"
                  >
                    {/* Cover thumbnail */}
                    <Link
                      href={`/books/${link.to_book_ol_id}`}
                      className="shrink-0"
                    >
                      {link.to_book_cover_url ? (
                        <img
                          src={link.to_book_cover_url}
                          alt={link.to_book_title}
                          className="w-10 h-14 rounded object-cover"
                        />
                      ) : (
                        <div className="w-10 h-14 rounded bg-stone-100" />
                      )}
                    </Link>

                    {/* Info */}
                    <div className="flex-1 min-w-0">
                      <Link
                        href={`/books/${link.to_book_ol_id}`}
                        className="text-sm font-medium text-stone-900 hover:underline line-clamp-1"
                      >
                        {link.to_book_title}
                      </Link>
                      {link.to_book_authors && (
                        <p className="text-xs text-stone-400 line-clamp-1">
                          {link.to_book_authors}
                        </p>
                      )}
                      {link.note && (
                        <p className="text-xs text-stone-500 mt-1 line-clamp-2">
                          {link.note}
                        </p>
                      )}
                      <p className="text-[10px] text-stone-400 mt-1">
                        by {link.display_name ?? link.username}
                      </p>
                    </div>

                    {/* Actions: vote + delete */}
                    <div className="shrink-0 flex items-center gap-1">
                      {isLoggedIn && (
                        <button
                          type="button"
                          onClick={() => handleVote(link.id, link.user_voted)}
                          disabled={votingIds.has(link.id)}
                          className={`flex flex-col items-center gap-0.5 px-2 py-1 rounded text-xs transition-colors ${
                            link.user_voted
                              ? "text-stone-900 bg-stone-100"
                              : "text-stone-400 hover:text-stone-600 hover:bg-stone-50"
                          } disabled:opacity-50`}
                          title={link.user_voted ? "Remove upvote" : "Upvote"}
                        >
                          <svg
                            viewBox="0 0 12 12"
                            className="w-3 h-3"
                            fill={link.user_voted ? "currentColor" : "none"}
                            stroke="currentColor"
                            strokeWidth={1.5}
                          >
                            <path d="M6 2L10 8H2L6 2Z" />
                          </svg>
                          <span>{link.votes}</span>
                        </button>
                      )}
                      {!isLoggedIn && link.votes > 0 && (
                        <span className="text-xs text-stone-400 px-2 py-1">
                          {link.votes}
                        </span>
                      )}
                      {isLoggedIn && (currentUsername === link.username || isModerator) && (
                        <button
                          type="button"
                          onClick={() => handleDelete(link.id)}
                          disabled={deletingIds.has(link.id)}
                          className="px-1.5 py-1 rounded text-stone-300 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                          title={currentUsername === link.username ? "Delete your link" : "Remove link (moderator)"}
                        >
                          <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={1.5}>
                            <path d="M2 3h8M4.5 3V2h3v1M3 3v7h6V3M5 5v3M7 5v3" />
                          </svg>
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
