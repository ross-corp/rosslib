"use client";

import { useEffect, useRef, useState } from "react";

export type Shelf = {
  id: string;
  name: string;
  slug: string;
};

export default function ShelfPicker({
  openLibraryId,
  title,
  coverUrl,
  shelves,
  initialShelfId,
}: {
  openLibraryId: string;
  title: string;
  coverUrl: string | null;
  shelves: Shelf[];
  initialShelfId: string | null;
}) {
  const [currentShelfId, setCurrentShelfId] = useState(initialShelfId);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  const currentShelf = shelves.find((s) => s.id === currentShelfId) ?? null;

  async function selectShelf(shelf: Shelf) {
    setOpen(false);
    if (shelf.id === currentShelfId) return;
    setLoading(true);

    const res = await fetch(`/api/shelves/${shelf.id}/books`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        open_library_id: openLibraryId,
        title,
        cover_url: coverUrl,
      }),
    });

    setLoading(false);
    if (res.ok) setCurrentShelfId(shelf.id);
  }

  async function removeFromShelf() {
    if (!currentShelfId) return;
    setOpen(false);
    setLoading(true);

    const res = await fetch(
      `/api/shelves/${currentShelfId}/books/${openLibraryId}`,
      { method: "DELETE" }
    );

    setLoading(false);
    if (res.ok) setCurrentShelfId(null);
  }

  return (
    <div ref={containerRef} className="relative shrink-0">
      <button
        onClick={() => setOpen(!open)}
        disabled={loading}
        className={`text-xs px-2.5 py-1 rounded border transition-colors disabled:opacity-50 whitespace-nowrap ${
          currentShelf
            ? "border-stone-900 bg-stone-900 text-white hover:bg-stone-700"
            : "border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700"
        }`}
      >
        {loading ? "..." : currentShelf ? currentShelf.name : "Add to shelf"}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-10 bg-white border border-stone-200 rounded shadow-sm min-w-[160px]">
          {shelves.map((shelf) => (
            <button
              key={shelf.id}
              onClick={() => selectShelf(shelf)}
              className={`w-full text-left px-3 py-2 text-xs transition-colors hover:bg-stone-50 ${
                shelf.id === currentShelfId
                  ? "text-stone-900 font-medium"
                  : "text-stone-600"
              }`}
            >
              {shelf.id === currentShelfId ? "✓ " : ""}
              {shelf.name}
            </button>
          ))}
          {currentShelfId && (
            <>
              <div className="border-t border-stone-100 mx-2" />
              <button
                onClick={removeFromShelf}
                className="w-full text-left px-3 py-2 text-xs text-stone-400 hover:text-stone-600 hover:bg-stone-50 transition-colors"
              >
                Remove
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
