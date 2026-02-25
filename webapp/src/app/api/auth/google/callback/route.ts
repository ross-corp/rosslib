import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const url = new URL(req.url);
  const code = url.searchParams.get("code");
  const error = url.searchParams.get("error");

  const baseUrl = process.env.NEXT_PUBLIC_URL || "http://localhost:3000";

  if (error || !code) {
    return NextResponse.redirect(`${baseUrl}/login?error=google_denied`);
  }

  const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
  const clientSecret = process.env.GOOGLE_CLIENT_SECRET;
  if (!clientId || !clientSecret) {
    return NextResponse.redirect(`${baseUrl}/login?error=google_not_configured`);
  }

  const redirectUri = `${baseUrl}/api/auth/google/callback`;

  // Exchange authorization code for tokens.
  const tokenRes = await fetch("https://oauth2.googleapis.com/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      code,
      client_id: clientId,
      client_secret: clientSecret,
      redirect_uri: redirectUri,
      grant_type: "authorization_code",
    }),
  });

  if (!tokenRes.ok) {
    return NextResponse.redirect(`${baseUrl}/login?error=google_token_failed`);
  }

  const tokenData = await tokenRes.json();

  // Fetch user info from Google.
  const userInfoRes = await fetch(
    "https://www.googleapis.com/oauth2/v2/userinfo",
    { headers: { Authorization: `Bearer ${tokenData.access_token}` } }
  );

  if (!userInfoRes.ok) {
    return NextResponse.redirect(`${baseUrl}/login?error=google_userinfo_failed`);
  }

  const googleUser = await userInfoRes.json();

  // Call Go API to find-or-create user.
  const apiRes = await fetch(`${process.env.API_URL}/auth/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      google_id: googleUser.id,
      email: googleUser.email,
      name: googleUser.name || "",
    }),
  });

  if (!apiRes.ok) {
    return NextResponse.redirect(`${baseUrl}/login?error=google_login_failed`);
  }

  const data = await apiRes.json();

  // Build the redirect with Set-Cookie headers manually.
  // NextResponse.redirect + response.cookies.set can silently drop cookies
  // when the redirect target differs from the request origin.
  const maxAge = 60 * 60 * 24 * 30;
  const secure = process.env.NODE_ENV === "production" ? "; Secure" : "";
  const tokenCookie = `token=${data.token}; HttpOnly; SameSite=Lax; Path=/; Max-Age=${maxAge}${secure}`;
  const usernameCookie = `username=${data.username}; HttpOnly; SameSite=Lax; Path=/; Max-Age=${maxAge}${secure}`;

  const headers = new Headers();
  headers.append("Location", baseUrl);
  headers.append("Set-Cookie", tokenCookie);
  headers.append("Set-Cookie", usernameCookie);

  return new Response(null, { status: 302, headers });
}
