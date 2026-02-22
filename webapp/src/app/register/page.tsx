import Link from "next/link";

export const metadata = {
  title: "Create account — rosslib",
};

export default function RegisterPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4">
      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-8">
          <Link href="/" className="font-semibold text-stone-900 text-xl">
            rosslib
          </Link>
          <h1 className="mt-4 text-2xl font-bold text-stone-900">
            Create your account
          </h1>
        </div>

        {/* Form */}
        <form className="space-y-4">
          <div>
            <label
              htmlFor="username"
              className="block text-sm font-medium text-stone-700 mb-1"
            >
              Username
            </label>
            <input
              id="username"
              name="username"
              type="text"
              autoComplete="username"
              required
              className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm"
              placeholder="yourname"
            />
            <p className="mt-1 text-xs text-stone-400">
              Lowercase letters, numbers, and hyphens only.
            </p>
          </div>

          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-stone-700 mb-1"
            >
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              required
              className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-stone-700 mb-1"
            >
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="new-password"
              required
              className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm"
              placeholder="••••••••"
            />
            <p className="mt-1 text-xs text-stone-400">At least 8 characters.</p>
          </div>

          <button
            type="submit"
            className="w-full bg-stone-900 text-white py-2.5 rounded font-medium hover:bg-stone-700 transition-colors text-sm"
          >
            Create account
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-stone-500">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-stone-900 font-medium hover:underline"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
