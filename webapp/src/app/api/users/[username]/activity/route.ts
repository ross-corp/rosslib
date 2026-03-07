import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const cursor = req.nextUrl.searchParams.get("cursor");

  const url = new URL(`${process.env.API_URL}/users/${username}/activity`);
  if (cursor) url.searchParams.set("cursor", cursor);

  const apiRes = await fetch(url.toString(), { cache: "no-store" });
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
