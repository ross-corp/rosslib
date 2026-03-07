"use client";

import Link from "next/link";
import { Suspense, useState } from "react";
import { useSearchParams } from "next/navigation";

function ResetPasswordContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");

    const form = e.currentTarget;
    const newPassword = (
      form.elements.namedItem("password") as HTMLInputElement
    ).value;
    const confirmPassword = (
      form.elements.namedItem("confirm_password") as HTMLInputElement
    ).value;

    if (newPassword !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    if (newPassword.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);

    const res = await fetch("/api/auth/reset-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token, new_password: newPassword }),
    });

    setLoading(false);

    if (!res.ok) {
      const body = await res.json();
      setError(body.error || "Something went wrong.");
      return;
    }

    setSuccess(true);
  }

  if (!token) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center">
        <div className="w-full max-w-sm text-center">
          <h1 className="text-2xl font-bold text-text-primary">
            Invalid reset link
          </h1>
          <p className="mt-2 text-sm text-text-secondary">
            This password reset link is invalid. Please request a new one.
          </p>
          <Link
            href="/forgot-password"
            className="inline-block mt-4 text-text-primary font-medium hover:underline text-sm"
          >
            Request new reset link
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-text-primary">
            Set new password
          </h1>
        </div>

        {success ? (
          <div className="text-sm text-text-secondary panel px-4 py-3">
            <p className="font-medium mb-1 text-text-primary">Password reset successful</p>
            <p>Your password has been updated. You can now sign in.</p>
            <Link
              href="/login"
              className="inline-block mt-4 text-text-primary font-medium hover:underline"
            >
              Sign in
            </Link>
          </div>
        ) : (
          <form className="space-y-4" onSubmit={handleSubmit}>
            {error && (
              <p role="alert" aria-live="assertive" className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded px-3 py-2">
                {error}
              </p>
            )}

            <div>
              <label
                htmlFor="password"
                className="label-mono block mb-1"
              >
                New password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                className="input-field"
                placeholder="••••••••"
              />
            </div>

            <div>
              <label
                htmlFor="confirm_password"
                className="label-mono block mb-1"
              >
                Confirm password
              </label>
              <input
                id="confirm_password"
                name="confirm_password"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                className="input-field"
                placeholder="••••••••"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full btn-primary py-2.5"
            >
              {loading ? "Resetting..." : "Reset password"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense>
      <ResetPasswordContent />
    </Suspense>
  );
}
