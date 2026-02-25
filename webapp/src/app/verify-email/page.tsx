"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get("token");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token) return;

    setLoading(true);
    fetch("/api/auth/verify-email", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token }),
    })
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) {
          setError(data.error || "Something went wrong.");
        } else {
          setSuccess(true);
        }
      })
      .catch(() => setError("Something went wrong."))
      .finally(() => setLoading(false));
  }, [token]);

  if (!token) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center px-4">
        <div className="w-full max-w-sm text-center">
          <Link href="/" className="font-semibold text-stone-900 text-xl">
            rosslib
          </Link>
          <h1 className="mt-4 text-2xl font-bold text-stone-900">
            Invalid verification link
          </h1>
          <p className="mt-2 text-sm text-stone-500">
            This verification link is invalid. You can request a new one from
            your settings page.
          </p>
          <Link
            href="/settings"
            className="inline-block mt-4 text-stone-900 font-medium hover:underline text-sm"
          >
            Go to settings
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4">
      <div className="w-full max-w-sm text-center">
        <Link href="/" className="font-semibold text-stone-900 text-xl">
          rosslib
        </Link>

        {loading && (
          <p className="mt-6 text-sm text-stone-500">Verifying your email...</p>
        )}

        {success && (
          <div className="mt-6">
            <h1 className="text-2xl font-bold text-stone-900">
              Email verified
            </h1>
            <p className="mt-2 text-sm text-stone-500">
              Your email has been verified. You now have full access.
            </p>
            <button
              onClick={() => {
                router.push("/");
                router.refresh();
              }}
              className="inline-block mt-4 bg-stone-900 text-white px-4 py-2 rounded text-sm font-medium hover:bg-stone-700 transition-colors"
            >
              Go to home
            </button>
          </div>
        )}

        {error && (
          <div className="mt-6">
            <h1 className="text-2xl font-bold text-stone-900">
              Verification failed
            </h1>
            <p className="mt-2 text-sm text-red-600">{error}</p>
            <Link
              href="/settings"
              className="inline-block mt-4 text-stone-900 font-medium hover:underline text-sm"
            >
              Request a new verification email
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}
