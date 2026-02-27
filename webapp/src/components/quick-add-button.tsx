"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import type { StatusValue } from "@/components/shelf-picker";

export default function QuickAddButton({
  openLibraryId,
  title,
  coverUrl,
  statusValues,
  statusKeyId,
}: {
  openLibraryId: string;
  title: string;
  coverUrl: string | null;
  statusValues: StatusValue[];
  statusKeyId: string;
}) {
  const [added, setAdded] = useState(false);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  async function addWithStatus(value: StatusValue) {
    setOpen(false);
    setLoading(true);
    const res = await fetch("/api/me/books", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        open_library_id: openLibraryId,
        title,
        cover_url: coverUrl,
        status_slug: value.slug,
      }),
    });
    setLoading(false);
    if (res.ok) setAdded(true);
  }

  async function quickAdd() {
    const wantToRead = statusValues.find((v) => v.slug === "want-to-read");
    if (wantToRead) {
      await addWithStatus(wantToRead);
    }
  }

  if (added) {
    return (
      <div className="absolute bottom-1 right-1 z-10">
        <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent text-white">
          Added
        </span>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className="absolute bottom-1 right-1 z-10 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity pointer-events-none group-hover:pointer-events-auto focus-within:pointer-events-auto"
    >
      <div className="flex items-center">
        <button
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            quickAdd();
          }}
          disabled={loading}
          className="text-[10px] px-1.5 py-0.5 rounded-l bg-surface-0 border border-border text-text-primary hover:bg-surface-2 transition-colors disabled:opacity-50 whitespace-nowrap"
        >
          {loading ? "..." : "Want to read"}
        </button>
        <button
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setOpen(!open);
          }}
          aria-label="More options"
          aria-haspopup="true"
          aria-expanded={open}
          className="text-[10px] px-1 py-0.5 rounded-r bg-surface-0 border border-l-0 border-border text-text-primary hover:bg-surface-2 transition-colors"
        >
          ▾
        </button>
      </div>

      {open && (
        <div className="absolute right-0 bottom-full mb-1 bg-surface-0 border border-border rounded shadow-sm min-w-[140px]">
          {statusValues.map((value) => (
            <button
              key={value.id}
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                addWithStatus(value);
              }}
              className="w-full text-left px-2.5 py-1.5 text-[11px] text-text-primary hover:bg-surface-2 transition-colors"
            >
              {value.name}
            </button>
          ))}
          <div className="border-t border-border mx-1.5" />
          <Link
            href={`/books/${openLibraryId}`}
            onClick={(e) => e.stopPropagation()}
            className="block w-full text-left px-2.5 py-1.5 text-[11px] text-text-primary hover:bg-surface-2 transition-colors"
          >
            Rate & review
          </Link>
        </div>
      )}
    </div>
  );
}
