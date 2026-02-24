"use client";

import Link from "next/link";
import { useState, ReactNode } from "react";
import { TagKey, TagValue } from "@/components/book-tag-picker";

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

type StatusFilter = { kind: "status"; slug: string; name: string };
type AllBooksFilter = { kind: "all" };
type TagFilter = { kind: "tag"; slug: string; name: string };
type LabelFilter = { kind: "label"; keySlug: string; keyName: string; valueSlug: string; valueName: string };
type ActiveFilter = StatusFilter | AllBooksFilter | TagFilter | LabelFilter;

type TagTreeNode = {
  collection: ShelfSummary;
  label: string;
  children: TagTreeNode[];
};

type ValueTreeNode = {
  value: TagValue;
  label: string;
  children: ValueTreeNode[];
};

// ── Tree builders ─────────────────────────────────────────────────────────────

function buildTagTree(collections: ShelfSummary[]): TagTreeNode[] {
  const sorted = [...collections].sort((a, b) => a.slug.localeCompare(b.slug));
  const root: TagTreeNode[] = [];
  const map = new Map<string, TagTreeNode>();
  for (const col of sorted) {
    const parts = col.slug.split("/");
    const node: TagTreeNode = {
      collection: col,
      label: col.name || parts[parts.length - 1],
      children: [],
    };
    map.set(col.slug, node);
    if (parts.length === 1) {
      root.push(node);
    } else {
      const parentSlug = parts.slice(0, -1).join("/");
      const parent = map.get(parentSlug);
      if (parent) parent.children.push(node);
      else root.push(node);
    }
  }
  return root;
}

function buildValueTree(values: TagValue[]): ValueTreeNode[] {
  const sorted = [...values].sort((a, b) => a.slug.localeCompare(b.slug));
  const root: ValueTreeNode[] = [];
  const map = new Map<string, ValueTreeNode>();
  for (const val of sorted) {
    const parts = val.slug.split("/");
    const node: ValueTreeNode = {
      value: val,
      label: parts[parts.length - 1],
      children: [],
    };
    map.set(val.slug, node);
    if (parts.length === 1) {
      root.push(node);
    } else {
      const parentSlug = parts.slice(0, -1).join("/");
      const parent = map.get(parentSlug);
      if (parent) parent.children.push(node);
      else root.push(node);
    }
  }
  return root;
}

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
    kind: "status",
    slug: initialShelf.slug,
    name: initialShelf.name,
  });
  const [loading, setLoading] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkWorking, setBulkWorking] = useState(false);
  const [showStatusMenu, setShowStatusMenu] = useState(false);
  const [showRateMenu, setShowRateMenu] = useState(false);
  const [showLabelsMenu, setShowLabelsMenu] = useState(false);
  const [showTagsMenu, setShowTagsMenu] = useState(false);
  const [localShelves, setLocalShelves] = useState(allShelves);

  // ── Navigation ───────────────────────────────────────────────────────────────

  // Extract status key from tagKeys
  const statusKey = tagKeys.find((k) => k.slug === "status") ?? null;
  const statusValues = statusKey?.values ?? [];
  const statusKeyId = statusKey?.id ?? null;
  const nonStatusTagKeys = tagKeys.filter((k) => k.slug !== "status");

  function closeMenus() {
    setShowStatusMenu(false);
    setShowRateMenu(false);
    setShowLabelsMenu(false);
    setShowTagsMenu(false);
  }

  async function navigateToStatus(slug: string, name: string) {
    if (filter.kind === "status" && filter.slug === slug) return;
    setLoading(true);
    setSelectedIds(new Set());
    closeMenus();
    const res = await fetch(`/api/users/${username}/books?status=${slug}`);
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      setBooks(data.books ?? []);
      setFilter({ kind: "status", slug, name });
    }
  }

  async function navigateToAllBooks() {
    if (filter.kind === "all") return;
    setLoading(true);
    setSelectedIds(new Set());
    closeMenus();
    const res = await fetch(`/api/users/${username}/books`);
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      // Flatten all statuses into a single book list
      const allBooks: Book[] = (data.statuses ?? []).flatMap(
        (s: { books: Book[] }) => s.books ?? []
      );
      setBooks(allBooks);
      setFilter({ kind: "all" });
    }
  }

  async function navigateToTag(slug: string, name: string) {
    if (filter.kind === "tag" && filter.slug === slug) return;
    setLoading(true);
    setSelectedIds(new Set());
    closeMenus();
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
    closeMenus();
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

  // ── Bulk actions ─────────────────────────────────────────────────────────────

  async function massRemove() {
    setBulkWorking(true);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/me/books/${b.open_library_id}`, {
          method: "DELETE",
        })
      )
    );
    setBooks((prev) => prev.filter((b) => !selectedIds.has(b.book_id)));
    setSelectedIds(new Set());
    setBulkWorking(false);
  }

  async function massChangeStatus(value: { id: string; name: string; slug: string }) {
    if (!statusKeyId) return;
    setBulkWorking(true);
    setShowStatusMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/me/books/${b.open_library_id}/tags/${statusKeyId}`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ value_id: value.id }),
        })
      )
    );
    // If viewing a specific status, remove moved books from view
    if (filter.kind === "status" && filter.slug !== value.slug) {
      setBooks((prev) => prev.filter((b) => !selectedIds.has(b.book_id)));
    }
    setSelectedIds(new Set());
    setBulkWorking(false);
  }

  async function massRate(rating: number) {
    setBulkWorking(true);
    setShowRateMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/me/books/${b.open_library_id}`, {
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

  async function massSetLabel(keyId: string, valueId: string) {
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

  async function massClearLabel(keyId: string) {
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

  async function massAddTag(tagCollection: ShelfSummary) {
    setBulkWorking(true);
    setShowTagsMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/shelves/${tagCollection.id}/books`, {
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
    setBulkWorking(false);
  }

  async function deleteShelf(shelf: ShelfSummary) {
    if (!confirm(`Delete "${shelf.name}"? This cannot be undone.`)) return;
    const res = await fetch(`/api/me/shelves/${shelf.id}`, { method: "DELETE" });
    if (!res.ok) return;
    const updated = localShelves.filter((s) => s.id !== shelf.id);
    setLocalShelves(updated);
    const isViewingDeleted =
      filter.kind === "tag" && filter.slug === shelf.slug;
    if (isViewingDeleted) {
      navigateToAllBooks();
    }
  }

  async function massRemoveTag(tagCollection: ShelfSummary) {
    setBulkWorking(true);
    setShowTagsMenu(false);
    const targets = books.filter((b) => selectedIds.has(b.book_id));
    await Promise.all(
      targets.map((b) =>
        fetch(`/api/shelves/${tagCollection.id}/books/${b.open_library_id}`, {
          method: "DELETE",
        })
      )
    );
    setBulkWorking(false);
  }

  // ── Derived ───────────────────────────────────────────────────────────────────

  const selectedCount = selectedIds.size;

  const tagCollections = localShelves.filter((s) => s.collection_type === "tag");
  const tagTree = buildTagTree(tagCollections);

  // Status values to show in "Change status" menu (exclude current if viewing a status)
  const changeStatusTargets = statusValues.filter(
    (v) => !(filter.kind === "status" && filter.slug === v.slug)
  );

  // ── Render ────────────────────────────────────────────────────────────────────

  return (
    <div className="flex flex-1 min-h-0 overflow-hidden max-w-7xl mx-auto w-full">
      {/* ── Sidebar ─────────────────────────────────────────────────────────── */}
      <aside className="w-48 shrink-0 border-r border-stone-200 overflow-y-auto py-3 flex flex-col gap-5">
        {/* All Books */}
        <div>
          <SidebarItem
            label="All Books"
            count={0}
            active={filter.kind === "all"}
            onClick={() => navigateToAllBooks()}
          />
        </div>

        {/* Status values */}
        {statusValues.length > 0 && (
          <div>
            <p className="px-4 mb-1 text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
              Status
            </p>
            {statusValues.map((v) => (
              <SidebarItem
                key={v.id}
                label={v.name}
                count={0}
                active={filter.kind === "status" && filter.slug === v.slug}
                onClick={() => navigateToStatus(v.slug, v.name)}
              />
            ))}
          </div>
        )}

        {/* Path-based tag collections */}
        {tagCollections.length > 0 && (
          <SidebarSection label="Tags" count={tagTree.length}>
            {tagTree.map((node) => (
              <TagTreeItem
                key={node.collection.id}
                node={node}
                depth={0}
                filter={filter}
                onNavigate={navigateToTag}
                onDelete={deleteShelf}
              />
            ))}
          </SidebarSection>
        )}

        {/* Key-value labels (excluding Status, which is shown above) */}
        {nonStatusTagKeys.length > 0 && (
          <SidebarSection label="Labels" count={nonStatusTagKeys.length}>
            {nonStatusTagKeys.map((key) => (
              <LabelKeyItem
                key={key.id}
                tagKey={key}
                filter={filter}
                onNavigate={navigateToLabel}
              />
            ))}
          </SidebarSection>
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

            {/* Rate dropdown */}
            <div className="relative">
              <button
                onClick={() => {
                  setShowRateMenu((v) => !v);
                  setShowStatusMenu(false);
                  setShowLabelsMenu(false);
                  setShowTagsMenu(false);
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

            {/* Change status dropdown */}
            {changeStatusTargets.length > 0 && (
              <div className="relative">
                <button
                  onClick={() => {
                    setShowStatusMenu((v) => !v);
                    setShowRateMenu(false);
                    setShowLabelsMenu(false);
                    setShowTagsMenu(false);
                  }}
                  disabled={bulkWorking}
                  className="text-xs px-3 py-1.5 rounded border border-stone-300 text-stone-600 hover:border-stone-500 hover:text-stone-800 disabled:opacity-50 transition-colors"
                >
                  Change status
                </button>
                {showStatusMenu && (
                  <div className="absolute top-full left-0 mt-1 z-20 bg-white border border-stone-200 rounded shadow-md min-w-[160px]">
                    {changeStatusTargets.map((v) => (
                      <button
                        key={v.id}
                        onClick={() => massChangeStatus(v)}
                        className="w-full text-left px-3 py-2 text-xs text-stone-700 hover:bg-stone-50 transition-colors"
                      >
                        {v.name}
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

            {/* Tags — available in all views */}
            {tagCollections.length > 0 && (
              <div className="relative">
                <button
                  onClick={() => {
                    setShowTagsMenu((v) => !v);
                    setShowRateMenu(false);
                    setShowStatusMenu(false);
                    setShowLabelsMenu(false);
                  }}
                  disabled={bulkWorking}
                  className="text-xs px-3 py-1.5 rounded border border-stone-300 text-stone-600 hover:border-stone-500 hover:text-stone-800 disabled:opacity-50 transition-colors"
                >
                  Tags
                </button>
                {showTagsMenu && (
                  <div className="absolute top-full left-0 mt-1 z-20 bg-white border border-stone-200 rounded shadow-md w-56 max-h-80 overflow-y-auto">
                    {tagCollections.map((t) => {
                      const label = t.name || t.slug.split("/").pop() || t.slug;
                      return (
                        <div key={t.id} className="border-b border-stone-100 last:border-0 flex items-center">
                          <button
                            onClick={() => massAddTag(t)}
                            className="flex-1 text-left px-3 py-2 text-xs text-stone-700 hover:bg-stone-50 transition-colors truncate"
                          >
                            {label}
                          </button>
                          <button
                            onClick={() => massRemoveTag(t)}
                            title="Remove from tag"
                            className="px-2.5 py-2 text-sm text-stone-300 hover:text-red-400 transition-colors shrink-0 leading-none"
                          >
                            ×
                          </button>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Labels — available in all views */}
            {nonStatusTagKeys.length > 0 && (
              <div className="relative">
                <button
                  onClick={() => {
                    setShowLabelsMenu((v) => !v);
                    setShowRateMenu(false);
                    setShowStatusMenu(false);
                    setShowTagsMenu(false);
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
                            onClick={() => massClearLabel(key.id)}
                            className="text-[10px] text-stone-300 hover:text-red-400 transition-colors shrink-0"
                          >
                            clear
                          </button>
                        </div>
                        <div className="pb-1.5">
                          {key.values.map((val) => (
                            <button
                              key={val.id}
                              onClick={() => massSetLabel(key.id, val.id)}
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
                : filter.kind === "all"
                  ? "All Books"
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

// ── Sidebar components ─────────────────────────────────────────────────────────

function SidebarItem({
  label,
  count,
  active,
  onClick,
  onDelete,
}: {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
  onDelete?: () => void;
}) {
  return (
    <div className="group flex items-center">
      <button
        onClick={onClick}
        className={`flex-1 text-left px-4 py-1.5 text-sm flex items-center justify-between transition-colors ${
          active
            ? "bg-stone-100 text-stone-900 font-medium"
            : "text-stone-600 hover:bg-stone-50 hover:text-stone-900"
        }`}
      >
        <span className="truncate">{label}</span>
        <span className="text-xs text-stone-400 ml-2 shrink-0">{count}</span>
      </button>
      {onDelete && (
        <button
          onClick={onDelete}
          title="Delete"
          className="opacity-0 group-hover:opacity-100 pr-3 text-stone-300 hover:text-red-400 transition-all leading-none shrink-0"
        >
          ×
        </button>
      )}
    </div>
  );
}

function SidebarSection({
  label,
  count,
  children,
}: {
  label: string;
  count: number;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  return (
    <div>
      <button
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center gap-1 px-4 mb-0.5 text-[10px] font-semibold text-stone-400 uppercase tracking-wider hover:text-stone-600 transition-colors"
      >
        <span
          className={`text-[13px] leading-none transition-transform inline-block ${
            open ? "rotate-90" : ""
          }`}
        >
          ›
        </span>
        <span className="ml-0.5">{label}</span>
        <span className="ml-auto font-normal normal-case text-stone-300">{count}</span>
      </button>
      {open && <div className="mb-1">{children}</div>}
    </div>
  );
}

function TagTreeItem({
  node,
  depth,
  filter,
  onNavigate,
  onDelete,
}: {
  node: TagTreeNode;
  depth: number;
  filter: ActiveFilter;
  onNavigate: (slug: string, name: string) => void;
  onDelete?: (collection: ShelfSummary) => void;
}) {
  const [open, setOpen] = useState(false);
  const hasChildren = node.children.length > 0;
  const isActive = filter.kind === "tag" && filter.slug === node.collection.slug;
  const pl = `${0.75 + depth * 0.75}rem`;

  return (
    <div>
      <div className="group flex items-center" style={{ paddingLeft: pl }}>
        <button
          onClick={() => setOpen((v) => !v)}
          tabIndex={hasChildren ? 0 : -1}
          className={`w-4 h-full flex items-center justify-center text-stone-300 hover:text-stone-500 shrink-0 transition-transform ${
            open ? "rotate-90" : ""
          } ${!hasChildren ? "invisible pointer-events-none" : ""}`}
        >
          <span className="text-[12px] leading-none">›</span>
        </button>
        <button
          onClick={() => onNavigate(node.collection.slug, node.label)}
          className={`flex-1 flex items-center justify-between py-1 text-xs transition-colors ${
            isActive
              ? "text-stone-900 font-medium"
              : "text-stone-500 hover:text-stone-900"
          }`}
        >
          <span className="truncate">{node.label}</span>
          <span className="text-[11px] text-stone-300 ml-2 shrink-0">
            {hasChildren ? node.children.length : node.collection.item_count}
          </span>
        </button>
        {onDelete && (
          <button
            onClick={() => onDelete(node.collection)}
            title="Delete"
            className="opacity-0 group-hover:opacity-100 pr-3 text-stone-300 hover:text-red-400 transition-all leading-none shrink-0 text-sm"
          >
            ×
          </button>
        )}
      </div>
      {open &&
        hasChildren &&
        node.children.map((child) => (
          <TagTreeItem
            key={child.collection.id}
            node={child}
            depth={depth + 1}
            filter={filter}
            onNavigate={onNavigate}
            onDelete={onDelete}
          />
        ))}
    </div>
  );
}

function LabelKeyItem({
  tagKey,
  filter,
  onNavigate,
}: {
  tagKey: TagKey;
  filter: ActiveFilter;
  onNavigate: (keySlug: string, keyName: string, valueSlug: string, valueName: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const valueTree = buildValueTree(tagKey.values);
  const hasValues = valueTree.length > 0;

  return (
    <div>
      <div className="flex items-center pl-3">
        <button
          onClick={() => setOpen((v) => !v)}
          tabIndex={hasValues ? 0 : -1}
          className={`w-4 h-full flex items-center justify-center text-stone-300 hover:text-stone-500 shrink-0 transition-transform ${
            open ? "rotate-90" : ""
          } ${!hasValues ? "invisible pointer-events-none" : ""}`}
        >
          <span className="text-[12px] leading-none">›</span>
        </button>
        <span className="flex-1 flex items-center justify-between pr-3 py-1 text-xs text-stone-500">
          <span className="truncate">{tagKey.name}</span>
          {hasValues && (
            <span className="text-[11px] text-stone-300 ml-2 shrink-0">
              {valueTree.length}
            </span>
          )}
        </span>
      </div>
      {open &&
        hasValues &&
        valueTree.map((node) => (
          <ValueTreeItem
            key={node.value.id}
            node={node}
            depth={1}
            tagKey={tagKey}
            filter={filter}
            onNavigate={onNavigate}
          />
        ))}
    </div>
  );
}

function ValueTreeItem({
  node,
  depth,
  tagKey,
  filter,
  onNavigate,
}: {
  node: ValueTreeNode;
  depth: number;
  tagKey: TagKey;
  filter: ActiveFilter;
  onNavigate: (keySlug: string, keyName: string, valueSlug: string, valueName: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const hasChildren = node.children.length > 0;
  const isActive =
    filter.kind === "label" &&
    filter.keySlug === tagKey.slug &&
    filter.valueSlug === node.value.slug;
  const pl = `${0.75 + depth * 0.75}rem`;

  return (
    <div>
      <div className="flex items-center" style={{ paddingLeft: pl }}>
        <button
          onClick={() => setOpen((v) => !v)}
          tabIndex={hasChildren ? 0 : -1}
          className={`w-4 h-full flex items-center justify-center text-stone-300 hover:text-stone-500 shrink-0 transition-transform ${
            open ? "rotate-90" : ""
          } ${!hasChildren ? "invisible pointer-events-none" : ""}`}
        >
          <span className="text-[12px] leading-none">›</span>
        </button>
        <button
          onClick={() =>
            onNavigate(tagKey.slug, tagKey.name, node.value.slug, node.value.name)
          }
          className={`flex-1 flex items-center justify-between pr-3 py-1 text-xs transition-colors ${
            isActive
              ? "text-stone-900 font-medium"
              : "text-stone-400 hover:text-stone-900"
          }`}
        >
          <span className="truncate">{node.label}</span>
          {hasChildren && (
            <span className="text-[11px] text-stone-300 ml-2 shrink-0">
              {node.children.length}
            </span>
          )}
        </button>
      </div>
      {open &&
        hasChildren &&
        node.children.map((child) => (
          <ValueTreeItem
            key={child.value.id}
            node={child}
            depth={depth + 1}
            tagKey={tagKey}
            filter={filter}
            onNavigate={onNavigate}
          />
        ))}
    </div>
  );
}
