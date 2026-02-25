import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const q = searchParams.get("q");

  if (!q) {
    return NextResponse.json({ total: 0, results: [] });
  }

  try {
    // Direct OpenLibrary search since we removed the custom search backend
    const res = await fetch(`https://openlibrary.org/search.json?q=${encodeURIComponent(q)}&limit=20`);

    if (!res.ok) {
        return NextResponse.json({ error: "OpenLibrary error" }, { status: res.status });
    }

    const data = await res.json();

    // Map OpenLibrary structure to expected API format
    const results = (data.docs || []).map((doc: any) => ({
      open_library_id: doc.key ? doc.key.replace("/works/", "") : null,
      title: doc.title,
      authors: doc.author_name ? doc.author_name.join(", ") : "Unknown",
      cover_url: doc.cover_i ? `https://covers.openlibrary.org/b/id/${doc.cover_i}-M.jpg` : null,
      publication_year: doc.first_publish_year,
      isbn13: doc.isbn ? doc.isbn[0] : null
    })).filter((r: any) => r.open_library_id);

    return NextResponse.json({ total: data.numFound, results }, { status: 200 });
  } catch (error) {
    return NextResponse.json({ error: "Failed to fetch" }, { status: 500 });
  }
}
