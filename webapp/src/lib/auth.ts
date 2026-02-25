import { cookies } from "next/headers";

export type AuthUser = {
  user_id: string;
  username: string;
  is_moderator: boolean;
  email_verified: boolean;
};

function decodeJWT(token: string): Record<string, unknown> | null {
  try {
    const payload = token.split(".")[1];
    return JSON.parse(Buffer.from(payload, "base64url").toString("utf-8"));
  } catch {
    return null;
  }
}

export async function getToken(): Promise<string | null> {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? null;
}

export async function getUser(): Promise<AuthUser | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return null;

  const payload = decodeJWT(token);
  if (!payload) return null;

  const exp = payload.exp as number | undefined;
  if (exp && Date.now() / 1000 > exp) return null;

  // PocketBase token might not have username, so check cookie
  let username = payload.username as string | undefined;
  if (!username) {
      username = cookieStore.get("username")?.value;
  }

  // If still no username (and it's required by type), we might need to fallback or return null?
  // But strictly speaking, if we just registered/logged in, the cookie should be there.
  // If the user clears cookies partly, they might be logged out effectively.
  if (!username) return null;

  return {
    user_id: (payload.id as string) || (payload.sub as string),
    username: username,
    is_moderator: (payload.is_moderator as boolean) ?? false,
    email_verified: (payload.email_verified as boolean) ?? false,
  };
}
