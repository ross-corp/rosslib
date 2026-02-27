"use client";

import { useState, useEffect, useRef } from "react";
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

type SearchResult = {
  open_library_id: string;
  title: string;
  authors: string | null;
  cover_url: string | null;
  publication_year: number | null;
};

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
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editType, setEditType] = useState("");
  const [editNote, setEditNote] = useState("");
  const [editSubmitting, setEditSubmitting] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);
  const [editSuccess, setEditSuccess] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [selectedBook, setSelectedBook] = useState<SearchResult | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const searchDebounce = useRef<ReturnType<typeof setTimeout> | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (searchDebounce.current) clearTimeout(searchDebounce.current);

    if (searchQuery.trim().length < 2) {
      setSearchResults([]);
      setSearchLoading(false);
      return;
    }

    setSearchLoading(true);
    searchDebounce.current = setTimeout(async () => {
      try {
        const res = await fetch(`/api/books/search?q=${encodeURIComponent(searchQuery.trim())}`);
        if (res.ok) {
          const data = await res.json();
          setSearchResults(data.results ?? []);
        }
      } catch {
        // ignore network errors for suggestions
      } finally {
        setSearchLoading(false);
      }
    }, 400);

    return () => {
      if (searchDebounce.current) clearTimeout(searchDebounce.current);
    };
  }, [searchQuery]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  function selectBook(result: SearchResult) {
    setSelectedBook(result);
    setToWorkId(result.open_library_id);
    setSearchQuery("");
    setSearchResults([]);
    setShowDropdown(false);
  }

  function clearSelectedBook() {
    setSelectedBook(null);
    setToWorkId("");
    setSearchQuery("");
  }

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
    setSelectedBook(null);
    setSearchQuery("");
    setSearchResults([]);
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

  function openEditForm(link: BookLink) {
    setEditingId(link.id);
    setEditType(link.link_type);
    setEditNote(link.note ?? "");
    setEditError(null);
    setEditSuccess(null);
  }

  async function handleEditSubmit(e: React.FormEvent, link: BookLink) {
    e.preventDefault();
    setEditSubmitting(true);
    setEditError(null);
    setEditSuccess(null);

    const proposed: Record<string, string | null> = {};
    if (editType !== link.link_type) proposed.proposed_type = editType;
    const newNote = editNote.trim() || null;
    if (newNote !== (link.note ?? null)) proposed.proposed_note = newNote;

    if (Object.keys(proposed).length === 0) {
      setEditError("No changes to propose");
      setEditSubmitting(false);
      return;
    }

    const res = await fetch(`/api/links/${link.id}/edits`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(proposed),
    });

    setEditSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => null);
      setEditError(data?.error ?? "Failed to propose edit");
      return;
    }

    setEditSuccess("Edit proposed — a moderator will review it.");
    setTimeout(() => {
      setEditingId(null);
      setEditSuccess(null);
    }, 2000);
  }

  // Group links by type.
  const grouped = links.reduce<Record<string, BookLink[]>>((acc, link) => {
    (acc[link.link_type] ??= []).push(link);
    return acc;
  }, {});

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
          {links.length > 0 ? `Related Books (${links.length})` : "Related Books"}
        </h2>
        {isLoggedIn && !showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded bg-accent text-white hover:bg-surface-3 transition-colors"
          >
            Suggest link
          </button>
        )}
      </div>

      {/* Add link form */}
      {showForm && (
        <form onSubmit={handleSubmit} className="mb-8 space-y-3">
          <div ref={dropdownRef} className="relative">
            <label className="block text-xs text-text-primary mb-1">
              Target book
            </label>
            {selectedBook ? (
              <div className="flex items-center gap-2 border border-border rounded px-3 py-2">
                {selectedBook.cover_url ? (
                  <img src={selectedBook.cover_url} alt="" className="w-6 h-8 rounded object-cover shrink-0" />
                ) : (
                  <div className="w-6 h-8 rounded bg-surface-2 shrink-0" />
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-text-primary line-clamp-1">{selectedBook.title}</p>
                  {selectedBook.authors && (
                    <p className="text-xs text-text-primary line-clamp-1">{selectedBook.authors}</p>
                  )}
                </div>
                <button
                  type="button"
                  onClick={clearSelectedBook}
                  disabled={submitting}
                  className="text-xs text-text-primary hover:text-red-500 transition-colors disabled:opacity-50"
                >
                  <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={1.5}>
                    <path d="M3 3l6 6M9 3l-6 6" />
                  </svg>
                </button>
              </div>
            ) : (
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setShowDropdown(true);
                }}
                onFocus={() => { if (searchResults.length > 0) setShowDropdown(true); }}
                placeholder="Search for a book..."
                disabled={submitting}
                className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
              />
            )}
            {showDropdown && !selectedBook && searchQuery.trim().length >= 2 && (
              <div className="absolute z-10 top-full left-0 right-0 mt-1 bg-surface-0 border border-border rounded-lg shadow-lg max-h-60 overflow-y-auto">
                {searchLoading && searchResults.length === 0 && (
                  <p className="px-3 py-2 text-xs text-text-primary">Searching...</p>
                )}
                {!searchLoading && searchResults.length === 0 && (
                  <p className="px-3 py-2 text-xs text-text-primary">No results found</p>
                )}
                {searchResults.map((result) => (
                  <button
                    key={result.open_library_id}
                    type="button"
                    onClick={() => selectBook(result)}
                    className="w-full flex items-center gap-2 px-3 py-2 hover:bg-surface-2 transition-colors text-left"
                  >
                    {result.cover_url ? (
                      <img src={result.cover_url} alt="" className="w-6 h-8 rounded object-cover shrink-0" />
                    ) : (
                      <div className="w-6 h-8 rounded bg-surface-2 shrink-0" />
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-text-primary line-clamp-1">{result.title}</p>
                      <p className="text-xs text-text-primary line-clamp-1">
                        {result.authors ?? "Unknown author"}
                        {result.publication_year ? ` (${result.publication_year})` : ""}
                      </p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
          <div>
            <label className="block text-xs text-text-primary mb-1">
              Relationship type
            </label>
            <select
              value={linkType}
              onChange={(e) => setLinkType(e.target.value)}
              disabled={submitting}
              className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
            >
              {LINK_TYPES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-text-primary mb-1">
              Note (optional)
            </label>
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Explain the connection..."
              disabled={submitting}
              className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
            />
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={submitting || !toWorkId.trim()}
              className="text-xs px-3 py-1.5 rounded bg-accent text-white hover:bg-surface-3 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
                setSelectedBook(null);
                setSearchQuery("");
                setSearchResults([]);
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

      {/* Link list grouped by type */}
      {links.length === 0 ? (
        <p className="text-text-primary text-sm">
          No related books yet.{isLoggedIn && " Be the first to suggest a connection."}
        </p>
      ) : (
        <div className="space-y-6">
          {Object.entries(grouped).map(([type, typeLinks]) => (
            <div key={type}>
              <h3 className="text-xs font-medium text-text-primary uppercase tracking-wider mb-3">
                {linkTypeLabel(type)}
              </h3>
              <div className="space-y-3">
                {typeLinks.map((link) => (
                  <div key={link.id}>
                    <div className="flex items-start gap-3 border border-border rounded-lg p-3">
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
                          <div className="w-10 h-14 rounded bg-surface-2" />
                        )}
                      </Link>

                      {/* Info */}
                      <div className="flex-1 min-w-0">
                        <Link
                          href={`/books/${link.to_book_ol_id}`}
                          className="text-sm font-medium text-text-primary hover:underline line-clamp-1"
                        >
                          {link.to_book_title}
                        </Link>
                        {link.to_book_authors && (
                          <p className="text-xs text-text-primary line-clamp-1">
                            {link.to_book_authors}
                          </p>
                        )}
                        {link.note && (
                          <p className="text-xs text-text-primary mt-1 line-clamp-2">
                            {link.note}
                          </p>
                        )}
                        <p className="text-[10px] text-text-primary mt-1">
                          by {link.display_name ?? link.username}
                        </p>
                      </div>

                      {/* Actions: vote + edit + delete */}
                      <div className="shrink-0 flex items-center gap-1">
                        {isLoggedIn && (
                          <button
                            type="button"
                            onClick={() => handleVote(link.id, link.user_voted)}
                            disabled={votingIds.has(link.id)}
                            className={`flex flex-col items-center gap-0.5 px-2 py-1 rounded text-xs transition-colors ${
                              link.user_voted
                                ? "text-text-primary bg-surface-2"
                                : "text-text-primary hover:text-text-primary hover:bg-surface-2"
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
                          <span className="text-xs text-text-primary px-2 py-1">
                            {link.votes}
                          </span>
                        )}
                        {isLoggedIn && (
                          <button
                            type="button"
                            onClick={() =>
                              editingId === link.id
                                ? setEditingId(null)
                                : openEditForm(link)
                            }
                            className="px-1.5 py-1 rounded text-text-primary hover:text-text-primary hover:bg-surface-2 transition-colors"
                            title="Suggest an edit"
                          >
                            <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={1.5}>
                              <path d="M8.5 1.5l2 2L4 10H2v-2L8.5 1.5Z" />
                            </svg>
                          </button>
                        )}
                        {isLoggedIn && (currentUsername === link.username || isModerator) && (
                          <button
                            type="button"
                            onClick={() => handleDelete(link.id)}
                            disabled={deletingIds.has(link.id)}
                            className="px-1.5 py-1 rounded text-text-primary hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                            title={currentUsername === link.username ? "Delete your link" : "Remove link (moderator)"}
                          >
                            <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={1.5}>
                              <path d="M2 3h8M4.5 3V2h3v1M3 3v7h6V3M5 5v3M7 5v3" />
                            </svg>
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Inline edit proposal form */}
                    {editingId === link.id && (
                      <form
                        onSubmit={(e) => handleEditSubmit(e, link)}
                        className="mt-2 ml-13 border border-border rounded-lg p-3 bg-surface-2 space-y-2"
                      >
                        <p className="text-xs font-medium text-text-primary">Suggest an edit</p>
                        <div>
                          <label className="block text-xs text-text-primary mb-1">
                            Relationship type
                          </label>
                          <select
                            value={editType}
                            onChange={(e) => setEditType(e.target.value)}
                            disabled={editSubmitting}
                            className="w-full border border-border rounded px-2 py-1.5 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50 bg-surface-0"
                          >
                            {LINK_TYPES.map((t) => (
                              <option key={t.value} value={t.value}>
                                {t.label}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div>
                          <label className="block text-xs text-text-primary mb-1">Note</label>
                          <input
                            type="text"
                            value={editNote}
                            onChange={(e) => setEditNote(e.target.value)}
                            placeholder="Explain the connection..."
                            disabled={editSubmitting}
                            className="w-full border border-border rounded px-2 py-1.5 text-xs text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50 bg-surface-0"
                          />
                        </div>
                        <div className="flex items-center gap-3">
                          <button
                            type="submit"
                            disabled={editSubmitting}
                            className="text-xs px-3 py-1 rounded bg-accent text-white hover:bg-surface-3 disabled:opacity-50 transition-colors"
                          >
                            {editSubmitting ? "Submitting..." : "Propose edit"}
                          </button>
                          <button
                            type="button"
                            onClick={() => setEditingId(null)}
                            disabled={editSubmitting}
                            className="text-xs text-text-primary hover:text-text-primary transition-colors"
                          >
                            Cancel
                          </button>
                          {editError && <span className="text-xs text-red-500">{editError}</span>}
                          {editSuccess && <span className="text-xs text-green-600">{editSuccess}</span>}
                        </div>
                      </form>
                    )}
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
