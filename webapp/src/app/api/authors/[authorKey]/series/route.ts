import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ authorKey: string }> }
) {
  const { authorKey } = await params;
  const { searchParams } = new URL(req.url);
  const name = searchParams.get("name") || "";
  const res = await fetch(
    `${process.env.API_URL}/authors/${authorKey}/series?name=${encodeURIComponent(name)}`,
    { cache: "no-store" }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
