import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ recId: string }> }
) {
  const { recId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = await req.json();
  const apiRes = await fetch(
    `${process.env.API_URL}/me/recommendations/${recId}`,
    {
      method: "PATCH",
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
