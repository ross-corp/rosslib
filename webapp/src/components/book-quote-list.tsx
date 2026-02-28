"use client";

import { useState } from "react";
import Link from "next/link";

type Quote = {
  id: string;
  user_id?: string;
  username?: string;
  display_name?: string | null;
  avatar_url?: string | null;
  text: string;
  page_number: number | null;
  note: string | null;
  is_public?: boolean;
  created_at: string;
};

type Props = {
  workId: string;
  initialQuotes: Quote[];
  myQuotes: Quote[];
  isLoggedIn: boolean;
  hasStatus: boolean;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function BookQuoteList({
  workId,
  initialQuotes,
  myQuotes: initialMyQuotes,
  isLoggedIn,
  hasStatus,
}: Props) {
  const [quotes, setQuotes] = useState<Quote[]>(initialQuotes);
  const [myQuotes, setMyQuotes] = useState<Quote[]>(initialMyQuotes);
  const [showForm, setShowForm] = useState(false);
  const [text, setText] = useState("");
  const [pageNumber, setPageNumber] = useState("");
  const [note, setNote] = useState("");
  const [isPublic, setIsPublic] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim()) return;

    setSubmitting(true);
    setError(null);

    const body: Record<string, unknown> = {
      text: text.trim(),
      is_public: isPublic,
    };
    if (pageNumber.trim()) {
      const num = parseInt(pageNumber.trim(), 10);
      if (!isNaN(num) && num > 0) {
        body.page_number = num;
      }
    }
    if (note.trim()) {
      body.note = note.trim();
    }

    const res = await fetch(`/api/me/books/${workId}/quotes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json();
      setError(data.error || "Failed to add quote");
      return;
    }

    // Refetch both lists
    const [pubRes, myRes] = await Promise.all([
      fetch(`/api/books/${workId}/quotes`),
      fetch(`/api/me/books/${workId}/quotes`),
    ]);
    if (pubRes.ok) setQuotes(await pubRes.json());
    if (myRes.ok) setMyQuotes(await myRes.json());

    setText("");
    setPageNumber("");
    setNote("");
    setIsPublic(true);
    setShowForm(false);
  }

  async function handleDelete(quoteId: string) {
    const res = await fetch(`/api/me/quotes/${quoteId}`, {
      method: "DELETE",
    });
    if (res.ok || res.status === 204) {
      setMyQuotes((prev) => prev.filter((q) => q.id !== quoteId));
      setQuotes((prev) => prev.filter((q) => q.id !== quoteId));
    }
  }

  // Deduplicate: don't show my public quotes twice
  const myQuoteIds = new Set(myQuotes.map((q) => q.id));
  const communityQuotes = quotes.filter((q) => !myQuoteIds.has(q.id));

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
          {quotes.length > 0 || myQuotes.length > 0
            ? `Quotes (${quotes.length + myQuotes.filter((q) => !q.is_public).length})`
            : "Quotes"}
        </h2>
        {isLoggedIn && hasStatus && !showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover transition-colors"
          >
            Add quote
          </button>
        )}
      </div>

      {/* Add quote form */}
      {showForm && (
        <form onSubmit={handleSubmit} className="mb-8 space-y-3">
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Enter a quote from the book..."
            rows={3}
            maxLength={2000}
            disabled={submitting}
            className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
          />
          <div className="flex gap-3">
            <input
              type="number"
              value={pageNumber}
              onChange={(e) => setPageNumber(e.target.value)}
              placeholder="Page #"
              min={1}
              disabled={submitting}
              className="w-24 border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
            />
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Note (optional)"
              maxLength={500}
              disabled={submitting}
              className="flex-1 border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
            />
          </div>
          <label className="flex items-center gap-1.5 text-xs text-text-primary">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              disabled={submitting}
              className="rounded border-border"
            />
            Visible to others
          </label>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !text.trim()}
              className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "Saving..." : "Save quote"}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowForm(false);
                setText("");
                setPageNumber("");
                setNote("");
                setIsPublic(true);
                setError(null);
              }}
              disabled={submitting}
              className="text-xs text-text-primary hover:text-text-primary transition-colors"
            >
              Cancel
            </button>
            {error && <span className="text-xs text-red-500">{error}</span>}
          </div>
        </form>
      )}

      {/* My quotes (private ones only visible to me) */}
      {myQuotes.length > 0 && (
        <div className="space-y-4 mb-6">
          {myQuotes.map((quote) => (
            <div
              key={quote.id}
              className="border border-border rounded-lg p-4"
            >
              <blockquote className="text-sm text-text-primary italic leading-relaxed">
                &ldquo;{quote.text}&rdquo;
              </blockquote>
              <div className="flex items-center gap-2 mt-2 text-xs text-text-tertiary">
                <span>You</span>
                {quote.page_number != null && (
                  <>
                    <span>&middot;</span>
                    <span>p. {quote.page_number}</span>
                  </>
                )}
                <span>&middot;</span>
                <span>{formatDate(quote.created_at)}</span>
                {!quote.is_public && (
                  <>
                    <span>&middot;</span>
                    <span className="text-[10px] font-medium border border-border rounded px-1.5 py-0.5 leading-none">
                      Private
                    </span>
                  </>
                )}
                <span>&middot;</span>
                <button
                  type="button"
                  onClick={() => handleDelete(quote.id)}
                  className="text-red-500 hover:text-red-700 transition-colors"
                >
                  Delete
                </button>
              </div>
              {quote.note && (
                <p className="text-xs text-text-tertiary mt-1">{quote.note}</p>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Community quotes */}
      {communityQuotes.length === 0 && myQuotes.length === 0 ? (
        <p className="text-text-primary text-sm">No quotes yet.</p>
      ) : (
        <div className="space-y-4">
          {communityQuotes.map((quote) => (
            <div
              key={quote.id}
              className="border border-border rounded-lg p-4"
            >
              <blockquote className="text-sm text-text-primary italic leading-relaxed">
                &ldquo;{quote.text}&rdquo;
              </blockquote>
              <div className="flex items-center gap-2 mt-2 text-xs text-text-tertiary">
                <Link
                  href={`/${quote.username}`}
                  className="flex items-center gap-1.5 hover:text-text-primary transition-colors"
                >
                  {quote.avatar_url ? (
                    <img
                      src={quote.avatar_url}
                      alt={quote.display_name ?? quote.username ?? ""}
                      className="w-4 h-4 rounded-full object-cover"
                    />
                  ) : (
                    <div className="w-4 h-4 rounded-full bg-surface-2" />
                  )}
                  <span>{quote.display_name ?? quote.username}</span>
                </Link>
                {quote.page_number != null && (
                  <>
                    <span>&middot;</span>
                    <span>p. {quote.page_number}</span>
                  </>
                )}
                <span>&middot;</span>
                <span>{formatDate(quote.created_at)}</span>
              </div>
              {quote.note && (
                <p className="text-xs text-text-tertiary mt-1">{quote.note}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
