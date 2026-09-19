import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

async function proxy(req: NextRequest, path: string[]) {
  const url = `${API_BASE}/api/v1/admin/${path.join("/")}${req.nextUrl.search}`;
  const jar = await cookies();
  const token = jar.get("ops_session")?.value;
  const headers = new Headers();
  headers.set("Accept", "application/json");
  if (req.headers.get("content-type")) {
    headers.set("Content-Type", req.headers.get("content-type")!);
  }
  if (token) {
    headers.set("Cookie", `ops_session=${token}`);
  }
  const body = req.method === "GET" || req.method === "HEAD" ? undefined : await req.text();
  const upstream = await fetch(url, { method: req.method, headers, body });
  const text = await upstream.text();
  const out = new NextResponse(text, { status: upstream.status });
  const ct = upstream.headers.get("Content-Type");
  if (ct) out.headers.set("Content-Type", ct);
  return out;
}

export async function GET(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  return proxy(req, (await ctx.params).path);
}
export async function POST(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  return proxy(req, (await ctx.params).path);
}
export async function PATCH(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  return proxy(req, (await ctx.params).path);
}
export async function PUT(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  return proxy(req, (await ctx.params).path);
}
export async function DELETE(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  return proxy(req, (await ctx.params).path);
}
