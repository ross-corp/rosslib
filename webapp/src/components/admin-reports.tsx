"use client";

import { useState, useEffect, useCallback } from "react";

type ReportItem = {
  id: string;
  reporter_id: string;
  reporter_username: string;
  reporter_display_name: string | null;
  content_type: string;
  content_id: string;
  reason: string;
  details: string | null;
  status: string;
  reviewer_id: string | null;
  reviewer_username: string | null;
  created_at: string;
  content_preview: string;
};

const REASON_CLASSES: Record<string, string> = {
  spam: "bg-yellow-900/30 text-yellow-300",
  harassment: "bg-red-900/30 text-red-300",
  inappropriate: "bg-orange-900/30 text-orange-300",
  other: "bg-surface-2 text-text-primary",
};

const TYPE_CLASSES: Record<string, string> = {
  review: "bg-blue-900/30 text-blue-300",
  thread: "bg-purple-900/30 text-purple-300",
  comment: "bg-green-900/30 text-green-300",
  link: "bg-cyan-900/30 text-cyan-300",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export default function AdminReports() {
  const [statusFilter, setStatusFilter] = useState("pending");
  const [items, setItems] = useState<ReportItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  const fetchReports = useCallback(async (status: string) => {
    setLoading(true);
    const res = await fetch(`/api/admin/reports?status=${status}`);
    if (res.ok) {
      setItems(await res.json());
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchReports(statusFilter);
  }, [statusFilter, fetchReports]);

  async function updateStatus(id: string, newStatus: "reviewed" | "dismissed") {
    setUpdatingId(id);
    const res = await fetch(`/api/admin/reports/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: newStatus }),
    });
    if (res.ok) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
    setUpdatingId(null);
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        {["pending", "reviewed", "dismissed"].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
              statusFilter === s
                ? "bg-accent text-white"
                : "bg-surface-2 text-text-primary hover:bg-surface-2"
            }`}
          >
            {s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-sm text-text-primary">Loading...</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-text-primary">
          No {statusFilter} reports.
        </p>
      ) : (
        <div className="space-y-4">
          {items.map((item) => (
            <div
              key={item.id}
              className="border border-border rounded-lg p-4"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${TYPE_CLASSES[item.content_type] ?? ""}`}
                    >
                      {item.content_type}
                    </span>
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${REASON_CLASSES[item.reason] ?? ""}`}
                    >
                      {item.reason}
                    </span>
                  </div>
                  <p className="text-xs text-text-secondary mt-1">
                    Reported by{" "}
                    <a
                      href={`/${item.reporter_username}`}
                      className="hover:underline font-medium"
                    >
                      {item.reporter_display_name ?? item.reporter_username}
                    </a>{" "}
                    &middot; {formatDate(item.created_at)}
                  </p>
                  {item.details && (
                    <p className="text-sm text-text-secondary mt-2">
                      <span className="font-medium text-text-primary">Details: </span>
                      {item.details}
                    </p>
                  )}
                </div>

                {statusFilter === "pending" && (
                  <div className="shrink-0 flex items-center gap-2">
                    <button
                      onClick={() => updateStatus(item.id, "reviewed")}
                      disabled={updatingId === item.id}
                      className="px-3 py-1 rounded text-xs font-medium bg-green-900/30 text-green-300 hover:bg-green-900/50 transition-colors disabled:opacity-50"
                    >
                      Review
                    </button>
                    <button
                      onClick={() => updateStatus(item.id, "dismissed")}
                      disabled={updatingId === item.id}
                      className="px-3 py-1 rounded text-xs font-medium bg-surface-2 text-text-primary hover:bg-surface-3 transition-colors disabled:opacity-50"
                    >
                      Dismiss
                    </button>
                  </div>
                )}

                {statusFilter !== "pending" && item.reviewer_username && (
                  <div className="shrink-0 text-xs text-text-primary">
                    {item.status === "reviewed" ? "Reviewed" : "Dismissed"} by{" "}
                    {item.reviewer_username}
                  </div>
                )}
              </div>

              {/* Content preview */}
              <div className="mt-3 p-3 bg-surface-2 rounded text-sm text-text-secondary">
                <p className="text-xs font-medium text-text-primary mb-1">
                  Reported content preview
                </p>
                <p className="text-xs whitespace-pre-wrap line-clamp-4">
                  {item.content_preview}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
