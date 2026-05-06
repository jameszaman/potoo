import { NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/sign-in", "/sign-up"];
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  const accessToken = req.cookies.get("nl_access");
  const refreshToken = req.cookies.get("nl_refresh");

  // Access token present — allow through.
  if (accessToken?.value) {
    return NextResponse.next();
  }

  // No access token but refresh token exists — attempt silent refresh.
  if (refreshToken?.value) {
    try {
      const res = await fetch(`${API_BASE}/v1/auth/refresh`, {
        method: "POST",
        headers: { Cookie: `nl_refresh=${refreshToken.value}` },
      });

      if (res.ok) {
        const response = NextResponse.next();
        // Forward Set-Cookie headers from the API to the browser.
        res.headers.getSetCookie().forEach((cookie) => {
          response.headers.append("Set-Cookie", cookie);
        });
        return response;
      }
    } catch {
      // Refresh failed — fall through to redirect.
    }
  }

  const signIn = req.nextUrl.clone();
  signIn.pathname = "/sign-in";
  return NextResponse.redirect(signIn);
}

export const config = {
  matcher: ["/((?!_next|favicon.ico|[^?]*\\.(?:css|js|png|jpg|svg|ico)).*)"],
};
