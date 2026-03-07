import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const { searchParams } = new URL(req.url);
  const query = searchParams.toString();
  const qs = query ? `?${query}` : "";

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/reviews${qs}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
