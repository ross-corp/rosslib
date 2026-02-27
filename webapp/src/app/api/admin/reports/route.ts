import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const url = new URL(req.url);
  const params = url.searchParams.toString();

  const apiRes = await fetch(
    `${process.env.API_URL}/admin/reports${params ? `?${params}` : ""}`,
    { headers: { Authorization: `Bearer ${token}` } }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
