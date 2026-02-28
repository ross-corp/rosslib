"use client";

import { useState, useEffect, useCallback } from "react";

type FeedbackItem = {
  id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  type: string;
  title: string;
  description: string;
  steps_to_reproduce: string | null;
  severity: string | null;
  status: string;
  created_at: string;
};

const SEVERITY_CLASSES: Record<string, string> = {
  low: "bg-blue-900/30 text-blue-300 border-blue-700",
  medium: "bg-yellow-900/30 text-yellow-300 border-yellow-700",
  high: "bg-red-900/30 text-red-300 border-red-700",
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

export default function AdminFeedback() {
  const [statusFilter, setStatusFilter] = useState("open");
  const [items, setItems] = useState<FeedbackItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchFeedback = useCallback(async (status: string) => {
    setLoading(true);
    const res = await fetch(`/api/admin/feedback?status=${status}`);
    if (res.ok) {
      setItems(await res.json());
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchFeedback(statusFilter);
  }, [statusFilter, fetchFeedback]);

  async function toggleStatus(id: string, currentStatus: string) {
    const newStatus = currentStatus === "open" ? "closed" : "open";
    setTogglingId(id);
    const res = await fetch(`/api/admin/feedback/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: newStatus }),
    });
    if (res.ok) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
    setTogglingId(null);
  }

  async function deleteFeedback(id: string) {
    if (!confirm("Delete this feedback? This cannot be undone.")) return;
    setDeletingId(id);
    const res = await fetch(`/api/admin/feedback/${id}`, {
      method: "DELETE",
    });
    if (res.ok) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
    setDeletingId(null);
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        {["open", "closed"].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
              statusFilter === s
                ? "bg-accent text-text-inverted"
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
          No {statusFilter} feedback.
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
                  <div className="flex items-center gap-2 mb-1">
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                        item.type === "bug"
                          ? "bg-red-900/30 text-red-300"
                          : "bg-purple-900/30 text-purple-300"
                      }`}
                    >
                      {item.type === "bug" ? "Bug" : "Feature"}
                    </span>
                    {item.severity && (
                      <span
                        className={`inline-block px-2 py-0.5 rounded border text-xs font-medium ${SEVERITY_CLASSES[item.severity] ?? ""}`}
                      >
                        {item.severity}
                      </span>
                    )}
                  </div>
                  <p className="text-sm font-medium text-text-primary">
                    {item.title}
                  </p>
                  <p className="text-xs text-text-secondary mt-0.5">
                    by{" "}
                    <a
                      href={`/${item.username}`}
                      className="hover:underline"
                    >
                      {item.display_name ?? item.username}
                    </a>{" "}
                    &middot; {formatDate(item.created_at)}
                  </p>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <button
                    onClick={() => toggleStatus(item.id, item.status)}
                    disabled={togglingId === item.id}
                    className="px-3 py-1 rounded text-xs font-medium transition-colors disabled:opacity-50 bg-surface-2 text-text-primary hover:bg-surface-3"
                  >
                    {item.status === "open" ? "Close" : "Reopen"}
                  </button>
                  <button
                    onClick={() => deleteFeedback(item.id)}
                    disabled={deletingId === item.id}
                    className="px-3 py-1 rounded text-xs font-medium transition-colors disabled:opacity-50 bg-red-900/30 text-red-300 hover:bg-red-900/50"
                  >
                    Delete
                  </button>
                </div>
              </div>

              <p className="text-sm text-text-secondary mt-3 whitespace-pre-wrap">
                {item.description}
              </p>

              {item.steps_to_reproduce && (
                <div className="mt-3">
                  <p className="text-xs font-medium text-text-primary mb-1">
                    Steps to Reproduce
                  </p>
                  <p className="text-xs text-text-secondary whitespace-pre-wrap">
                    {item.steps_to_reproduce}
                  </p>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
