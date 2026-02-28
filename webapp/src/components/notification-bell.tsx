"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

export default function NotificationBell() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    fetch("/api/me/notifications/unread-count")
      .then((r) => (r.ok ? r.json() : { count: 0 }))
      .then((d) => setCount(d.count ?? 0))
      .catch(() => {});

    // Poll every 60 seconds.
    const interval = setInterval(() => {
      fetch("/api/me/notifications/unread-count")
        .then((r) => (r.ok ? r.json() : { count: 0 }))
        .then((d) => setCount(d.count ?? 0))
        .catch(() => {});
    }, 60_000);

    return () => clearInterval(interval);
  }, []);

  return (
    <Link
      href="/notifications"
      className="relative text-sm text-text-secondary hover:text-text-primary px-2 py-1 transition-colors"
      title="Notifications"
      aria-label={count > 0 ? `Notifications (${count} unread)` : "Notifications"}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 20 20"
        fill="currentColor"
        className="w-4 h-4"
        aria-hidden="true"
      >
        <path
          fillRule="evenodd"
          d="M10 2a6 6 0 00-6 6c0 1.887-.454 3.665-1.257 5.234a.75.75 0 00.515 1.076 32.91 32.91 0 003.256.508 3.5 3.5 0 006.972 0 32.903 32.903 0 003.256-.508.75.75 0 00.515-1.076A11.448 11.448 0 0116 8a6 6 0 00-6-6zM8.05 14.943a33.54 33.54 0 003.9 0 2 2 0 01-3.9 0z"
          clipRule="evenodd"
        />
      </svg>
      {count > 0 && (
        <span
          className="absolute -top-0.5 -right-0.5 bg-badge text-badge-text text-[10px] font-bold rounded-full w-4 h-4 flex items-center justify-center leading-none"
          aria-hidden="true"
        >
          {count > 9 ? "9+" : count}
        </span>
      )}
    </Link>
  );
}
