import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const isbn = searchParams.get("isbn");

  if (!isbn) {
    return NextResponse.json({ error: "isbn parameter is required" }, { status: 400 });
  }

  try {
    const res = await fetch(
      `${process.env.API_URL}/books/lookup?isbn=${encodeURIComponent(isbn)}`,
      { cache: "no-store" },
    );

    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Failed to look up book" }, { status: 500 });
  }
}
