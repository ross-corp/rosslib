"use client";

import Link from "next/link";
import { Suspense, useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";

function VerifyEmailContent() {
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
      <div className="min-h-[70vh] flex flex-col items-center justify-center">
        <div className="w-full max-w-sm text-center">
          <h1 className="text-2xl font-bold text-text-primary">
            Invalid verification link
          </h1>
          <p className="mt-2 text-sm text-text-secondary">
            This verification link is invalid. You can request a new one from
            your settings page.
          </p>
          <Link
            href="/settings"
            className="inline-block mt-4 text-text-primary font-medium hover:underline text-sm"
          >
            Go to settings
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center">
      <div className="w-full max-w-sm text-center">
        {loading && (
          <p className="text-sm text-text-secondary">Verifying your email...</p>
        )}

        {success && (
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              Email verified
            </h1>
            <p className="mt-2 text-sm text-text-secondary">
              Your email has been verified. You now have full access.
            </p>
            <button
              onClick={() => {
                router.push("/");
                router.refresh();
              }}
              className="inline-block mt-4 btn-primary"
            >
              Go to home
            </button>
          </div>
        )}

        {error && (
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              Verification failed
            </h1>
            <p className="mt-2 text-sm text-semantic-error">{error}</p>
            <Link
              href="/settings"
              className="inline-block mt-4 text-text-primary font-medium hover:underline text-sm"
            >
              Request a new verification email
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmailContent />
    </Suspense>
  );
}
