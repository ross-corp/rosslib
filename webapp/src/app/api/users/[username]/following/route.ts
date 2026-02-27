import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  const { searchParams } = new URL(req.url);
  const qs = searchParams.toString();

  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/following${qs ? `?${qs}` : ""}`,
    { cache: "no-store", headers }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
