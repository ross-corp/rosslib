"use client";

import Link from "next/link";

export default function Error({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <h1 className="text-2xl font-bold text-text-primary mb-2">
        Something went wrong
      </h1>
      <p className="text-text-secondary text-sm mb-6">
        We had trouble loading this settings page. Please try again.
      </p>
      <div className="flex items-center gap-3">
        <button
          onClick={reset}
          className="btn-primary font-mono text-xs px-4 py-2"
        >
          Try again
        </button>
        <Link
          href="/settings"
          className="text-sm text-accent hover:underline transition-colors"
        >
          Back to settings
        </Link>
      </div>
    </div>
  );
}
