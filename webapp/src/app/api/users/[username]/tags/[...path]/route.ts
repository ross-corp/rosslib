import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ username: string; path: string[] }> }
) {
  const { username, path } = await params;
  const tagPath = path.join("/");

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/tags/${tagPath}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
