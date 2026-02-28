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
  const [confirmRemove, setConfirmRemove] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
        setConfirmRemove(false);
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
    setConfirmRemove(false);
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
            ? "border-accent bg-accent text-white hover:bg-surface-3"
            : "border-border text-text-primary hover:border-border-strong hover:text-text-primary"
        }`}
      >
        {loading ? "..." : currentValue ? currentValue.name : "Add to library"}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-10 bg-surface-0 border border-border rounded shadow-sm min-w-[160px]">
          {statusValues.map((value) => (
            <button
              key={value.id}
              onClick={() => selectStatus(value)}
              className={`w-full text-left px-3 py-2 text-xs transition-colors hover:bg-surface-2 ${
                value.id === activeValueId
                  ? "text-text-primary font-medium"
                  : "text-text-primary"
              }`}
            >
              {value.id === activeValueId ? "\u2713 " : ""}
              {value.name}
            </button>
          ))}
          {activeValueId && !confirmRemove && (
            <>
              <div className="border-t border-border mx-2" />
              <button
                onClick={() => setConfirmRemove(true)}
                className="w-full text-left px-3 py-2 text-xs text-text-primary hover:text-text-primary hover:bg-surface-2 transition-colors"
              >
                Remove
              </button>
            </>
          )}
          {confirmRemove && (
            <div className="border-t border-border mx-2 mt-0">
              <p className="px-3 pt-2 pb-1 text-xs text-text-secondary">
                Remove this book? Rating, review, and progress will be deleted.
              </p>
              <div className="flex gap-1 px-3 pb-2">
                <button
                  onClick={removeFromLibrary}
                  className="text-xs px-2 py-1 rounded bg-red-600 text-white hover:bg-red-700 transition-colors"
                >
                  Remove
                </button>
                <button
                  onClick={() => setConfirmRemove(false)}
                  className="text-xs px-2 py-1 rounded border border-border text-text-secondary hover:text-text-primary transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
