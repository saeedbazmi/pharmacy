import { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(
  /\/+$/,
  "",
);

/**
 * Proxies GET /go/{offerId} to the API without following the 302, so the
 * browser lands on the pharmacy URL stored in the database.
 */
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ offerId: string }> },
) {
  const { offerId } = await params;
  if (!/^\d{1,18}$/.test(offerId)) {
    return gone();
  }

  const upstream = await fetch(`${API_BASE}/go/${offerId}`, {
    redirect: "manual",
    headers: {
      Accept: "text/html",
      Referer: request.headers.get("referer") ?? "",
    },
    cache: "no-store",
  });

  const location = upstream.headers.get("location");
  if (upstream.status === 302 && location) {
    return new Response(null, {
      status: 302,
      headers: {
        Location: location,
        "X-Robots-Tag": "noindex, nofollow",
      },
    });
  }

  const body = await upstream.text();
  return new Response(body, {
    status: upstream.status,
    headers: {
      "content-type":
        upstream.headers.get("content-type") ?? "text/html; charset=utf-8",
      "X-Robots-Tag": "noindex, nofollow",
    },
  });
}

function gone() {
  return new Response(
    `<!doctype html>
<html lang="fa" dir="rtl">
<meta charset="utf-8">
<title>پیشنهاد نامعتبر</title>
<body>
<main>
<h1>این پیشنهاد دیگر معتبر نیست</h1>
<p>لینک خرید منقضی شده یا این کالا در داروخانه موجود نیست.</p>
<p><a href="/">بازگشت به صفحه اصلی</a></p>
</main>
</body>
</html>`,
    {
      status: 404,
      headers: {
        "content-type": "text/html; charset=utf-8",
        "X-Robots-Tag": "noindex, nofollow",
      },
    },
  );
}
