"use client";

import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function RegisterPage() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [registered, setRegistered] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const googleEnabled = !!process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setLoading(true);

    const form = e.currentTarget;
    const emailValue = (form.elements.namedItem("email") as HTMLInputElement).value;
    const data = {
      username: (form.elements.namedItem("username") as HTMLInputElement).value,
      email: emailValue,
      password: (form.elements.namedItem("password") as HTMLInputElement).value,
    };

    const res = await fetch("/api/auth/register", {
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

    setRegisteredEmail(emailValue);
    setRegistered(true);
  }

  if (registered) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center">
        <div className="w-full max-w-sm text-center">
          <h1 className="text-2xl font-bold text-text-primary">
            Check your email
          </h1>
          <p className="mt-3 text-sm text-text-secondary">
            We sent a verification link to{" "}
            <span className="font-medium text-text-primary">{registeredEmail}</span>.
            Click the link to verify your email and get full access.
          </p>
          <p className="mt-4 text-xs text-text-tertiary">
            {"Didn't"} receive it? Check your spam folder, or{" "}
            <button
              onClick={() => {
                router.push("/");
                router.refresh();
              }}
              className="text-text-primary font-medium hover:underline"
            >
              continue to rosslib
            </button>{" "}
            and resend from your settings page.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-text-primary">
            Create your account
          </h1>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          {error && (
            <p role="alert" aria-live="assertive" className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded px-3 py-2">
              {error}
            </p>
          )}

          <div>
            <label
              htmlFor="username"
              className="label-mono block mb-1"
            >
              Username
            </label>
            <input
              id="username"
              name="username"
              type="text"
              autoComplete="username"
              required
              className="input-field"
              placeholder="yourname"
            />
            <p className="mt-1 text-xs text-text-tertiary">
              Lowercase letters, numbers, and hyphens only.
            </p>
          </div>

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
            <label
              htmlFor="password"
              className="label-mono block mb-1"
            >
              Password
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
            <p className="mt-1 text-xs text-text-tertiary">At least 8 characters.</p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full btn-primary py-2.5"
          >
            {loading ? "Creating account..." : "Create account"}
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
              Sign up with Google
            </a>
          </>
        )}

        <p className="mt-6 text-center text-sm text-text-secondary">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-text-primary font-medium hover:underline"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
