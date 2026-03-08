"use client";

import { useState, useEffect, useCallback } from "react";

type LinkEdit = {
  id: string;
  book_link_id: string;
  username: string;
  display_name: string | null;
  proposed_type: string | null;
  proposed_note: string | null;
  current_type: string;
  current_note: string | null;
  from_book_ol_id: string;
  from_book_title: string;
  to_book_ol_id: string;
  to_book_title: string;
  status: string;
  reviewer_name: string | null;
  reviewer_comment: string | null;
  created_at: string;
  reviewed_at: string | null;
};

const TYPE_LABELS: Record<string, string> = {
  sequel: "Sequel",
  prequel: "Prequel",
  companion: "Companion",
  similar: "Similar",
  mentioned_in: "Mentioned in",
  adaptation: "Adaptation",
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

export default function AdminLinkEdits() {
  const [statusFilter, setStatusFilter] = useState("pending");
  const [edits, setEdits] = useState<LinkEdit[]>([]);
  const [loading, setLoading] = useState(true);
  const [reviewingId, setReviewingId] = useState<string | null>(null);

  const fetchEdits = useCallback(async (status: string) => {
    setLoading(true);
    const res = await fetch(`/api/admin/link-edits?status=${status}`);
    if (res.ok) {
      setEdits(await res.json());
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchEdits(statusFilter);
  }, [statusFilter, fetchEdits]);

  async function handleReview(editId: string, action: "approve" | "reject") {
    setReviewingId(editId);
    const res = await fetch(`/api/admin/link-edits/${editId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action }),
    });
    if (res.ok) {
      setEdits((prev) => prev.filter((e) => e.id !== editId));
    }
    setReviewingId(null);
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        {["pending", "approved", "rejected"].map((s) => (
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
      ) : edits.length === 0 ? (
        <p className="text-sm text-text-primary">
          No {statusFilter} link edits.
        </p>
      ) : (
        <div className="space-y-4">
          {edits.map((edit) => (
            <div
              key={edit.id}
              className="border border-border rounded-lg p-4"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-sm text-text-primary">
                    <span className="font-medium">
                      {edit.display_name ?? edit.username}
                    </span>{" "}
                    proposed an edit on the link between{" "}
                    <a
                      href={`/books/${edit.from_book_ol_id}`}
                      className="font-medium hover:underline"
                    >
                      {edit.from_book_title}
                    </a>{" "}
                    and{" "}
                    <a
                      href={`/books/${edit.to_book_ol_id}`}
                      className="font-medium hover:underline"
                    >
                      {edit.to_book_title}
                    </a>
                  </p>
                  <p className="text-xs text-text-primary mt-1">
                    {formatDate(edit.created_at)}
                  </p>
                </div>

                {statusFilter === "pending" && (
                  <div className="shrink-0 flex items-center gap-2">
                    <button
                      onClick={() => handleReview(edit.id, "approve")}
                      disabled={reviewingId === edit.id}
                      className="px-3 py-1 rounded text-xs font-medium bg-semantic-success-bg text-semantic-success hover:bg-semantic-success-bg/70 transition-colors disabled:opacity-50"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => handleReview(edit.id, "reject")}
                      disabled={reviewingId === edit.id}
                      className="px-3 py-1 rounded text-xs font-medium bg-semantic-error-bg text-semantic-error hover:bg-semantic-error-bg/70 transition-colors disabled:opacity-50"
                    >
                      Reject
                    </button>
                  </div>
                )}
              </div>

              <div className="mt-3 grid grid-cols-2 gap-4 text-xs">
                <div>
                  <p className="font-medium text-text-primary mb-1">Current</p>
                  <p className="text-text-primary">
                    Type: {TYPE_LABELS[edit.current_type] ?? edit.current_type}
                  </p>
                  <p className="text-text-primary">
                    Note: {edit.current_note ?? <span className="text-text-primary italic">none</span>}
                  </p>
                </div>
                <div>
                  <p className="font-medium text-text-primary mb-1">Proposed</p>
                  <p className="text-text-primary">
                    Type:{" "}
                    {edit.proposed_type ? (
                      <span className="font-medium text-semantic-info">
                        {TYPE_LABELS[edit.proposed_type] ?? edit.proposed_type}
                      </span>
                    ) : (
                      <span className="text-text-primary italic">no change</span>
                    )}
                  </p>
                  <p className="text-text-primary">
                    Note:{" "}
                    {edit.proposed_note != null ? (
                      <span className="font-medium text-semantic-info">
                        {edit.proposed_note || <span className="italic">empty</span>}
                      </span>
                    ) : (
                      <span className="text-text-primary italic">no change</span>
                    )}
                  </p>
                </div>
              </div>

              {edit.reviewer_name && (
                <p className="text-xs text-text-primary mt-2">
                  Reviewed by {edit.reviewer_name}
                  {edit.reviewed_at && ` on ${formatDate(edit.reviewed_at)}`}
                  {edit.reviewer_comment && ` — ${edit.reviewer_comment}`}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
