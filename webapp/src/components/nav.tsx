import Link from "next/link";
import { getUser } from "@/lib/auth";
import NotificationBell from "@/components/notification-bell";

export default async function Nav() {
  const user = await getUser();

  return (
    <header className="bg-surface-1 border-b-2 border-border-strong">
      <div className="max-w-shell mx-auto px-6 h-11 flex items-center gap-4">
        <Link
          href="/"
          className="font-mono font-bold text-sm text-text-primary tracking-tight shrink-0"
        >
          rosslib
        </Link>
        <form action="/search" method="get" className="flex-1 max-w-xs">
          <input
            name="q"
            type="search"
            placeholder="Search books..."
            className="w-full px-3 py-1 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
          />
        </form>
        <nav className="flex items-center gap-1 ml-auto shrink-0">
          <Link href="/genres" className="nav-link">
            genres
          </Link>
          <Link href="/users" className="nav-link">
            people
          </Link>
          <span className="text-text-tertiary select-none">|</span>
          {user ? (
            <>
              <Link href="/feed" className="nav-link">
                feed
              </Link>
              <Link href="/scan" className="nav-link">
                scan
              </Link>
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
