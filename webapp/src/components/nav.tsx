import Link from "next/link";
import { getUser } from "@/lib/auth";

export default async function Nav() {
  const user = await getUser();

  return (
    <header className="border-b border-stone-200 bg-white">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        <Link
          href="/"
          className="font-semibold text-stone-900 tracking-tight text-lg"
        >
          rosslib
        </Link>
        <nav className="flex items-center gap-2">
          {user ? (
            <>
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
