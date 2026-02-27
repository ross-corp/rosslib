import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function POST(
  _req: Request,
  { params }: { params: Promise<{ workId: string; userId: string }> }
) {
  const { workId, userId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/reviews/${userId}/like`,
    {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ workId: string; userId: string }> }
) {
  const { workId, userId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(
    `${process.env.API_URL}/books/${workId}/reviews/${userId}/like`,
    {
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
