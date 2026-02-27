import Link from "next/link";

type GenreInfo = {
  slug: string;
  name: string;
  book_count: number;
};

const genreColors: Record<string, { bg: string; border: string; text: string }> = {
  fiction:        { bg: "bg-blue-950/50",    border: "border-blue-800/40",    text: "text-blue-300" },
  nonfiction:     { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  "non-fiction":  { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  mystery:        { bg: "bg-violet-950/50",  border: "border-violet-800/40",  text: "text-violet-300" },
  "science-fiction": { bg: "bg-cyan-950/50", border: "border-cyan-800/40",    text: "text-cyan-300" },
  fantasy:        { bg: "bg-purple-950/50",  border: "border-purple-800/40",  text: "text-purple-300" },
  romance:        { bg: "bg-rose-950/50",    border: "border-rose-800/40",    text: "text-rose-300" },
  horror:         { bg: "bg-red-950/50",     border: "border-red-800/40",     text: "text-red-300" },
  thriller:       { bg: "bg-orange-950/50",  border: "border-orange-800/40",  text: "text-orange-300" },
  biography:      { bg: "bg-teal-950/50",    border: "border-teal-800/40",    text: "text-teal-300" },
  history:        { bg: "bg-yellow-950/50",  border: "border-yellow-800/40",  text: "text-yellow-300" },
  poetry:         { bg: "bg-pink-950/50",    border: "border-pink-800/40",    text: "text-pink-300" },
  science:        { bg: "bg-emerald-950/50", border: "border-emerald-800/40", text: "text-emerald-300" },
  philosophy:     { bg: "bg-indigo-950/50",  border: "border-indigo-800/40",  text: "text-indigo-300" },
  children:       { bg: "bg-lime-950/50",    border: "border-lime-800/40",    text: "text-lime-300" },
  "young-adult":  { bg: "bg-fuchsia-950/50", border: "border-fuchsia-800/40", text: "text-fuchsia-300" },
};

const fallbackPalette = [
  { bg: "bg-sky-950/50",     border: "border-sky-800/40",     text: "text-sky-300" },
  { bg: "bg-emerald-950/50", border: "border-emerald-800/40", text: "text-emerald-300" },
  { bg: "bg-violet-950/50",  border: "border-violet-800/40",  text: "text-violet-300" },
  { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  { bg: "bg-rose-950/50",    border: "border-rose-800/40",    text: "text-rose-300" },
  { bg: "bg-cyan-950/50",    border: "border-cyan-800/40",    text: "text-cyan-300" },
  { bg: "bg-orange-950/50",  border: "border-orange-800/40",  text: "text-orange-300" },
  { bg: "bg-teal-950/50",    border: "border-teal-800/40",    text: "text-teal-300" },
];

function getGenreColor(slug: string, index: number) {
  return genreColors[slug] ?? fallbackPalette[index % fallbackPalette.length];
}

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
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-text-primary mb-8">Genres</h1>

        {genres.length === 0 ? (
          <p className="text-sm text-text-primary">No genres available.</p>
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
