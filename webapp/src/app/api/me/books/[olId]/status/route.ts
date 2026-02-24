import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ olId: string }> }
) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const { olId } = await params;
  const res = await fetch(`${process.env.API_URL}/me/books/${olId}/status`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
