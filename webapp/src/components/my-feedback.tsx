"use client";

import { useState, useEffect, useCallback } from "react";

type FeedbackItem = {
  id: string;
  type: string;
  title: string;
  description: string;
  steps_to_reproduce: string | null;
  severity: string | null;
  status: string;
  created_at: string;
};

const SEVERITY_CLASSES: Record<string, string> = {
  low: "bg-semantic-info-bg text-semantic-info border-semantic-info-border",
  medium: "bg-semantic-warning-bg text-semantic-warning border-semantic-warning-border",
  high: "bg-semantic-error-bg text-semantic-error border-semantic-error-border",
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

export default function MyFeedback() {
  const [items, setItems] = useState<FeedbackItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchFeedback = useCallback(async () => {
    setLoading(true);
    const res = await fetch("/api/me/feedback");
    if (res.ok) {
      setItems(await res.json());
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchFeedback();
  }, [fetchFeedback]);

  async function deleteFeedback(id: string) {
    if (!confirm("Delete this ticket? This cannot be undone.")) return;
    setDeletingId(id);
    const res = await fetch(`/api/me/feedback/${id}`, {
      method: "DELETE",
    });
    if (res.ok) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
    setDeletingId(null);
  }

  if (loading) {
    return <p className="text-sm text-text-secondary">Loading your tickets...</p>;
  }

  if (items.length === 0) {
    return null;
  }

  return (
    <div>
      <h2 className="text-lg font-semibold text-text-primary mb-4">
        Your Tickets
      </h2>
      <div className="space-y-3">
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
                  <span
                    className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      item.status === "open"
                        ? "bg-semantic-success-bg text-semantic-success"
                        : "bg-surface-2 text-text-tertiary"
                    }`}
                  >
                    {item.status}
                  </span>
                </div>
                <p className="text-sm font-medium text-text-primary">
                  {item.title}
                </p>
                <p className="text-xs text-text-secondary mt-0.5">
                  {formatDate(item.created_at)}
                </p>
              </div>

              <button
                onClick={() => deleteFeedback(item.id)}
                disabled={deletingId === item.id}
                className="px-3 py-1 rounded text-xs font-medium transition-colors disabled:opacity-50 bg-danger-bg text-danger hover:bg-danger-bg-hover shrink-0"
              >
                Delete
              </button>
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
    </div>
  );
}
