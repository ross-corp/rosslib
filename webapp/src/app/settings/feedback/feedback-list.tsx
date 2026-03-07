"use client";

import { useState } from "react";

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

export default function FeedbackList({
  initialItems,
}: {
  initialItems: FeedbackItem[];
}) {
  const [items, setItems] = useState(initialItems);
  const [deleting, setDeleting] = useState<Record<string, boolean>>({});

  async function handleDelete(id: string) {
    setDeleting((prev) => ({ ...prev, [id]: true }));
    const res = await fetch(`/api/me/feedback/${id}`, { method: "DELETE" });
    if (res.ok) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
    setDeleting((prev) => ({ ...prev, [id]: false }));
  }

  if (items.length === 0) {
    return (
      <p className="text-sm text-text-primary">
        You haven&apos;t submitted any feedback yet.
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {items.map((item) => (
        <div
          key={item.id}
          className="flex items-start justify-between py-3 border-b border-border gap-4"
        >
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 mb-1">
              <span
                className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                  item.type === "bug"
                    ? "bg-red-100 text-red-700"
                    : "bg-blue-100 text-blue-700"
                }`}
              >
                {item.type === "bug" ? "Bug" : "Feature"}
              </span>
              <span
                className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                  item.status === "open"
                    ? "bg-green-100 text-green-700"
                    : "bg-gray-100 text-gray-600"
                }`}
              >
                {item.status === "open" ? "Open" : "Closed"}
              </span>
              {item.severity && (
                <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-700">
                  {item.severity}
                </span>
              )}
            </div>
            <h3 className="text-sm font-medium text-text-primary truncate">
              {item.title}
            </h3>
            <p className="text-xs text-text-primary mt-0.5 line-clamp-2">
              {item.description}
            </p>
            <p className="text-xs text-text-primary mt-1">
              {new Date(item.created_at).toLocaleDateString(undefined, {
                year: "numeric",
                month: "short",
                day: "numeric",
              })}
            </p>
          </div>
          <button
            onClick={() => handleDelete(item.id)}
            disabled={deleting[item.id]}
            className="text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-red-300 hover:text-red-600 transition-colors disabled:opacity-50 shrink-0"
          >
            Delete
          </button>
        </div>
      ))}
    </div>
  );
}
