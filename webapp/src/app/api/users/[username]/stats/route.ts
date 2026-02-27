import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/stats`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
