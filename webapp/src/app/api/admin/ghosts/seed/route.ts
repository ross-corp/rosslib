import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function POST() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const apiRes = await fetch(`${process.env.API_URL}/admin/ghosts/seed`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
