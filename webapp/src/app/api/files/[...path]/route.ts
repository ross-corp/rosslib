import { NextRequest, NextResponse } from "next/server";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  const filePath = path.join("/");
  const url = `${process.env.API_URL}/api/files/${filePath}`;

  const res = await fetch(url);
  if (!res.ok) {
    return new NextResponse(null, { status: res.status });
  }

  const contentType = res.headers.get("content-type");
  const body = await res.arrayBuffer();

  return new NextResponse(body, {
    status: 200,
    headers: {
      "Content-Type": contentType || "application/octet-stream",
      "Cache-Control": "public, max-age=31536000, immutable",
    },
  });
}
