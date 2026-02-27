import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const url = new URL(`${process.env.API_URL}/users`);
  const q = searchParams.get("q");
  if (q) url.searchParams.set("q", q);

  const apiRes = await fetch(url.toString());
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
