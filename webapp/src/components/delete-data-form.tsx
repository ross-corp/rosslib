"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function DeleteDataForm() {
  const router = useRouter();
  const [showConfirm, setShowConfirm] = useState(false);
  const [confirmText, setConfirmText] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleDelete() {
    setError("");
    setLoading(true);

    const res = await fetch("/api/me/account/data", { method: "DELETE" });

    setLoading(false);

    if (!res.ok) {
      const data = await res.json();
      setError(data.error || "Something went wrong.");
      return;
    }

    router.push("/");
    router.refresh();
  }

  return (
    <div className="border-t border-semantic-error-border pt-8 mt-8">
      <h2 className="text-lg font-bold text-semantic-error mb-1">Danger zone</h2>
      <p className="text-sm text-text-primary mb-5">
        Permanently delete all your books, reviews, tags, shelves, follows, threads, and notifications.
        Your account will remain but all data will be removed. This cannot be undone.
      </p>

      {error && (
        <p className="text-sm text-semantic-error bg-semantic-error-bg border border-semantic-error-border rounded px-3 py-2 mb-4">
          {error}
        </p>
      )}

      {!showConfirm ? (
        <button
          onClick={() => setShowConfirm(true)}
          className="bg-danger text-badge-text px-4 py-2 rounded text-sm font-medium hover:bg-danger transition-colors"
        >
          Delete all my data
        </button>
      ) : (
        <div className="space-y-3 max-w-md">
          <p className="text-sm text-text-primary">
            This will remove all your books, reviews, tags, and follows. This cannot be undone.
          </p>
          <p className="text-sm text-text-primary">
            Type <span className="font-mono font-bold">delete my data</span> to confirm:
          </p>
          <input
            type="text"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            placeholder="delete my data"
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-semantic-error focus:border-transparent text-sm"
          />
          <div className="flex gap-3">
            <button
              onClick={handleDelete}
              disabled={loading || confirmText !== "delete my data"}
              className="bg-danger text-badge-text px-4 py-2 rounded text-sm font-medium hover:bg-danger transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Deleting..." : "Permanently delete all data"}
            </button>
            <button
              onClick={() => {
                setShowConfirm(false);
                setConfirmText("");
              }}
              className="px-4 py-2 rounded text-sm font-medium border border-border text-text-primary hover:bg-surface-2 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
