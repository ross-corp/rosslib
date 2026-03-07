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
  const sort = new URL(req.url).searchParams.get("sort");
  if (sort) url.searchParams.set("sort", sort);

  const res = await fetch(url.toString(), {
    headers,
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
