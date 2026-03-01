"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function DeleteDataForm() {
  const router = useRouter();
  const [showConfirm, setShowConfirm] = useState(false);
  const [confirmText, setConfirmText] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [showAccountConfirm, setShowAccountConfirm] = useState(false);
  const [accountConfirmText, setAccountConfirmText] = useState("");
  const [accountLoading, setAccountLoading] = useState(false);
  const [accountError, setAccountError] = useState("");

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

  async function handleDeleteAccount() {
    setAccountError("");
    setAccountLoading(true);

    const res = await fetch("/api/me/account", { method: "DELETE" });

    setAccountLoading(false);

    if (!res.ok) {
      const data = await res.json();
      setAccountError(data.error || "Something went wrong.");
      return;
    }

    router.push("/");
    router.refresh();
  }

  return (
    <div className="border-t border-red-300 pt-8 mt-8">
      <h2 className="text-lg font-bold text-red-600 mb-1">Danger zone</h2>

      {/* Delete all data */}
      <p className="text-sm text-text-primary mb-5">
        Permanently delete all your books, reviews, tags, shelves, follows, threads, and notifications.
        Your account will remain but all data will be removed. This cannot be undone.
      </p>

      {error && (
        <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2 mb-4">
          {error}
        </p>
      )}

      {!showConfirm ? (
        <button
          onClick={() => setShowConfirm(true)}
          className="bg-red-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-red-700 transition-colors"
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
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent text-sm"
          />
          <div className="flex gap-3">
            <button
              onClick={handleDelete}
              disabled={loading || confirmText !== "delete my data"}
              className="bg-red-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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

      {/* Delete account */}
      <div className="mt-8 pt-6 border-t border-red-200">
        <p className="text-sm text-text-primary mb-5">
          Permanently delete your account and all associated data. This removes everything
          including your profile, books, reviews, and follows. This cannot be undone.
        </p>

        {accountError && (
          <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2 mb-4">
            {accountError}
          </p>
        )}

        {!showAccountConfirm ? (
          <button
            onClick={() => setShowAccountConfirm(true)}
            className="bg-red-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-red-700 transition-colors"
          >
            Delete my account permanently
          </button>
        ) : (
          <div className="space-y-3 max-w-md">
            <p className="text-sm text-text-primary">
              This will permanently delete your account and all data. You will be logged out
              and will not be able to recover your account. This cannot be undone.
            </p>
            <p className="text-sm text-text-primary">
              Type <span className="font-mono font-bold">delete my account</span> to confirm:
            </p>
            <input
              type="text"
              value={accountConfirmText}
              onChange={(e) => setAccountConfirmText(e.target.value)}
              placeholder="delete my account"
              className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent text-sm"
            />
            <div className="flex gap-3">
              <button
                onClick={handleDeleteAccount}
                disabled={accountLoading || accountConfirmText !== "delete my account"}
                className="bg-red-600 text-white px-4 py-2 rounded text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {accountLoading ? "Deleting..." : "Permanently delete my account"}
              </button>
              <button
                onClick={() => {
                  setShowAccountConfirm(false);
                  setAccountConfirmText("");
                }}
                className="px-4 py-2 rounded text-sm font-medium border border-border text-text-primary hover:bg-surface-2 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
