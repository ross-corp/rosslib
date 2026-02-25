import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ count: 0 });
  }

  const apiRes = await fetch(
    `${process.env.API_URL}/me/notifications/unread-count`,
    { headers: { Authorization: `Bearer ${token}` } }
  );
  if (!apiRes.ok) return NextResponse.json({ count: 0 });
  const data = await apiRes.json();
  return NextResponse.json(data);
}
