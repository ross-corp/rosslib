import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const { searchParams } = new URL(req.url);
  const qs = searchParams.toString();
  const url = `${process.env.API_URL}/users/${username}/year-in-review${qs ? `?${qs}` : ""}`;

  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const apiRes = await fetch(url, { cache: "no-store", headers });
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
