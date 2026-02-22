import { cookies } from "next/headers";

export type AuthUser = {
  user_id: string;
  username: string;
};

function decodeJWT(token: string): Record<string, unknown> | null {
  try {
    const payload = token.split(".")[1];
    return JSON.parse(Buffer.from(payload, "base64url").toString("utf-8"));
  } catch {
    return null;
  }
}

export async function getUser(): Promise<AuthUser | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;
  if (!token) return null;

  const payload = decodeJWT(token);
  if (!payload) return null;

  const exp = payload.exp as number | undefined;
  if (exp && Date.now() / 1000 > exp) return null;

  return {
    user_id: payload.sub as string,
    username: payload.username as string,
  };
}
