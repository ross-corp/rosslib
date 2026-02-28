"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import BookCoverPlaceholder from "@/components/book-cover-placeholder";

type PendingImport = {
  id: string;
  title: string;
  author: string;
  isbn13: string;
  exclusive_shelf: string;
  custom_shelves: string[];
  rating: number | null;
  review_text: string;
  date_read: string;
  date_added: string;
  created: string;
};

type SearchResult = {
  open_library_id: string;
  title: string;
  authors: string;
  cover_url: string | null;
  publication_year: number | null;
  isbn13: string | null;
};

function shelfLabel(slug: string): string {
  const labels: Record<string, string> = {
    read: "Read",
    "to-read": "Want to Read",
    "currently-reading": "Currently Reading",
    "want-to-read": "Want to Read",
    "owned-to-read": "Owned to Read",
    dnf: "Did Not Finish",
  };
  return labels[slug] ?? slug;
}

export default function PendingImportsManager({
  initialItems,
}: {
  initialItems: PendingImport[];
}) {
  const [items, setItems] = useState<PendingImport[]>(initialItems);
  const [searchModalId, setSearchModalId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [resolving, setResolving] = useState<string | null>(null);
  const [dismissing, setDismissing] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const activeItem = items.find((i) => i.id === searchModalId);

  // Focus search input when modal opens
  useEffect(() => {
    if (searchModalId && inputRef.current) {
      inputRef.current.focus();
    }
  }, [searchModalId]);

  const doSearch = useCallback(async (q: string) => {
    if (!q.trim()) {
      setSearchResults([]);
      return;
    }
    setSearching(true);
    try {
      const res = await fetch(`/api/books/search?q=${encodeURIComponent(q)}`);
      if (res.ok) {
        const data = await res.json();
        setSearchResults(data.results ?? []);
      }
    } catch {
      // ignore
    } finally {
      setSearching(false);
    }
  }, []);

  function handleSearchInput(value: string) {
    setSearchQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(value), 400);
  }

  function openSearchModal(id: string) {
    const item = items.find((i) => i.id === id);
    setSearchModalId(id);
    setSearchResults([]);
    const q = item ? `${item.title} ${item.author}`.trim() : "";
    setSearchQuery(q);
    if (q) doSearch(q);
  }

  function closeSearchModal() {
    setSearchModalId(null);
    setSearchQuery("");
    setSearchResults([]);
  }

  async function resolveImport(importId: string, olId: string) {
    setResolving(importId);
    try {
      const res = await fetch(`/api/me/imports/pending/${importId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "resolve", ol_id: olId }),
      });
      if (res.ok) {
        setItems((prev) => prev.filter((i) => i.id !== importId));
        closeSearchModal();
      }
    } catch {
      // ignore
    } finally {
      setResolving(null);
    }
  }

  async function dismissImport(id: string) {
    setDismissing(id);
    try {
      await fetch(`/api/me/imports/pending/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "dismiss" }),
      });
      setItems((prev) => prev.filter((i) => i.id !== id));
    } catch {
      // ignore
    } finally {
      setDismissing(null);
    }
  }

  async function deleteImport(id: string) {
    setDismissing(id);
    try {
      await fetch(`/api/me/imports/pending/${id}`, { method: "DELETE" });
      setItems((prev) => prev.filter((i) => i.id !== id));
    } catch {
      // ignore
    } finally {
      setDismissing(null);
    }
  }

  if (items.length === 0) {
    return (
      <div className="text-center py-16">
        <p className="text-text-primary text-sm">No pending imports.</p>
        <p className="text-text-primary text-xs mt-1">
          All your imported books have been matched or dismissed.
        </p>
      </div>
    );
  }

  return (
    <>
      <p className="text-sm text-text-primary mb-4">
        {items.length} unmatched book{items.length !== 1 ? "s" : ""} from previous imports.
        Search for the correct book to link them, or dismiss rows you don&apos;t need.
      </p>

      <div className="divide-y divide-border">
        {items.map((item) => (
          <div key={item.id} className="py-4 flex items-start gap-4">
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-text-primary truncate">
                {item.title}
              </p>
              <p className="text-xs text-text-primary mt-0.5">
                {item.author}
                {item.isbn13 && (
                  <span className="ml-2 text-text-primary">ISBN: {item.isbn13}</span>
                )}
              </p>
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                {item.exclusive_shelf && (
                  <span className="inline-block text-xs px-2 py-0.5 rounded-full bg-surface-2 text-text-primary">
                    {shelfLabel(item.exclusive_shelf)}
                  </span>
                )}
                {item.rating != null && item.rating > 0 && (
                  <span className="text-xs text-text-primary">
                    {"★".repeat(item.rating)}
                  </span>
                )}
                {item.date_read && (
                  <span className="text-xs text-text-primary">
                    Read {item.date_read}
                  </span>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2 shrink-0">
              <button
                type="button"
                onClick={() => openSearchModal(item.id)}
                className="px-3 py-1.5 text-xs font-medium border border-accent text-accent rounded-lg hover:bg-accent hover:text-white transition-colors"
              >
                Search &amp; Link
              </button>
              <button
                type="button"
                onClick={() => dismissImport(item.id)}
                disabled={dismissing === item.id}
                className="px-3 py-1.5 text-xs text-text-primary border border-border rounded-lg hover:bg-surface-2 transition-colors disabled:opacity-40"
              >
                Dismiss
              </button>
              <button
                type="button"
                onClick={() => deleteImport(item.id)}
                disabled={dismissing === item.id}
                className="px-3 py-1.5 text-xs text-red-600 border border-red-200 rounded-lg hover:bg-red-50 transition-colors disabled:opacity-40"
                title="Permanently delete"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Search & Link modal */}
      {searchModalId && activeItem && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          onClick={(e) => {
            if (e.target === e.currentTarget) closeSearchModal();
          }}
        >
          <div className="bg-surface-1 rounded-xl shadow-xl w-full max-w-lg mx-4 max-h-[80vh] flex flex-col">
            <div className="p-4 border-b border-border">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-text-primary">
                  Link &ldquo;{activeItem.title}&rdquo;
                </h3>
                <button
                  type="button"
                  onClick={closeSearchModal}
                  className="text-text-primary hover:text-text-primary transition-colors text-lg"
                  aria-label="Close"
                >
                  &#10005;
                </button>
              </div>

              <input
                ref={inputRef}
                type="text"
                value={searchQuery}
                onChange={(e) => handleSearchInput(e.target.value)}
                placeholder="Search by title or author..."
                className="w-full px-3 py-2 text-sm border border-border rounded-lg bg-surface-1 text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-2 focus:ring-accent"
              />
            </div>

            <div className="flex-1 overflow-y-auto p-4">
              {searching && (
                <p className="text-xs text-text-primary text-center py-4">
                  Searching...
                </p>
              )}

              {!searching && searchResults.length === 0 && searchQuery.trim() && (
                <p className="text-xs text-text-primary text-center py-4">
                  No results found.
                </p>
              )}

              {searchResults.length > 0 && (
                <ul className="space-y-2">
                  {searchResults.map((book) => (
                    <li
                      key={book.open_library_id}
                      className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-2 transition-colors"
                    >
                      {book.cover_url ? (
                        <img
                          src={book.cover_url}
                          alt=""
                          className="w-10 h-14 object-cover rounded shrink-0"
                        />
                      ) : (
                        <BookCoverPlaceholder title={book.title} className="w-10 h-14 shrink-0" />
                      )}
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-text-primary truncate">
                          {book.title}
                        </p>
                        <p className="text-xs text-text-primary truncate">
                          {book.authors}
                          {book.publication_year && ` (${book.publication_year})`}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() =>
                          resolveImport(searchModalId, book.open_library_id)
                        }
                        disabled={resolving === searchModalId}
                        className="px-3 py-1.5 text-xs font-medium bg-accent text-text-inverted rounded-lg hover:bg-surface-3 transition-colors disabled:opacity-40 shrink-0"
                      >
                        {resolving === searchModalId ? "Linking..." : "Link"}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
