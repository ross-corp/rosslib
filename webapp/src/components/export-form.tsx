"use client";

import { useState } from "react";

type Shelf = {
  id: string;
  name: string;
  slug: string;
  item_count: number;
};

export default function ExportForm({ shelves }: { shelves: Shelf[] }) {
  const [selected, setSelected] = useState("all");
  const [downloading, setDownloading] = useState(false);

  const totalBooks = shelves.reduce((sum, s) => sum + s.item_count, 0);

  async function handleExport() {
    setDownloading(true);
    try {
      const url =
        selected === "all"
          ? "/api/me/export/csv"
          : `/api/me/export/csv?shelf=${encodeURIComponent(selected)}`;

      const res = await fetch(url);
      if (!res.ok) throw new Error("Export failed");

      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "rosslib-export.csv";
      a.click();
      URL.revokeObjectURL(a.href);
    } catch {
      alert("Export failed. Please try again.");
    } finally {
      setDownloading(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <label htmlFor="shelf-select" className="block text-sm font-medium text-stone-700 mb-2">
          What to export
        </label>
        <select
          id="shelf-select"
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
          className="w-full max-w-sm rounded-md border border-stone-300 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-500 focus:outline-none focus:ring-1 focus:ring-stone-500"
        >
          <option value="all">All shelves ({totalBooks} books)</option>
          {shelves.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} ({s.item_count} books)
            </option>
          ))}
        </select>
      </div>

      <div className="text-sm text-stone-500">
        <p className="font-medium text-stone-700 mb-1">Columns included:</p>
        <p>Title, Author, ISBN13, Collection, Rating, Review, Date Added, Date Read</p>
      </div>

      <button
        onClick={handleExport}
        disabled={downloading}
        className="rounded-md bg-stone-900 px-4 py-2 text-sm font-medium text-white hover:bg-stone-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {downloading ? "Downloading..." : "Download CSV"}
      </button>
    </div>
  );
}
