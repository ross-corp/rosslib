export type Edition = {
  key: string;
  title: string;
  publisher: string | null;
  publish_date: string;
  page_count: number | null;
  isbn: string | null;
  cover_url: string | null;
  format: string;
  language: string;
};

export const LANG_NAMES: Record<string, string> = {
  eng: "English",
  spa: "Spanish",
  fre: "French",
  ger: "German",
  por: "Portuguese",
  ita: "Italian",
  dut: "Dutch",
  rus: "Russian",
  jpn: "Japanese",
  chi: "Chinese",
  kor: "Korean",
  ara: "Arabic",
  hin: "Hindi",
  pol: "Polish",
  swe: "Swedish",
  nor: "Norwegian",
  dan: "Danish",
  fin: "Finnish",
  tur: "Turkish",
  heb: "Hebrew",
};

export function langName(code: string): string {
  return LANG_NAMES[code] ?? code;
}

export function formatLabel(format: string): string {
  if (!format) return "";
  const lower = format.toLowerCase();
  if (lower.includes("hardcover") || lower.includes("hardback") || lower === "capa dura")
    return "Hardcover";
  if (lower.includes("paperback") || lower.includes("softcover") || lower === "mass market")
    return "Paperback";
  if (lower.includes("ebook") || lower.includes("e-book") || lower.includes("kindle"))
    return "eBook";
  if (lower.includes("audio")) return "Audiobook";
  return format.charAt(0).toUpperCase() + format.slice(1);
}
