import { NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;

  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const url = new URL(`${process.env.API_URL}/books/${workId}/reviews`);
  const incoming = new URL(req.url).searchParams;
  for (const key of ["sort", "limit", "offset"]) {
    const value = incoming.get(key);
    if (value) url.searchParams.set(key, value);
  }

  const res = await fetch(url.toString(), {
    headers,
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
