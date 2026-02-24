import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const shelf = req.nextUrl.searchParams.get("shelf") || "";
  const url = shelf
    ? `${process.env.API_URL}/me/export/csv?shelf=${encodeURIComponent(shelf)}`
    : `${process.env.API_URL}/me/export/csv`;

  const res = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!res.ok) {
    return NextResponse.json({ error: "export failed" }, { status: res.status });
  }

  return new NextResponse(res.body, {
    status: 200,
    headers: {
      "Content-Type": "text/csv",
      "Content-Disposition": res.headers.get("Content-Disposition") || 'attachment; filename="rosslib-export.csv"',
    },
  });
}
