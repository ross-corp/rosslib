import { cookies } from "next/headers";
import { NextResponse } from "next/server";

async function proxy(
  req: Request,
  authorKey: string,
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

  if (method === "POST") {
    const body = await req.text();
    opts.headers = {
      ...opts.headers,
      "Content-Type": "application/json",
    };
    opts.body = body;
  }

  const apiRes = await fetch(
    `${process.env.API_URL}/authors/${authorKey}/follow`,
    opts
  );

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}

export async function POST(
  req: Request,
  { params }: { params: Promise<{ authorKey: string }> }
) {
  const { authorKey } = await params;
  return proxy(req, authorKey, "POST");
}

export async function DELETE(
  req: Request,
  { params }: { params: Promise<{ authorKey: string }> }
) {
  const { authorKey } = await params;
  return proxy(req, authorKey, "DELETE");
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ authorKey: string }> }
) {
  const { authorKey } = await params;
  return proxy(req, authorKey, "GET");
}
