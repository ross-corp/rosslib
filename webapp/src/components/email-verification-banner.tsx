"use client";

import { useEffect, useState } from "react";

export default function EmailVerificationBanner() {
  const [emailVerified, setEmailVerified] = useState<boolean | null>(null);
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);

  useEffect(() => {
    fetch("/api/me/account")
      .then((res) => res.json())
      .then((data) => {
        setEmailVerified(data.email_verified ?? true);
      });
  }, []);

  if (emailVerified === null || emailVerified) return null;

  async function handleResend() {
    setSending(true);
    setSent(false);
    await fetch("/api/auth/resend-verification", { method: "POST" });
    setSending(false);
    setSent(true);
  }

  return (
    <div className="bg-amber-50 border border-amber-200 rounded px-4 py-3 mb-6">
      <p className="text-sm text-amber-800 font-medium">
        Your email is not verified.
      </p>
      <p className="text-sm text-amber-700 mt-1">
        Please check your inbox for a verification link. Some features may be
        limited until your email is verified.
      </p>
      <div className="mt-2">
        {sent ? (
          <p className="text-sm text-green-700">
            Verification email sent. Check your inbox.
          </p>
        ) : (
          <button
            onClick={handleResend}
            disabled={sending}
            className="text-sm font-medium text-amber-900 hover:underline disabled:opacity-50"
          >
            {sending ? "Sending..." : "Resend verification email"}
          </button>
        )}
      </div>
    </div>
  );
}
