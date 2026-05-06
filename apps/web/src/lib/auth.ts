const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export interface AuthUser {
  id: string;
  email: string;
  created_at: string;
}

export async function login(email: string, password: string): Promise<AuthUser> {
  const res = await fetch(`${API_BASE}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.message ?? "Login failed");
  }
  return res.json();
}

export async function register(email: string, password: string): Promise<AuthUser> {
  const res = await fetch(`${API_BASE}/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.message ?? "Registration failed");
  }
  return res.json();
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE}/v1/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

export async function getMe(): Promise<AuthUser | null> {
  const res = await fetch(`${API_BASE}/v1/auth/me`, {
    credentials: "include",
  });
  if (!res.ok) return null;
  return res.json();
}

export async function refresh(): Promise<AuthUser | null> {
  const res = await fetch(`${API_BASE}/v1/auth/refresh`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) return null;
  return res.json();
}
