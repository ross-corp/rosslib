import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  {
    params,
  }: {
    params: Promise<{ username: string; keySlug: string; valuePath: string[] }>;
  }
) {
  const { username, keySlug, valuePath } = await params;
  const valuePathStr = valuePath.join("/");

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/labels/${keySlug}/${valuePathStr}`,
    { cache: "no-store" }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
