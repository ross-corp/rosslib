"use client";

import { useState } from "react";

type Props = {
  contentType: "review" | "thread" | "comment" | "link";
  contentId: string;
  onClose: () => void;
};

const REASONS = [
  { value: "spam", label: "Spam" },
  { value: "harassment", label: "Harassment" },
  { value: "inappropriate", label: "Inappropriate content" },
  { value: "other", label: "Other" },
];

export default function ReportModal({ contentType, contentId, onClose }: Props) {
  const [reason, setReason] = useState("");
  const [details, setDetails] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!reason) return;

    setSubmitting(true);
    setError(null);

    const res = await fetch("/api/reports", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        content_type: contentType,
        content_id: contentId,
        reason,
        details: details.trim() || undefined,
      }),
    });

    setSubmitting(false);

    if (!res.ok) {
      const data = await res.json().catch(() => null);
      setError(data?.error ?? "Failed to submit report");
      return;
    }

    setSuccess(true);
    setTimeout(onClose, 1500);
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-overlay"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-surface-0 border border-border rounded-lg p-6 w-full max-w-md mx-4 shadow-lg">
        {success ? (
          <div className="text-center py-4">
            <p className="text-sm text-semantic-success font-medium">
              Report submitted. Thank you.
            </p>
          </div>
        ) : (
          <>
            <h3 className="text-sm font-semibold text-text-primary mb-4">
              Report {contentType}
            </h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs text-text-primary mb-2">
                  Reason
                </label>
                <div className="space-y-2">
                  {REASONS.map((r) => (
                    <label
                      key={r.value}
                      className="flex items-center gap-2 text-sm text-text-primary cursor-pointer"
                    >
                      <input
                        type="radio"
                        name="reason"
                        value={r.value}
                        checked={reason === r.value}
                        onChange={() => setReason(r.value)}
                        disabled={submitting}
                        className="accent-accent"
                      />
                      {r.label}
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-xs text-text-primary mb-1">
                  Details (optional)
                </label>
                <textarea
                  value={details}
                  onChange={(e) => setDetails(e.target.value)}
                  placeholder="Provide additional context..."
                  rows={3}
                  disabled={submitting}
                  className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
                />
              </div>

              {error && (
                <p className="text-xs text-semantic-error">{error}</p>
              )}

              <div className="flex items-center gap-3">
                <button
                  type="submit"
                  disabled={submitting || !reason}
                  className="text-xs px-4 py-2 rounded bg-danger text-badge-text hover:bg-danger-bg-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {submitting ? "Submitting..." : "Submit report"}
                </button>
                <button
                  type="button"
                  onClick={onClose}
                  disabled={submitting}
                  className="text-xs text-text-primary hover:text-text-primary transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
