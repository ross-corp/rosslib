import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ shelfId: string; olId: string }> }
) {
  const { shelfId, olId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(
    `${process.env.API_URL}/shelves/${shelfId}/books/${olId}`,
    {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(await req.json()),
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _req: Request,
  { params }: { params: Promise<{ shelfId: string; olId: string }> }
) {
  const { shelfId, olId } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(
    `${process.env.API_URL}/shelves/${shelfId}/books/${olId}`,
    {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
