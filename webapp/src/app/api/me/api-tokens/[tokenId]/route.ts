import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function DELETE(
  _req: Request,
  { params }: { params: Promise<{ tokenId: string }> }
) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const { tokenId } = await params;

  const apiRes = await fetch(
    `${process.env.API_URL}/me/api-tokens/${tokenId}`,
    {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
