import { cookies } from "next/headers";
import { NextResponse } from "next/server";

async function proxyFollow(
  username: string,
  method: "POST" | "DELETE"
): Promise<NextResponse> {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const apiRes = await fetch(
    `${process.env.API_URL}/users/${username}/follow`,
    { method, headers: { Authorization: `Bearer ${token}` } }
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}

export async function POST(
  _req: Request,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  return proxyFollow(username, "POST");
}

export async function DELETE(
  _req: Request,
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  return proxyFollow(username, "DELETE");
}
