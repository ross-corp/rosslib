import Link from "next/link";

export default function Nav() {
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
        </nav>
      </div>
    </header>
  );
}
