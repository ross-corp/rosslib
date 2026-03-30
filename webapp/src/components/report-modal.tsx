"use client";

import { useState, useEffect, useRef, useCallback } from "react";

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
  const dialogRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<Element | null>(null);

  const handleClose = useCallback(() => {
    if (previousFocusRef.current instanceof HTMLElement) {
      previousFocusRef.current.focus();
    }
    onClose();
  }, [onClose]);

  useEffect(() => {
    previousFocusRef.current = document.activeElement;

    // Focus the first radio button on open
    const timer = setTimeout(() => {
      const first = dialogRef.current?.querySelector<HTMLElement>(
        'input[type="radio"]'
      );
      first?.focus();
    }, 0);

    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        e.stopPropagation();
        handleClose();
        return;
      }

      if (e.key === "Tab") {
        const dialog = dialogRef.current;
        if (!dialog) return;

        const focusable = dialog.querySelectorAll<HTMLElement>(
          'input, textarea, button, [tabindex]:not([tabindex="-1"])'
        );
        if (focusable.length === 0) return;

        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault();
            last.focus();
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault();
            first.focus();
          }
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown, true);
    return () => document.removeEventListener("keydown", handleKeyDown, true);
  }, [handleClose]);

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
    setTimeout(handleClose, 1500);
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-overlay"
      onClick={(e) => {
        if (e.target === e.currentTarget) handleClose();
      }}
    >
      <div ref={dialogRef} role="dialog" aria-modal="true" aria-labelledby="report-modal-title" className="bg-surface-0 border border-border rounded-lg p-6 w-full max-w-md mx-4 shadow-lg">
        {success ? (
          <div className="text-center py-4">
            <p className="text-sm text-semantic-success font-medium">
              Report submitted. Thank you.
            </p>
          </div>
        ) : (
          <>
            <h3 id="report-modal-title" className="text-sm font-semibold text-text-primary mb-4">
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
                  onClick={handleClose}
                  disabled={submitting}
                  aria-label="Close report dialog"
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
