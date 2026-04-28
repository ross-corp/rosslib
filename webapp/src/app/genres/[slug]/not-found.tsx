import Link from "next/link";

export default function GenreNotFound() {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <h1 className="text-2xl font-bold text-text-primary mb-2">
        Genre not found
      </h1>
      <p className="text-text-secondary text-sm mb-6">
        We couldn&apos;t find the genre you&apos;re looking for.
      </p>
      <Link href="/genres" className="btn-primary font-mono text-xs px-4 py-2">
        Browse genres
      </Link>
    </div>
  );
}
