import Link from "next/link";

export default function AuthorNotFound() {
  return (
    <div className="text-center py-16">
      <h1 className="text-2xl font-bold text-text-primary mb-2">
        Author not found
      </h1>
      <p className="text-sm text-text-tertiary mb-6">
        The author you&apos;re looking for doesn&apos;t exist or may have been
        removed.
      </p>
      <Link
        href="/search"
        className="inline-block text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
      >
        Search for authors
      </Link>
    </div>
  );
}
