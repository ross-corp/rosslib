"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { label: "Profile", href: "/settings" },
  { label: "Labels", href: "/settings/tags" },
  { label: "Import", href: "/settings/import" },
  { label: "Pending imports", href: "/settings/imports/pending" },
  { label: "Export", href: "/settings/export" },
  { label: "API Tokens", href: "/settings/api-tokens" },
  { label: "Blocked users", href: "/settings/blocked" },
  { label: "Follow requests", href: "/settings/follow-requests" },
  { label: "Followed books", href: "/settings/followed-books" },
  { label: "Followed authors", href: "/settings/followed-authors" },
  { label: "Ghost Activity", href: "/settings/ghost-activity" },
  { label: "My Feedback", href: "/settings/feedback" },
];

export default function SettingsNav() {
  const pathname = usePathname();
  const [requestCount, setRequestCount] = useState(0);
  const [pendingImportCount, setPendingImportCount] = useState(0);

  useEffect(() => {
    fetch("/api/me/follow-requests")
      .then((res) => (res.ok ? res.json() : []))
      .then((data: unknown[]) => setRequestCount(data.length))
      .catch(() => {});
    fetch("/api/me/imports/pending")
      .then((res) => (res.ok ? res.json() : []))
      .then((data: unknown[]) => setPendingImportCount(data.length))
      .catch(() => {});
  }, []);

  function getBadgeCount(href: string): number {
    if (href === "/settings/follow-requests") return requestCount;
    if (href === "/settings/imports/pending") return pendingImportCount;
    return 0;
  }

  return (
    <nav className="flex flex-wrap gap-2 mb-8">
      {navItems.map(({ label, href }) => {
        const isActive = pathname === href;
        const badgeCount = getBadgeCount(href);
        return (
          <Link
            key={href}
            href={href}
            className={`relative px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              isActive
                ? "bg-accent text-text-inverted"
                : "bg-surface-2 text-text-primary hover:bg-surface-3"
            }`}
          >
            {label}
            {badgeCount > 0 && (
              <span className="absolute -top-1.5 -right-1.5 inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold leading-none">
                {badgeCount}
              </span>
            )}
          </Link>
        );
      })}
    </nav>
  );
}
