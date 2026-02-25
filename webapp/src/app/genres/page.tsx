import Link from "next/link";
import Nav from "@/components/nav";

type GenreInfo = {
  slug: string;
  name: string;
  book_count: number;
};

async function fetchGenres(): Promise<GenreInfo[]> {
  const res = await fetch(`${process.env.API_URL}/genres`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function GenresPage() {
  const genres = await fetchGenres();

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-stone-900 mb-8">Genres</h1>

        {genres.length === 0 ? (
          <p className="text-sm text-stone-400">No genres available.</p>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
            {genres.map((genre) => (
              <Link
                key={genre.slug}
                href={`/genres/${genre.slug}`}
                className="group block border border-stone-200 rounded-lg p-5 hover:border-stone-400 hover:shadow-sm transition-all"
              >
                <h2 className="text-sm font-semibold text-stone-900 group-hover:text-stone-700 transition-colors">
                  {genre.name}
                </h2>
                <p className="text-xs text-stone-400 mt-1">
                  {genre.book_count.toLocaleString()} book{genre.book_count === 1 ? "" : "s"}
                </p>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
