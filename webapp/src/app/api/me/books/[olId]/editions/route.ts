import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ olId: string }> }
) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const { olId } = await params;
  const { searchParams } = new URL(req.url);
  const url = new URL(`${process.env.API_URL}/me/books/${olId}/editions`);
  const limit = searchParams.get("limit");
  const offset = searchParams.get("offset");
  if (limit) url.searchParams.set("limit", limit);
  if (offset) url.searchParams.set("offset", offset);

  const res = await fetch(url.toString(), {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
