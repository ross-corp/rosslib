import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ username: string; slug: string }> }
) {
  const { username, slug } = await params;

  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves/${slug}`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
