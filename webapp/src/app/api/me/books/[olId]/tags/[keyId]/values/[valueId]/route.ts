import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function DELETE(
  _req: Request,
  { params }: { params: Promise<{ olId: string; keyId: string; valueId: string }> }
) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const { olId, keyId, valueId } = await params;
  const res = await fetch(
    `${process.env.API_URL}/me/books/${olId}/tags/${keyId}/values/${valueId}`,
    { method: "DELETE", headers: { Authorization: `Bearer ${token}` } }
  );
  return NextResponse.json(await res.json(), { status: res.status });
}
