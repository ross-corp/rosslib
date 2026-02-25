import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  const { searchParams } = new URL(req.url);
  const limit = searchParams.get("limit") ?? "50";
  const offset = searchParams.get("offset") ?? "0";

  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/editions?limit=${limit}&offset=${offset}`,
    { cache: "no-store" }
  );

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
