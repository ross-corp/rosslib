"use client";

import { useEffect, useState } from "react";

export default function EmailForm() {
  const [currentEmail, setCurrentEmail] = useState("");
  const [newEmail, setNewEmail] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetch("/api/me/account")
      .then((res) => res.json())
      .then((data) => {
        setCurrentEmail(data.email ?? "");
      });
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaved(false);

    if (!newEmail || !currentPassword) {
      setError("New email and current password are required.");
      return;
    }

    setLoading(true);

    const res = await fetch("/api/me/email", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        new_email: newEmail,
        current_password: currentPassword,
      }),
    });

    setLoading(false);

    if (!res.ok) {
      const data = await res.json();
      setError(data.error || "Something went wrong.");
      return;
    }

    setSaved(true);
    setCurrentEmail(newEmail);
    setNewEmail("");
    setCurrentPassword("");
  }

  return (
    <div className="border-t border-border pt-8 mt-8">
      <h2 className="text-lg font-bold text-text-primary mb-1">Email</h2>
      <p className="text-sm text-text-primary mb-5">
        Change your email address. You&apos;ll need to verify the new address.
      </p>

      {currentEmail && (
        <p className="text-sm text-text-secondary mb-4">
          Current email: <span className="font-medium text-text-primary">{currentEmail}</span>
        </p>
      )}

      <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
        {error && (
          <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">
            {error}
          </p>
        )}
        {saved && (
          <p className="text-sm text-green-700 bg-green-50 border border-green-200 rounded px-3 py-2">
            Email updated. Please check your inbox to verify your new address.
          </p>
        )}

        <div>
          <label
            htmlFor="new_email"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            New email
          </label>
          <input
            id="new_email"
            type="email"
            value={newEmail}
            onChange={(e) => setNewEmail(e.target.value)}
            required
            placeholder="you@example.com"
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>

        <div>
          <label
            htmlFor="email_current_password"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            Current password
          </label>
          <input
            id="email_current_password"
            type="password"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
            className="w-full px-3 py-2 border border-border rounded text-text-primary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>

        <div className="pt-1">
          <button
            type="submit"
            disabled={loading}
            className="bg-accent text-text-inverted px-4 py-2 rounded text-sm font-medium hover:bg-surface-3 transition-colors disabled:opacity-50"
          >
            {loading ? "Saving..." : "Change email"}
          </button>
        </div>
      </form>
    </div>
  );
}
