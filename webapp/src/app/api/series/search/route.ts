import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const q = searchParams.get("q") || "";
  const res = await fetch(
    `${process.env.API_URL}/series/search?q=${encodeURIComponent(q)}`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
