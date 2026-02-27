import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const { searchParams } = new URL(req.url);
  const qs = searchParams.toString();
  const url = `${process.env.API_URL}/users/${username}/timeline${qs ? `?${qs}` : ""}`;

  const apiRes = await fetch(url, { cache: "no-store" });
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
