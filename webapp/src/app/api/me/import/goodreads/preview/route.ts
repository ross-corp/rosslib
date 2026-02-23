import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function POST(req: Request) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  // Forward the multipart form data as-is — do NOT set Content-Type manually
  // so that fetch fills in the correct boundary for multipart/form-data.
  const formData = await req.formData();
  const res = await fetch(
    `${process.env.API_URL}/me/import/goodreads/preview`,
    {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
