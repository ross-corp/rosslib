import { cookies } from "next/headers";
import { NextResponse } from "next/server";

async function proxy(
  req: Request,
  workId: string,
  method: "POST" | "DELETE" | "GET"
): Promise<NextResponse> {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const opts: RequestInit = {
    method,
    headers: { Authorization: `Bearer ${token}` },
  };

  const apiRes = await fetch(
    `${process.env.API_URL}/books/${workId}/follow`,
    opts
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}

export async function POST(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  return proxy(req, workId, "POST");
}

export async function DELETE(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  return proxy(req, workId, "DELETE");
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ workId: string }> }
) {
  const { workId } = await params;
  return proxy(req, workId, "GET");
}
