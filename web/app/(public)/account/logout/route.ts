import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export async function POST() {
  const jar = await cookies();
  const token = jar.get("user_session")?.value;
  await fetch(`${API_BASE}/api/v1/auth/logout`, {
    method: "POST",
    headers: token ? { Cookie: `user_session=${token}` } : undefined,
  }).catch(() => undefined);
  jar.delete("user_session");
  redirect("/");
}
