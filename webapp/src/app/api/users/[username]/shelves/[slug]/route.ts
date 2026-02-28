import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ username: string; slug: string }> }
) {
  const { username, slug } = await params;
  const sort = req.nextUrl.searchParams.get("sort") ?? "";
  const qs = sort ? `?sort=${sort}` : "";

  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves/${slug}${qs}`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
