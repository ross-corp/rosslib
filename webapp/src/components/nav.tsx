import Link from "next/link";
import { getUser } from "@/lib/auth";
import NotificationBell from "@/components/notification-bell";
import NavDropdown from "@/components/nav-dropdown";
import KeyboardShortcutHint from "@/components/keyboard-shortcut-hint";
import MobileNav from "@/components/mobile-nav";

export default async function Nav() {
  const user = await getUser();

  const browseItems = [
    { href: "/search", label: "Search books" },
    { href: "/genres", label: "Genres" },
    ...(user ? [{ href: "/scan", label: "Scan ISBN" }] : []),
  ];

  const communityItems = [
    { href: "/users", label: "Browse users" },
    ...(user ? [{ href: "/feed", label: "My feed" }] : []),
  ];

  return (
    <header className="bg-surface-1 border-b-2 border-border-strong relative">
      <div className="max-w-shell mx-auto px-6 h-11 flex items-center gap-4">
        <Link
          href="/"
          className="font-mono font-bold text-sm text-text-primary tracking-tight shrink-0"
        >
          rosslib
        </Link>
        <form action="/search" method="get" className="hidden md:flex flex-1 max-w-xs relative">
          <input
            id="nav-search"
            name="q"
            type="search"
            placeholder="Search books..."
            aria-label="Search books"
            className="w-full px-3 py-1 pr-12 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
          />
          <KeyboardShortcutHint
            keys={{ mac: "⌘K", other: "Ctrl+K" }}
            className="absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none text-[10px] font-mono text-text-tertiary bg-surface-1 border border-border rounded px-1 py-0.5 leading-none"
          />
        </form>
        <nav className="hidden md:flex items-center gap-1 ml-auto shrink-0">
          <NavDropdown label="browse" items={browseItems} />
          <NavDropdown label="community" items={communityItems} />
          {user ? (
            <>
              <NotificationBell />
              {user.is_moderator && (
                <Link href="/admin" className="nav-link">
                  admin
                </Link>
              )}
              <span className="text-text-tertiary select-none">|</span>
              <Link
                href={`/${user.username}`}
                className="nav-link text-text-primary"
                aria-label="User menu"
              >
                {user.username}
              </Link>
              <Link href="/api/auth/logout" className="nav-link" aria-label="Sign out">
                sign out
              </Link>
            </>
          ) : (
            <>
              <Link href="/login" className="nav-link">
                sign in
              </Link>
              <Link
                href="/register"
                className="btn-primary font-mono text-xs px-3 py-1"
              >
                sign up
              </Link>
            </>
          )}
        </nav>
        <MobileNav
          user={user ? { username: user.username, is_moderator: user.is_moderator } : null}
          browseItems={browseItems}
          communityItems={communityItems}
        />
      </div>
    </header>
  );
}
