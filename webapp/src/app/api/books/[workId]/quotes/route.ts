import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  const { searchParams } = new URL(req.url);
  const page = searchParams.get("page") || "1";
  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/quotes?page=${page}`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
