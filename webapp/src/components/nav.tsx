import Link from "next/link";
import { getUser } from "@/lib/auth";

export default async function Nav() {
  const user = await getUser();

  return (
    <header className="border-b border-stone-200 bg-white">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 h-14 flex items-center gap-4">
        <Link
          href="/"
          className="font-semibold text-stone-900 tracking-tight text-lg shrink-0"
        >
          rosslib
        </Link>
        <form action="/search" method="get" className="flex-1 max-w-xs">
          <input
            name="q"
            type="search"
            placeholder="Search books..."
            className="w-full px-3 py-1.5 text-sm border border-stone-200 rounded bg-stone-50 text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent"
          />
        </form>
        <nav className="flex items-center gap-2 ml-auto shrink-0">
          <Link
            href="/users"
            className="text-sm text-stone-500 hover:text-stone-900 px-3 py-1.5 rounded transition-colors hidden sm:inline-flex"
          >
            People
          </Link>
          {user ? (
            <>
              <Link
                href="/feed"
                className="text-sm text-stone-500 hover:text-stone-900 px-3 py-1.5 rounded transition-colors hidden sm:inline-flex"
              >
                Feed
              </Link>
              <Link
                href={`/${user.username}`}
                className="text-sm text-stone-600 hover:text-stone-900 px-3 py-1.5 rounded transition-colors"
              >
                {user.username}
              </Link>
              <Link
                href="/api/auth/logout"
                className="text-sm text-stone-600 hover:text-stone-900 px-3 py-1.5 rounded transition-colors"
              >
                Sign out
              </Link>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm text-stone-600 hover:text-stone-900 px-3 py-1.5 rounded transition-colors"
              >
                Sign in
              </Link>
              <Link
                href="/register"
                className="text-sm bg-stone-900 text-white px-3 py-1.5 rounded hover:bg-stone-700 transition-colors"
              >
                Sign up
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
