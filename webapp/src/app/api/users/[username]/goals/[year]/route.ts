import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ username: string; year: string }> }
) {
  const { username, year } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  const headers: HeadersInit = token
    ? { Authorization: `Bearer ${token}` }
    : {};

  const res = await fetch(
    `${process.env.API_URL}/users/${username}/goals/${year}`,
    { headers, cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
