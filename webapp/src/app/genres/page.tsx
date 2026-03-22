import type { Metadata } from "next";
import Link from "next/link";
import EmptyState from "@/components/empty-state";
import { getGenreColor } from "@/lib/genre-colors";

export const metadata: Metadata = {
  title: "Genres",
};

type GenreInfo = {
  slug: string;
  name: string;
  book_count: number;
};

async function fetchGenres(): Promise<GenreInfo[]> {
  const res = await fetch(`${process.env.API_URL}/genres`, {
    cache: "no-store",
  });
  if (!res.ok) throw new Error("Failed to load genres");
  return res.json();
}

export default async function GenresPage() {
  const genres = await fetchGenres();

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-text-primary mb-8">Genres</h1>

        {genres.length === 0 ? (
          <EmptyState
            message="No genres available."
            actionLabel="Browse books"
            actionHref="/search"
          />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
            {genres.map((genre, i) => {
              const color = getGenreColor(genre.slug, i);
              return (
                <Link
                  key={genre.slug}
                  href={`/genres/${genre.slug}`}
                  className={`group block rounded-lg p-5 border transition-all hover:shadow-sm hover:brightness-110 ${color.bg} ${color.border}`}
                >
                  <h2 className={`text-sm font-semibold transition-colors ${color.text}`}>
                    {genre.name}
                  </h2>
                  <p className="text-xs text-text-tertiary mt-1">
                    {genre.book_count.toLocaleString()} book{genre.book_count === 1 ? "" : "s"}
                  </p>
                </Link>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}
