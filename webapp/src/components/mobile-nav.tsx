"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";

interface MobileNavProps {
  user: { username: string; is_moderator: boolean } | null;
  browseItems: { href: string; label: string }[];
  communityItems: { href: string; label: string }[];
}

export default function MobileNav({
  user,
  browseItems,
  communityItems,
}: MobileNavProps) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [open]);

  return (
    <div ref={menuRef} className="md:hidden ml-auto">
      <button
        onClick={() => setOpen((o) => !o)}
        className="nav-link text-lg leading-none"
      >
        {open ? "✕" : "☰"}
      </button>
      {open && (
        <div className="absolute left-0 right-0 top-full bg-surface-1 border-b-2 border-border-strong z-50">
          <nav className="max-w-shell mx-auto px-6 py-3 flex flex-col gap-1">
            <span className="section-heading mb-1">Browse</span>
            {browseItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="nav-link py-1.5"
                onClick={() => setOpen(false)}
              >
                {item.label}
              </Link>
            ))}
            <span className="section-heading mt-2 mb-1">Community</span>
            {communityItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="nav-link py-1.5"
                onClick={() => setOpen(false)}
              >
                {item.label}
              </Link>
            ))}
            <div className="divider my-2" />
            {user ? (
              <>
                <Link
                  href="/notifications"
                  className="nav-link py-1.5"
                  onClick={() => setOpen(false)}
                >
                  Notifications
                </Link>
                {user.is_moderator && (
                  <Link
                    href="/admin"
                    className="nav-link py-1.5"
                    onClick={() => setOpen(false)}
                  >
                    Admin
                  </Link>
                )}
                <Link
                  href={`/${user.username}`}
                  className="nav-link py-1.5 text-text-primary"
                  onClick={() => setOpen(false)}
                >
                  {user.username}
                </Link>
                <Link
                  href="/api/auth/logout"
                  className="nav-link py-1.5"
                  onClick={() => setOpen(false)}
                >
                  Sign out
                </Link>
              </>
            ) : (
              <>
                <Link
                  href="/login"
                  className="nav-link py-1.5"
                  onClick={() => setOpen(false)}
                >
                  Sign in
                </Link>
                <Link
                  href="/register"
                  className="btn-primary font-mono text-xs px-3 py-1.5 text-center mt-1"
                  onClick={() => setOpen(false)}
                >
                  Sign up
                </Link>
              </>
            )}
          </nav>
        </div>
      )}
    </div>
  );
}
