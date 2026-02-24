import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  {
    params,
  }: {
    params: Promise<{ username: string; keySlug: string; valueSlug: string }>;
  }
) {
  const { username, keySlug, valueSlug } = await params;

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/labels/${keySlug}/${valueSlug}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
