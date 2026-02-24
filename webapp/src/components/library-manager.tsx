"use client";

import Link from "next/link";
import { useState } from "react";
import { TagKey } from "@/components/book-tag-picker";

// ── Types ─────────────────────────────────────────────────────────────────────

type Book = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
};

export type ShelfSummary = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  item_count: number;
  collection_type: string;
};

type ShelfFilter = { kind: "shelf"; id: string; name: string; slug: string };
type TagFilter = { kind: "tag"; slug: string; name: string };
type LabelFilter = { kind: "label"; keySlug: string; keyName: string; valueSlug: string; valueName: string };
type ActiveFilter = ShelfFilter | TagFilter | LabelFilter;

// ── Component ─────────────────────────────────────────────────────────────────

export default function LibraryManager({
  username,
  initialBooks,
  initialShelf,
  allShelves,
  tagKeys,
}: {
  username: string;
  initialBooks: Book[];
  initialShelf: { id: string; name: string; slug: string };
  allShelves: ShelfSummary[];
  tagKeys: TagKey[];
}) {
  const [books, setBooks] = useState(initialBooks);
  const [filter, setFilter] = useState<ActiveFilter>({
    kind: "shelf",
    id: initialShelf.id,
    name: initialShelf.name,
    slug: initialShelf.slug,
  });
  const [loading, setLoading] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkWorking, setBulkWorking] = useState(false);
  const [showMoveMenu, setShowMoveMenu] = useState(false);
  const [showRateMenu, setShowRateMenu] = useState(false);
  const [showLabelsMenu, setShowLabelsMenu] = useState(false);

  // ── Navigation ───────────────────────────────────────────────────────────────

  async function navigateToShelf(shelf: ShelfSummary) {
    if (filter.kind === "shelf" && filter.id === shelf.id) return;
    setLoading(true);
    setSelectedIds(new Set());
    setShowMoveMenu(false);
    setShowRateMenu(false);
    setShowLabelsMenu(false);
    const res = await fetch(`/api/users/${username}/shelves/${shelf.slug}`);
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      setBooks(data.books ?? []);
      setFilter({ kind: "shelf", id: shelf.id, name: shelf.name, slug: shelf.slug });
    }
  }

  async function navigateToTag(slug: string, name: string) {
    if (filter.kind === "tag" && filter.slug === slug) return;
    setLoading(true);
    setSelectedIds(new Set());
    setShowMoveMenu(false);
    setShowRateMenu(false);
    setShowLabelsMenu(false);
    const res = await fetch(`/api/users/${username}/tags/${slug}`);
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      setBooks(data.books ?? []);
      setFilter({ kind: "tag", slug, name });
    }
  }

  async function navigateToLabel(keySlug: string, keyName: string, valueSlug: string, valueName: string) {
    if (filter.kind === "label" && filter.keySlug === keySlug && filter.valueSlug === valueSlug) return;
    setLoading(true);
    setSelectedIds(new Set());
    setShowMoveMenu(false);
    setShowRateMenu(false);
    setShowLabelsMenu(false);
    const res = await fetch(`/api/users/${username}/labels/${keySlug}/${valueSlug}`);
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      setBooks(data.books ?? []);
      setFilter({ kind: "label", keySlug, keyName, valueSlug, valueName });
    }
  }

  // ── Selection ────────────────────────────────────────────────────────────────

  function toggleSelect(bookId: string) {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(bookId)) next.delete(bookId);
      else next.add(bookId);
      return next;
    });
  }

  // ── Bulk actions (shelf view only) ───────────────────────────────────────────

  async function massRemove() {
    if (filter.kind !== "shelf") return;
    const shelfId = filter.id;
    setBulkWorking(true);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/shelves/${shelfId}/books/${b.open_library_id}`, {
          method: "DELETE",
        })
      )
    );
    setBooks((prev) => prev.filter((b) => !selectedIds.has(b.book_id)));
    setSelectedIds(new Set());
    setBulkWorking(false);
  }

  async function massMoveToShelf(target: ShelfSummary) {
    if (filter.kind !== "shelf") return;
    setBulkWorking(true);
    setShowMoveMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/shelves/${target.id}/books`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            open_library_id: b.open_library_id,
            title: b.title,
            cover_url: b.cover_url,
          }),
        })
      )
    );
    // Refresh current shelf (books in exclusive groups will have moved away)
    const res = await fetch(`/api/users/${username}/shelves/${filter.slug}`);
    if (res.ok) {
      const data = await res.json();
      setBooks(data.books ?? []);
    }
    setSelectedIds(new Set());
    setBulkWorking(false);
  }

  async function massRate(rating: number) {
    if (filter.kind !== "shelf") return;
    const shelfId = filter.id;
    setBulkWorking(true);
    setShowRateMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/shelves/${shelfId}/books/${b.open_library_id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ rating }),
        })
      )
    );
    setBooks((prev) =>
      prev.map((b) => (selectedIds.has(b.book_id) ? { ...b, rating } : b))
    );
    setBulkWorking(false);
  }

  async function massSetTag(keyId: string, valueId: string) {
    setBulkWorking(true);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/me/books/${b.open_library_id}/tags/${keyId}`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ value_id: valueId }),
        })
      )
    );
    setBulkWorking(false);
  }

  async function massClearTag(keyId: string) {
    setBulkWorking(true);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/me/books/${b.open_library_id}/tags/${keyId}`, {
          method: "DELETE",
        })
      )
    );
    setBulkWorking(false);
  }

  // ── Derived ───────────────────────────────────────────────────────────────────

  const selectedCount = selectedIds.size;
  const isShelfView = filter.kind === "shelf";

  const defaultShelves = allShelves.filter(
    (s) => s.exclusive_group === "read_status"
  );
  const customShelves = allShelves.filter(
    (s) =>
      s.exclusive_group !== "read_status" && s.collection_type === "shelf"
  );
  const tagCollections = allShelves.filter((s) => s.collection_type === "tag");

  const moveTargets = isShelfView
    ? allShelves.filter(
        (s) =>
          s.collection_type !== "tag" &&
          !(filter.kind === "shelf" && s.id === filter.id)
      )
    : [];

  // ── Render ────────────────────────────────────────────────────────────────────

  return (
    <div className="flex flex-1 min-h-0 overflow-hidden">
      {/* ── Sidebar ─────────────────────────────────────────────────────────── */}
      <aside className="w-48 shrink-0 border-r border-stone-200 overflow-y-auto py-3 flex flex-col gap-5">
        {/* Default shelves */}
        <div>
          <p className="px-4 mb-1 text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
            Shelves
          </p>
          {defaultShelves.map((s) => (
            <SidebarItem
              key={s.id}
              label={s.name}
              count={s.item_count}
              active={filter.kind === "shelf" && filter.id === s.id}
              onClick={() => navigateToShelf(s)}
            />
          ))}
        </div>

        {/* Custom shelves */}
        {customShelves.length > 0 && (
          <div>
            <p className="px-4 mb-1 text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
              Custom
            </p>
            {customShelves.map((s) => (
              <SidebarItem
                key={s.id}
                label={s.name}
                count={s.item_count}
                active={filter.kind === "shelf" && filter.id === s.id}
                onClick={() => navigateToShelf(s)}
              />
            ))}
          </div>
        )}

        {/* Path-based tag collections */}
        {tagCollections.length > 0 && (
          <div>
            <p className="px-4 mb-1 text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
              Tags
            </p>
            {tagCollections.map((t) => {
              const label = t.name || t.slug.split("/").pop() || t.slug;
              return (
                <SidebarItem
                  key={t.id}
                  label={label}
                  count={t.item_count}
                  active={filter.kind === "tag" && filter.slug === t.slug}
                  onClick={() => navigateToTag(t.slug, label)}
                />
              );
            })}
          </div>
        )}

        {/* Key-value labels */}
        {tagKeys.length > 0 && (
          <div>
            <p className="px-4 mb-1 text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
              Labels
            </p>
            {tagKeys.map((key) => (
              <div key={key.id}>
                <p className="px-4 py-1 text-xs text-stone-500">{key.name}</p>
                {key.values.map((val) => {
                  const depth = (val.slug.match(/\//g) ?? []).length;
                  const displayName = val.name.split("/").pop() ?? val.name;
                  return (
                    <button
                      key={val.id}
                      onClick={() => navigateToLabel(key.slug, key.name, val.slug, val.name)}
                      style={{ paddingLeft: `${1.75 + depth * 0.75}rem` }}
                      className={`w-full text-left pr-4 py-1 text-xs transition-colors ${
                        filter.kind === "label" &&
                        filter.keySlug === key.slug &&
                        filter.valueSlug === val.slug
                          ? "bg-stone-100 text-stone-900 font-medium"
                          : "text-stone-400 hover:bg-stone-50 hover:text-stone-900"
                      }`}
                    >
                      {displayName}
                    </button>
                  );
                })}
              </div>
            ))}
          </div>
        )}
      </aside>

      {/* ── Main ────────────────────────────────────────────────────────────── */}
      <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
        {/* Top bar */}
        {selectedCount > 0 ? (
          <div className="border-b border-stone-200 bg-stone-50 px-5 py-2.5 flex items-center gap-3 flex-wrap shrink-0">
            <span className="text-sm font-medium text-stone-700">
              {selectedCount} {selectedCount === 1 ? "book" : "books"} selected
            </span>

            {isShelfView && (
              <>
                {/* Rate dropdown */}
                <div className="relative">
                  <button
                    onClick={() => {
                      setShowRateMenu((v) => !v);
                      setShowMoveMenu(false);
                    }}
                    disabled={bulkWorking}
                    className="text-xs px-3 py-1.5 rounded border border-stone-300 text-stone-600 hover:border-stone-500 hover:text-stone-800 disabled:opacity-50 transition-colors"
                  >
                    Rate
                  </button>
                  {showRateMenu && (
                    <div className="absolute top-full left-0 mt-1 z-20 bg-white border border-stone-200 rounded shadow-md flex gap-0.5 p-1.5">
                      {[1, 2, 3, 4, 5].map((n) => (
                        <button
                          key={n}
                          onClick={() => massRate(n)}
                          className="text-xl text-stone-300 hover:text-amber-500 transition-colors px-1"
                        >
                          ★
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Move to shelf dropdown */}
                {moveTargets.length > 0 && (
                  <div className="relative">
                    <button
                      onClick={() => {
                        setShowMoveMenu((v) => !v);
                        setShowRateMenu(false);
                      }}
                      disabled={bulkWorking}
                      className="text-xs px-3 py-1.5 rounded border border-stone-300 text-stone-600 hover:border-stone-500 hover:text-stone-800 disabled:opacity-50 transition-colors"
                    >
                      Move to shelf
                    </button>
                    {showMoveMenu && (
                      <div className="absolute top-full left-0 mt-1 z-20 bg-white border border-stone-200 rounded shadow-md min-w-[160px]">
                        {moveTargets.map((s) => (
                          <button
                            key={s.id}
                            onClick={() => massMoveToShelf(s)}
                            className="w-full text-left px-3 py-2 text-xs text-stone-700 hover:bg-stone-50 transition-colors"
                          >
                            {s.name}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Remove */}
                <button
                  onClick={massRemove}
                  disabled={bulkWorking}
                  className="text-xs px-3 py-1.5 rounded border border-red-200 text-red-500 hover:border-red-400 hover:text-red-700 disabled:opacity-50 transition-colors"
                >
                  {bulkWorking ? "Working..." : "Remove"}
                </button>
              </>
            )}

            {/* Labels — available in both shelf and tag views */}
            {tagKeys.length > 0 && (
              <div className="relative">
                <button
                  onClick={() => {
                    setShowLabelsMenu((v) => !v);
                    setShowRateMenu(false);
                    setShowMoveMenu(false);
                  }}
                  disabled={bulkWorking}
                  className="text-xs px-3 py-1.5 rounded border border-stone-300 text-stone-600 hover:border-stone-500 hover:text-stone-800 disabled:opacity-50 transition-colors"
                >
                  Labels
                </button>
                {showLabelsMenu && (
                  <div className="absolute top-full left-0 mt-1 z-20 bg-white border border-stone-200 rounded shadow-md w-56 max-h-80 overflow-y-auto">
                    {tagKeys.map((key) => (
                      <div key={key.id} className="border-b border-stone-100 last:border-0">
                        <div className="px-3 pt-2.5 pb-1 flex items-center justify-between gap-2">
                          <span className="text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
                            {key.name}
                          </span>
                          <button
                            onClick={() => massClearTag(key.id)}
                            className="text-[10px] text-stone-300 hover:text-red-400 transition-colors shrink-0"
                          >
                            clear
                          </button>
                        </div>
                        <div className="pb-1.5">
                          {key.values.map((val) => (
                            <button
                              key={val.id}
                              onClick={() => massSetTag(key.id, val.id)}
                              className="w-full text-left px-3 py-1.5 text-xs text-stone-600 hover:bg-stone-50 hover:text-stone-900 transition-colors"
                            >
                              {val.name}
                            </button>
                          ))}
                          {key.values.length === 0 && (
                            <p className="px-3 py-1 text-xs text-stone-300">No values defined</p>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            <div className="ml-auto flex items-center gap-3">
              <button
                onClick={() =>
                  setSelectedIds(new Set(books.map((b) => b.book_id)))
                }
                className="text-xs text-stone-500 hover:text-stone-800 transition-colors"
              >
                Select all
              </button>
              <button
                onClick={() => setSelectedIds(new Set())}
                className="text-xs text-stone-500 hover:text-stone-800 transition-colors"
              >
                Clear
              </button>
            </div>
          </div>
        ) : (
          <div className="border-b border-stone-200 px-5 py-2.5 flex items-center gap-3 shrink-0">
            <span className="text-sm font-semibold text-stone-800">
              {filter.kind === "label"
                ? `${filter.keyName}: ${filter.valueName}`
                : filter.name}
            </span>
            <span className="text-xs text-stone-400">
              {books.length} {books.length === 1 ? "book" : "books"}
            </span>
          </div>
        )}

        {/* Book grid */}
        <div className="flex-1 overflow-y-auto p-5">
          {loading ? (
            <p className="text-sm text-stone-400">Loading...</p>
          ) : books.length === 0 ? (
            <p className="text-sm text-stone-400">No books here yet.</p>
          ) : (
            <ul className="grid grid-cols-5 sm:grid-cols-7 md:grid-cols-8 lg:grid-cols-10 xl:grid-cols-12 gap-3">
              {books.map((book) => {
                const selected = selectedIds.has(book.book_id);
                const anySelected = selectedIds.size > 0;
                return (
                  <li key={book.book_id} className="relative group">
                    {/* Checkbox */}
                    <button
                      onClick={() => toggleSelect(book.book_id)}
                      className={`absolute top-1 left-1 z-10 w-4 h-4 rounded border flex items-center justify-center transition-all ${
                        selected
                          ? "bg-stone-800 border-stone-800 text-white opacity-100"
                          : `bg-white border-stone-300 ${anySelected ? "opacity-100" : "opacity-0 group-hover:opacity-100"}`
                      }`}
                      aria-label={selected ? "Deselect" : "Select"}
                    >
                      {selected && (
                        <svg
                          viewBox="0 0 10 8"
                          className="w-2.5 h-2.5"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="1.5"
                        >
                          <path d="M1 4l3 3 5-6" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                      )}
                    </button>

                    {/* Cover */}
                    <Link
                      href={`/books/${book.open_library_id}`}
                      tabIndex={-1}
                      onClick={(e) => { if (anySelected) { e.preventDefault(); toggleSelect(book.book_id); } }}
                      className={`block transition-all ${selected ? "ring-2 ring-stone-700 ring-offset-1 rounded" : ""}`}
                    >
                      {book.cover_url ? (
                        <img
                          src={book.cover_url}
                          alt={book.title}
                          className={`w-full aspect-[2/3] object-cover rounded shadow-sm transition-all ${
                            selected
                              ? "opacity-70"
                              : "group-hover:shadow-md"
                          }`}
                          draggable={false}
                        />
                      ) : (
                        <div
                          className={`w-full aspect-[2/3] bg-stone-100 rounded shadow-sm flex items-end p-1.5 transition-all ${
                            selected ? "opacity-70" : "group-hover:shadow-md"
                          }`}
                        >
                          <span className="text-[9px] text-stone-400 leading-tight line-clamp-3">
                            {book.title}
                          </span>
                        </div>
                      )}
                    </Link>

                    {/* Rating */}
                    {book.rating != null && book.rating > 0 && (
                      <div
                        className="flex gap-px mt-1 px-0.5"
                        aria-label={`${book.rating} out of 5 stars`}
                      >
                        {[1, 2, 3, 4, 5].map((n) => (
                          <span
                            key={n}
                            className={`text-[8px] leading-none ${
                              n <= book.rating! ? "text-amber-500" : "text-stone-200"
                            }`}
                          >
                            ★
                          </span>
                        ))}
                      </div>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Sidebar item ──────────────────────────────────────────────────────────────

function SidebarItem({
  label,
  count,
  active,
  onClick,
}: {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-1.5 text-sm flex items-center justify-between transition-colors ${
        active
          ? "bg-stone-100 text-stone-900 font-medium"
          : "text-stone-600 hover:bg-stone-50 hover:text-stone-900"
      }`}
    >
      <span className="truncate">{label}</span>
      <span className="text-xs text-stone-400 ml-2 shrink-0">{count}</span>
    </button>
  );
}
