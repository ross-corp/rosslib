"use client";

export default function Error({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <h1 className="text-2xl font-bold text-text-primary mb-2">
        Failed to load admin panel
      </h1>
      <p className="text-text-secondary text-sm mb-6">
        Something went wrong. Make sure you have moderator permissions.
      </p>
      <button
        onClick={reset}
        className="btn-primary font-mono text-xs px-4 py-2"
      >
        Try again
      </button>
    </div>
  );
}
