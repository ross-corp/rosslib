"use client";

import { useState } from "react";
import Link from "next/link";

type Shelf = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  collection_type: string;
  item_count: number;
};

type Book = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
};

type Operation = "union" | "intersection" | "difference";

const operationLabels: Record<Operation, { label: string; description: string }> = {
  union: {
    label: "Union",
    description: "Books in either list (combine both)",
  },
  intersection: {
    label: "Intersection",
    description: "Books in both lists",
  },
  difference: {
    label: "Difference",
    description: "Books in List A but not in List B",
  },
};

export default function SetOperationForm({
  shelves,
  username,
}: {
  shelves: Shelf[];
  username: string;
}) {
  const [collectionA, setCollectionA] = useState("");
  const [collectionB, setCollectionB] = useState("");
  const [operation, setOperation] = useState<Operation>("intersection");
  const [books, setBooks] = useState<Book[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [hasResult, setHasResult] = useState(false);

  const [saveName, setSaveName] = useState("");
  const [saving, setSaving] = useState(false);
  const [savedSlug, setSavedSlug] = useState("");
  const [isContinuous, setIsContinuous] = useState(false);

  const canCompute = collectionA && collectionB && collectionA !== collectionB;

  async function compute() {
    if (!canCompute) return;
    setLoading(true);
    setError("");
    setHasResult(false);
    setSavedSlug("");

    try {
      const res = await fetch("/api/me/shelves/set-operation", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          collection_a: collectionA,
          collection_b: collectionB,
          operation,
        }),
      });
      if (!res.ok) {
        const data = await res.json();
        setError(data.error || "Something went wrong");
        return;
      }
      const data = await res.json();
      setBooks(data.books || []);
      setHasResult(true);
    } catch {
      setError("Failed to compute set operation");
    } finally {
      setLoading(false);
    }
  }

  async function saveAsNewList() {
    if (!saveName.trim() || !canCompute) return;
    setSaving(true);
    setError("");

    try {
      const res = await fetch("/api/me/shelves/set-operation/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          collection_a: collectionA,
          collection_b: collectionB,
          operation,
          name: saveName.trim(),
          is_continuous: isContinuous,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "Failed to save");
        return;
      }
      setSavedSlug(data.slug);
      setSaveName("");
    } catch {
      setError("Failed to save list");
    } finally {
      setSaving(false);
    }
  }

  const nameA = shelves.find((s) => s.id === collectionA)?.name ?? "List A";
  const nameB = shelves.find((s) => s.id === collectionB)?.name ?? "List B";

  return (
    <div className="space-y-8">
      {/* Controls */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {/* List A */}
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">
            List A
          </label>
          <select
            value={collectionA}
            onChange={(e) => {
              setCollectionA(e.target.value);
              setHasResult(false);
            }}
            className="w-full px-3 py-2 text-sm border border-border rounded bg-surface-0 text-text-primary focus:outline-none focus:ring-2 focus:ring-accent"
          >
            <option value="">Select a list...</option>
            {shelves.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name} ({s.item_count})
              </option>
            ))}
          </select>
        </div>

        {/* Operation */}
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">
            Operation
          </label>
          <select
            value={operation}
            onChange={(e) => {
              setOperation(e.target.value as Operation);
              setHasResult(false);
            }}
            className="w-full px-3 py-2 text-sm border border-border rounded bg-surface-0 text-text-primary focus:outline-none focus:ring-2 focus:ring-accent"
          >
            {(Object.keys(operationLabels) as Operation[]).map((op) => (
              <option key={op} value={op}>
                {operationLabels[op].label}
              </option>
            ))}
          </select>
          <p className="text-xs text-text-primary mt-1">
            {operationLabels[operation].description}
          </p>
        </div>

        {/* List B */}
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">
            List B
          </label>
          <select
            value={collectionB}
            onChange={(e) => {
              setCollectionB(e.target.value);
              setHasResult(false);
            }}
            className="w-full px-3 py-2 text-sm border border-border rounded bg-surface-0 text-text-primary focus:outline-none focus:ring-2 focus:ring-accent"
          >
            <option value="">Select a list...</option>
            {shelves.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name} ({s.item_count})
              </option>
            ))}
          </select>
        </div>
      </div>

      {collectionA && collectionB && collectionA === collectionB && (
        <p className="text-sm text-amber-600">
          Please select two different lists.
        </p>
      )}

      <button
        onClick={compute}
        disabled={!canCompute || loading}
        className="px-4 py-2 text-sm bg-accent text-white rounded hover:bg-surface-3 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? "Computing..." : "Compare"}
      </button>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {/* Results */}
      {hasResult && (
        <div>
          <div className="flex items-baseline gap-3 mb-4">
            <h2 className="text-lg font-semibold text-text-primary">
              Result: {books.length} book{books.length !== 1 ? "s" : ""}
            </h2>
            <span className="text-sm text-text-primary">
              {operationLabels[operation].label} of {nameA} and {nameB}
            </span>
          </div>

          {books.length === 0 ? (
            <p className="text-sm text-text-primary">
              No books match this operation.
            </p>
          ) : (
            <>
              <ul className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-5 mb-8">
                {books.map((book) => (
                  <li key={book.book_id} className="group flex flex-col gap-2">
                    <Link
                      href={`/books/${book.open_library_id}`}
                      className="block"
                    >
                      {book.cover_url ? (
                        <img
                          src={book.cover_url}
                          alt={book.title}
                          className="w-full aspect-[2/3] object-cover rounded shadow-sm bg-surface-2 group-hover:shadow-md transition-shadow"
                        />
                      ) : (
                        <div className="w-full aspect-[2/3] bg-surface-2 rounded shadow-sm flex items-end p-2 group-hover:shadow-md transition-shadow">
                          <span className="text-xs text-text-primary leading-tight line-clamp-3">
                            {book.title}
                          </span>
                        </div>
                      )}
                    </Link>
                    <div className="min-w-0">
                      <Link
                        href={`/books/${book.open_library_id}`}
                        className="text-xs font-medium text-text-primary hover:text-text-primary line-clamp-2 leading-snug"
                      >
                        {book.title}
                      </Link>
                      {book.rating != null && book.rating > 0 && (
                        <div
                          className="flex gap-px mt-1"
                          aria-label={`${book.rating} out of 5 stars`}
                        >
                          {[1, 2, 3, 4, 5].map((n) => (
                            <span
                              key={n}
                              className={`text-[10px] leading-none ${
                                n <= book.rating!
                                  ? "text-amber-500"
                                  : "text-text-primary"
                              }`}
                            >
                              &#9733;
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </li>
                ))}
              </ul>

              {/* Save as new list */}
              <div className="border-t border-border pt-6">
                <h3 className="text-sm font-medium text-text-primary mb-3">
                  Save as new list
                </h3>
                {savedSlug ? (
                  <p className="text-sm text-green-700">
                    Saved!{" "}
                    <Link
                      href={`/${username}/library/${savedSlug}`}
                      className="underline hover:text-green-900"
                    >
                      View list
                    </Link>
                  </p>
                ) : (
                  <div className="space-y-3">
                    <div className="flex gap-2 items-center">
                      <input
                        type="text"
                        value={saveName}
                        onChange={(e) => setSaveName(e.target.value)}
                        placeholder="New list name..."
                        className="px-3 py-1.5 text-sm border border-border rounded bg-surface-0 text-text-primary focus:outline-none focus:ring-2 focus:ring-accent w-64"
                      />
                      <button
                        onClick={saveAsNewList}
                        disabled={!saveName.trim() || saving}
                        className="px-3 py-1.5 text-sm bg-accent text-white rounded hover:bg-surface-3 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {saving ? "Saving..." : "Save"}
                      </button>
                    </div>
                    <label className="flex items-center gap-2 text-sm text-text-primary cursor-pointer">
                      <input
                        type="checkbox"
                        checked={isContinuous}
                        onChange={(e) => setIsContinuous(e.target.checked)}
                        className="rounded border-border text-text-primary focus:ring-accent"
                      />
                      Keep updated — list auto-refreshes when source lists change
                    </label>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
