import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const apiRes = await fetch(
    `${process.env.API_URL}/me/theme`,
    { headers: { Authorization: `Bearer ${token}` } }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}

export async function PUT(req: Request) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = await req.json();
  const apiRes = await fetch(
    `${process.env.API_URL}/me/theme`,
    {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
