const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function doFetch(path: string, options: RequestInit): Promise<Response> {
  return fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {}),
    },
  });
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  let res = await doFetch(path, options);

  // Access token expired — attempt a silent refresh then retry once.
  if (res.status === 401) {
    const refreshed = await fetch(`${API_BASE}/v1/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });

    if (refreshed.ok) {
      res = await doFetch(path, options);
    } else {
      // Refresh token also expired — kick to sign-in.
      if (typeof window !== "undefined") {
        window.location.href = "/sign-in";
      }
      throw new Error("Session expired");
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.message ?? `HTTP ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
