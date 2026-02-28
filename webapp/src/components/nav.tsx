import Link from "next/link";
import { getUser } from "@/lib/auth";
import NotificationBell from "@/components/notification-bell";
import NavDropdown from "@/components/nav-dropdown";

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
    <header className="bg-surface-1 border-b-2 border-border-strong">
      <div className="max-w-shell mx-auto px-6 h-11 flex items-center gap-4">
        <Link
          href="/"
          className="font-mono font-bold text-sm text-text-primary tracking-tight shrink-0"
        >
          rosslib
        </Link>
        <form action="/search" method="get" className="flex-1 max-w-xs relative">
          <input
            id="nav-search"
            name="q"
            type="search"
            placeholder="Search books..."
            className="w-full px-3 py-1 pr-12 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
          />
          <kbd className="absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none text-[10px] font-mono text-text-tertiary bg-surface-1 border border-border rounded px-1 py-0.5 leading-none">
            ⌘K
          </kbd>
        </form>
        <nav className="flex items-center gap-1 ml-auto shrink-0">
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
              >
                {user.username}
              </Link>
              <Link href="/api/auth/logout" className="nav-link">
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
      </div>
    </header>
  );
}
