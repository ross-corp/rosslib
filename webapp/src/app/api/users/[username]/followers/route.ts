import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const { searchParams } = new URL(req.url);
  const qs = searchParams.toString();

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/followers${qs ? `?${qs}` : ""}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
