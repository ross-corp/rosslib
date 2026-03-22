export type GenreColor = { bg: string; border: string; text: string };

const genreColors: Record<string, GenreColor> = {
  fiction:            { bg: "bg-blue-950/50",    border: "border-blue-800/40",    text: "text-blue-300" },
  nonfiction:         { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  "non-fiction":      { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  mystery:            { bg: "bg-violet-950/50",  border: "border-violet-800/40",  text: "text-violet-300" },
  "science fiction":  { bg: "bg-cyan-950/50",    border: "border-cyan-800/40",    text: "text-cyan-300" },
  "science-fiction":  { bg: "bg-cyan-950/50",    border: "border-cyan-800/40",    text: "text-cyan-300" },
  fantasy:            { bg: "bg-purple-950/50",  border: "border-purple-800/40",  text: "text-purple-300" },
  romance:            { bg: "bg-rose-950/50",    border: "border-rose-800/40",    text: "text-rose-300" },
  horror:             { bg: "bg-red-950/50",     border: "border-red-800/40",     text: "text-red-300" },
  thriller:           { bg: "bg-orange-950/50",  border: "border-orange-800/40",  text: "text-orange-300" },
  biography:          { bg: "bg-teal-950/50",    border: "border-teal-800/40",    text: "text-teal-300" },
  history:            { bg: "bg-yellow-950/50",  border: "border-yellow-800/40",  text: "text-yellow-300" },
  poetry:             { bg: "bg-pink-950/50",    border: "border-pink-800/40",    text: "text-pink-300" },
  science:            { bg: "bg-emerald-950/50", border: "border-emerald-800/40", text: "text-emerald-300" },
  philosophy:         { bg: "bg-indigo-950/50",  border: "border-indigo-800/40",  text: "text-indigo-300" },
  children:           { bg: "bg-lime-950/50",    border: "border-lime-800/40",    text: "text-lime-300" },
  "young adult":      { bg: "bg-fuchsia-950/50", border: "border-fuchsia-800/40", text: "text-fuchsia-300" },
  "young-adult":      { bg: "bg-fuchsia-950/50", border: "border-fuchsia-800/40", text: "text-fuchsia-300" },
};

const fallbackGenreColors: GenreColor[] = [
  { bg: "bg-sky-950/50",     border: "border-sky-800/40",     text: "text-sky-300" },
  { bg: "bg-emerald-950/50", border: "border-emerald-800/40", text: "text-emerald-300" },
  { bg: "bg-violet-950/50",  border: "border-violet-800/40",  text: "text-violet-300" },
  { bg: "bg-amber-950/50",   border: "border-amber-800/40",   text: "text-amber-300" },
  { bg: "bg-rose-950/50",    border: "border-rose-800/40",    text: "text-rose-300" },
  { bg: "bg-cyan-950/50",    border: "border-cyan-800/40",    text: "text-cyan-300" },
  { bg: "bg-orange-950/50",  border: "border-orange-800/40",  text: "text-orange-300" },
  { bg: "bg-teal-950/50",    border: "border-teal-800/40",    text: "text-teal-300" },
];

export function getGenreColor(key: string, index: number): GenreColor {
  return genreColors[key] ?? fallbackGenreColors[index % fallbackGenreColors.length];
}
