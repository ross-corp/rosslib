import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const { searchParams } = new URL(req.url);
  const url = new URL(`${process.env.API_URL}/me/notifications`);
  const cursor = searchParams.get("cursor");
  if (cursor) url.searchParams.set("cursor", cursor);

  const apiRes = await fetch(url.toString(), {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
