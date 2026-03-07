import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  const apiRes = await fetch(
    `${process.env.API_URL}/books/${workId}/followers/count`,
    { cache: "no-store" }
  );
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
