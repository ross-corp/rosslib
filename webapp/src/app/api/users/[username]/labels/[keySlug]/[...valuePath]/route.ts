import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  {
    params,
  }: {
    params: Promise<{ username: string; keySlug: string; valuePath: string[] }>;
  }
) {
  const { username, keySlug, valuePath } = await params;
  const valuePathStr = valuePath.join("/");
  const { searchParams } = new URL(req.url);
  const qs = searchParams.toString();

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/labels/${keySlug}/${valuePathStr}${qs ? `?${qs}` : ""}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
