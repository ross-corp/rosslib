"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { label: "Profile", href: "/settings" },
  { label: "Labels", href: "/settings/tags" },
  { label: "Import", href: "/settings/import" },
  { label: "Export", href: "/settings/export" },
  { label: "Ghost Activity", href: "/settings/ghost-activity" },
];

export default function SettingsNav() {
  const pathname = usePathname();

  return (
    <nav className="flex flex-wrap gap-2 mb-8">
      {navItems.map(({ label, href }) => {
        const isActive = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              isActive
                ? "bg-accent text-white"
                : "bg-surface-2 text-text-primary hover:bg-surface-3"
            }`}
          >
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
