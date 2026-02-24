"use client";

import { useEffect, useRef, useState } from "react";

export type StatusValue = {
  id: string;
  name: string;
  slug: string;
};

export default function StatusPicker({
  openLibraryId,
  title,
  coverUrl,
  statusValues,
  statusKeyId,
  currentStatusValueId,
}: {
  openLibraryId: string;
  title: string;
  coverUrl: string | null;
  statusValues: StatusValue[];
  statusKeyId: string;
  currentStatusValueId: string | null;
}) {
  const [activeValueId, setActiveValueId] = useState(currentStatusValueId);
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

  const currentValue = statusValues.find((v) => v.id === activeValueId) ?? null;

  async function selectStatus(value: StatusValue) {
    setOpen(false);
    if (value.id === activeValueId) return;
    setLoading(true);

    if (!activeValueId) {
      // Book not in library yet — add it with status
      const res = await fetch("/api/me/books", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          open_library_id: openLibraryId,
          title,
          cover_url: coverUrl,
          status_value_id: value.id,
        }),
      });
      setLoading(false);
      if (res.ok) setActiveValueId(value.id);
    } else {
      // Change status via tag endpoint
      const res = await fetch(`/api/me/books/${openLibraryId}/tags/${statusKeyId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ value_id: value.id }),
      });
      setLoading(false);
      if (res.ok) setActiveValueId(value.id);
    }
  }

  async function removeFromLibrary() {
    if (!activeValueId) return;
    setOpen(false);
    setLoading(true);

    const res = await fetch(`/api/me/books/${openLibraryId}`, {
      method: "DELETE",
    });

    setLoading(false);
    if (res.ok) setActiveValueId(null);
  }

  return (
    <div ref={containerRef} className="relative shrink-0">
      <button
        onClick={() => setOpen(!open)}
        disabled={loading}
        className={`text-xs px-2.5 py-1 rounded border transition-colors disabled:opacity-50 whitespace-nowrap ${
          currentValue
            ? "border-stone-900 bg-stone-900 text-white hover:bg-stone-700"
            : "border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700"
        }`}
      >
        {loading ? "..." : currentValue ? currentValue.name : "Add to library"}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-10 bg-white border border-stone-200 rounded shadow-sm min-w-[160px]">
          {statusValues.map((value) => (
            <button
              key={value.id}
              onClick={() => selectStatus(value)}
              className={`w-full text-left px-3 py-2 text-xs transition-colors hover:bg-stone-50 ${
                value.id === activeValueId
                  ? "text-stone-900 font-medium"
                  : "text-stone-600"
              }`}
            >
              {value.id === activeValueId ? "\u2713 " : ""}
              {value.name}
            </button>
          ))}
          {activeValueId && (
            <>
              <div className="border-t border-stone-100 mx-2" />
              <button
                onClick={removeFromLibrary}
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
