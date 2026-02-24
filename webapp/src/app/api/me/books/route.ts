import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function POST(req: Request) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const res = await fetch(`${process.env.API_URL}/me/books`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: JSON.stringify(await req.json()),
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
