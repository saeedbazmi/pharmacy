import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export async function POST() {
  const jar = await cookies();
  const token = jar.get("ops_session")?.value;
  await fetch(`${API_BASE}/api/v1/ops/logout`, {
    method: "POST",
    headers: token ? { Cookie: `ops_session=${token}` } : undefined,
  }).catch(() => undefined);
  jar.delete("ops_session");
  redirect("/ops/login");
}
