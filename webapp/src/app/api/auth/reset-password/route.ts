import { NextResponse } from "next/server";

export async function POST(req: Request) {
  const body = await req.json();

  const apiRes = await fetch(`${process.env.API_URL}/auth/reset-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  const data = await apiRes.json();
  return NextResponse.json(data, { status: apiRes.status });
}
