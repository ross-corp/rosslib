"use client";

import { useState } from "react";
import { useToast } from "@/components/toast";

export default function ExportForm() {
  const [downloading, setDownloading] = useState(false);
  const toast = useToast();

  async function handleExport() {
    setDownloading(true);
    try {
      const res = await fetch("/api/me/export/csv");
      if (!res.ok) throw new Error("Export failed");

      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "rosslib-export.csv";
      a.click();
      URL.revokeObjectURL(a.href);
      toast.success("Export downloaded");
    } catch {
      toast.error("Export failed. Please try again.");
    } finally {
      setDownloading(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-sm text-text-primary">
        <p className="font-medium text-text-primary mb-1">Columns included:</p>
        <p>Title, Author, ISBN13, Rating, Review, Date Added, Date Read, Status</p>
      </div>

      <button
        onClick={handleExport}
        disabled={downloading}
        className="rounded-md bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {downloading ? "Downloading..." : "Download CSV"}
      </button>
    </div>
  );
}
