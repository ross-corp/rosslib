import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const q = searchParams.get("q");

  if (!q) {
    return NextResponse.json({ total: 0, results: [] });
  }

  try {
    const res = await fetch(`${process.env.API_URL}/books/search?q=${encodeURIComponent(q)}`, {
      cache: "no-store",
    });

    if (!res.ok) {
        return NextResponse.json({ error: "Backend error" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data, { status: 200 });
  } catch (error) {
    return NextResponse.json({ error: "Failed to fetch" }, { status: 500 });
  }
}
