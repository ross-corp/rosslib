import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const apiRes = await fetch(`${process.env.API_URL}/me/account`, {
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}

export async function DELETE() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(`${process.env.API_URL}/me/account`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await res.json();

  if (res.ok) {
    const response = NextResponse.json(data, { status: res.status });
    response.cookies.set("token", "", { maxAge: 0, path: "/" });
    return response;
  }

  return NextResponse.json(data, { status: res.status });
}
