const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export interface AuthUser {
  id: string;
  email: string;
  created_at: string;
}

export interface OrgSummary {
  id: string;
  name: string;
  slug: string;
  type: "platform" | "customer";
  role: "owner" | "member";
}

async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data?.error?.message ?? `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export async function setup(name: string, email: string, password: string): Promise<AuthUser> {
  return apiPost<AuthUser>("/v1/setup", { name, email, password });
}

export async function login(email: string, password: string): Promise<AuthUser> {
  return apiPost<AuthUser>("/v1/auth/login", { email, password });
}

export async function register(email: string, password: string): Promise<AuthUser> {
  return apiPost<AuthUser>("/v1/auth/register", { email, password });
}

export async function logout(): Promise<void> {
  await apiPost<void>("/v1/auth/logout");
}

export async function getMe(): Promise<AuthUser | null> {
  const res = await fetch(`${API_BASE}/v1/auth/me`, { credentials: "include" });
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

export async function listMyOrgs(): Promise<OrgSummary[]> {
  const res = await fetch(`${API_BASE}/v1/me/orgs`, { credentials: "include" });
  if (!res.ok) return [];
  const data = await res.json();
  return data.data ?? [];
}

export async function selectOrg(orgId: string): Promise<AuthUser> {
  return apiPost<AuthUser>("/v1/auth/select-org", { org_id: orgId });
}
