"use client";

import { useEffect, useState } from "react";

export default function PasswordForm() {
  const [hasPassword, setHasPassword] = useState<boolean | null>(null);
  const [hasGoogle, setHasGoogle] = useState(false);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetch("/api/me/account")
      .then((res) => res.json())
      .then((data) => {
        setHasPassword(data.has_password ?? false);
        setHasGoogle(data.has_google ?? false);
      });
  }, []);

  if (hasPassword === null) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaved(false);

    if (newPassword.length < 8) {
      setError("New password must be at least 8 characters.");
      return;
    }

    if (newPassword !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setLoading(true);

    const body: Record<string, string> = { new_password: newPassword };
    if (hasPassword) {
      body.current_password = currentPassword;
    }

    const res = await fetch("/api/me/password", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    setLoading(false);

    if (!res.ok) {
      const data = await res.json();
      setError(data.error || "Something went wrong.");
      return;
    }

    setSaved(true);
    setCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");
    setHasPassword(true);
  }

  return (
    <div className="border-t border-border pt-8 mt-8">
      <h2 className="text-lg font-bold text-text-primary mb-1">Password</h2>
      <p className="text-sm text-text-primary mb-5">
        {hasPassword
          ? "Change your password."
          : hasGoogle
            ? "Set a password so you can also sign in with email and password."
            : "Set a password for your account."}
      </p>

      <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
        {error && (
          <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">
            {error}
          </p>
        )}
        {saved && (
          <p className="text-sm text-green-700 bg-green-50 border border-green-200 rounded px-3 py-2">
            {hasPassword ? "Password updated." : "Password set. You can now sign in with email and password."}
          </p>
        )}

        {hasPassword && (
          <div>
            <label
              htmlFor="current_password"
              className="block text-sm font-medium text-text-primary mb-1"
            >
              Current password
            </label>
            <input
              id="current_password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
              className="w-full px-3 py-2 border border-border rounded text-text-primary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
            />
          </div>
        )}

        <div>
          <label
            htmlFor="new_password"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            {hasPassword ? "New password" : "Password"}
          </label>
          <input
            id="new_password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={8}
            required
            placeholder="At least 8 characters"
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>

        <div>
          <label
            htmlFor="confirm_password"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            Confirm {hasPassword ? "new " : ""}password
          </label>
          <input
            id="confirm_password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            minLength={8}
            required
            className="w-full px-3 py-2 border border-border rounded text-text-primary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>

        <div className="pt-1">
          <button
            type="submit"
            disabled={loading}
            className="bg-accent text-white px-4 py-2 rounded text-sm font-medium hover:bg-surface-3 transition-colors disabled:opacity-50"
          >
            {loading
              ? "Saving..."
              : hasPassword
                ? "Change password"
                : "Set password"}
          </button>
        </div>
      </form>
    </div>
  );
}
