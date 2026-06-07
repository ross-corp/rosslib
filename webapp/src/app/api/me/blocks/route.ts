import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const { searchParams } = new URL(request.url);
  const query = searchParams.toString();
  const url = `${process.env.API_URL}/me/blocks${query ? `?${query}` : ""}`;

  const apiRes = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
