import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ authorKey: string }> }
) {
  const { authorKey } = await params;
  const limit = req.nextUrl.searchParams.get("limit") ?? "24";
  const offset = req.nextUrl.searchParams.get("offset") ?? "0";

  const res = await fetch(
    `${process.env.API_URL}/authors/${authorKey}?limit=${limit}&offset=${offset}`,
    { cache: "no-store" }
  );

  if (!res.ok) {
    return NextResponse.json({ works: [] }, { status: res.status });
  }

  const data = await res.json();
  return NextResponse.json({ works: data.works ?? [], work_count: data.work_count ?? 0 });
}
