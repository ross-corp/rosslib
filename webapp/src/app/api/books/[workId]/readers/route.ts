import { NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;

  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${process.env.API_URL}/books/${workId}/readers`, {
    headers,
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
