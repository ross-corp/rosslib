import Link from "next/link";
import { getUser } from "@/lib/auth";

export default async function Nav() {
  const user = await getUser();

  return (
    <header className="border-b-2 border-black bg-white sticky top-0 z-50">
      <div className="max-w-5xl mx-auto px-4 h-16 flex items-center gap-4">
        <Link
          href="/"
          className="font-bold text-2xl tracking-tighter uppercase border-2 border-black p-1 shadow-[4px_4px_0_0_#000] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0_0_#000] transition-all bg-white text-black shrink-0"
        >
          ROSSLIB
        </Link>
        <form action="/search" method="get" className="flex-1 max-w-xs">
          <input
            name="q"
            type="search"
            placeholder="SEARCH..."
            className="w-full uppercase placeholder:text-stone-400"
          />
        </form>
        <nav className="flex items-center divide-x-2 divide-black border-2 border-black bg-white shadow-[4px_4px_0_0_#000] ml-auto shrink-0">
          <Link
            href="/genres"
            className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors hidden sm:inline-block"
          >
            Genres
          </Link>
          <Link
            href="/users"
            className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors hidden sm:inline-block"
          >
            People
          </Link>
          {user ? (
            <>
              <Link
                href="/feed"
                className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors hidden sm:inline-block"
              >
                Feed
              </Link>
              <Link
                href={`/${user.username}`}
                className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors"
              >
                {user.username}
              </Link>
              <Link
                href="/api/auth/logout"
                className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors"
              >
                Sign out
              </Link>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="px-4 py-2 text-sm font-bold uppercase hover:bg-black hover:text-white transition-colors"
              >
                Sign in
              </Link>
              <Link
                href="/register"
                className="px-4 py-2 text-sm font-bold uppercase bg-black text-white hover:bg-white hover:text-black transition-colors"
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
