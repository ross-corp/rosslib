import Link from "next/link";

export const metadata = {
  title: "Sign in — rosslib",
};

export default function LoginPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4">
      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-8">
          <Link href="/" className="font-semibold text-stone-900 text-xl">
            rosslib
          </Link>
          <h1 className="mt-4 text-2xl font-bold text-stone-900">
            Sign in to your account
          </h1>
        </div>

        {/* Form */}
        <form className="space-y-4">
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
            <div className="flex items-center justify-between mb-1">
              <label
                htmlFor="password"
                className="block text-sm font-medium text-stone-700"
              >
                Password
              </label>
              <Link
                href="/forgot-password"
                className="text-xs text-stone-500 hover:text-stone-900"
              >
                Forgot password?
              </Link>
            </div>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            className="w-full bg-stone-900 text-white py-2.5 rounded font-medium hover:bg-stone-700 transition-colors text-sm"
          >
            Sign in
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-stone-500">
          Don&rsquo;t have an account?{" "}
          <Link
            href="/register"
            className="text-stone-900 font-medium hover:underline"
          >
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
