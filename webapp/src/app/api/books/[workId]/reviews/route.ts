import { NextResponse } from "next/server";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;

  const res = await fetch(`${process.env.API_URL}/books/${workId}/reviews`, {
    cache: "no-store",
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
