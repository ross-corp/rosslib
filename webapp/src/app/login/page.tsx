"use client";

import Link from "next/link";
import { Suspense, useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

const googleErrorMessages: Record<string, string> = {
  google_denied: "Google sign-in was cancelled.",
  google_not_configured: "Google sign-in is not configured.",
  google_token_failed: "Failed to authenticate with Google. Please try again.",
  google_userinfo_failed: "Could not retrieve your Google account info.",
  google_login_failed: "Failed to sign in with Google. Please try again.",
};

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const googleEnabled = !!process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;

  useEffect(() => {
    const googleError = searchParams.get("error");
    if (googleError && googleErrorMessages[googleError]) {
      setError(googleErrorMessages[googleError]);
    }
  }, [searchParams]);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setLoading(true);

    const form = e.currentTarget;
    const data = {
      email: (form.elements.namedItem("email") as HTMLInputElement).value,
      password: (form.elements.namedItem("password") as HTMLInputElement).value,
    };

    const res = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });

    setLoading(false);

    if (!res.ok) {
      const body = await res.json();
      setError(body.error || "Something went wrong.");
      return;
    }

    router.push("/");
    router.refresh();
  }

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-text-primary">
            Sign in to your account
          </h1>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          {error && (
            <p className="text-sm text-semantic-error bg-semantic-error-bg border border-semantic-error-border rounded px-3 py-2">
              {error}
            </p>
          )}

          <div>
            <label
              htmlFor="email"
              className="label-mono block mb-1"
            >
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              required
              className="input-field"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label
                htmlFor="password"
                className="label-mono"
              >
                Password
              </label>
              <Link
                href="/forgot-password"
                className="text-xs text-text-tertiary hover:text-text-primary"
              >
                Forgot password?
              </Link>
            </div>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              className="input-field"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full btn-primary py-2.5"
          >
            {loading ? "Signing in..." : "Sign in"}
          </button>
        </form>

        {googleEnabled && (
          <>
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="divider w-full" />
              </div>
              <div className="relative flex justify-center text-xs">
                <span className="bg-surface-0 px-2 text-text-tertiary">or</span>
              </div>
            </div>

            {/* eslint-disable-next-line @next/next/no-html-link-for-pages */}
            <a
              href="/api/auth/google"
              className="w-full flex items-center justify-center gap-2 border border-border rounded py-2.5 text-sm font-medium text-text-secondary hover:bg-surface-2 transition-colors"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24">
                <path
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
                  fill="#4285F4"
                />
                <path
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  fill="#34A853"
                />
                <path
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18A10.96 10.96 0 0 0 1 12c0 1.77.42 3.45 1.18 4.93l3.66-2.84z"
                  fill="#FBBC05"
                />
                <path
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                  fill="#EA4335"
                />
              </svg>
              Continue with Google
            </a>
          </>
        )}

        <p className="mt-6 text-center text-sm text-text-secondary">
          Don&rsquo;t have an account?{" "}
          <Link
            href="/register"
            className="text-text-primary font-medium hover:underline"
          >
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
