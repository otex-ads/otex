// Publisher API client. All requests attach the JWT from localStorage.
// Base URL is a plain-HTTP dev endpoint; browsers on HTTPS may block mixed
// content — errors surface via TanStack Query as network errors.
export const API_BASE = "https://api.otexads.com";
export const TOKEN_KEY = "admin_token";
export const USER_KEY = "admin_user";

export function getToken(): string | null {
  if (typeof localStorage === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  if (typeof window !== "undefined" && window.localStorage) {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function clearToken() {
  if (typeof window !== "undefined" && window.localStorage) {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
}

export function getUser(): { email: string; name?: string } | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function setUser(u: { email: string; name?: string }) {
  if (typeof window !== "undefined" && window.localStorage) {
    localStorage.setItem(USER_KEY, JSON.stringify(u));
  }
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function apiFetch<T>(
  path: string,
  opts: RequestInit & { auth?: boolean } = {},
): Promise<T> {
  const { auth = true, headers, body, ...rest } = opts;
  const h: Record<string, string> = { Accept: "application/json", ...(headers as Record<string, string>) };
  if (body && !(body instanceof FormData)) h["Content-Type"] = "application/json";
  if (auth) {
    const t = getToken();
    if (t) h.Authorization = `Bearer ${t}`;
  }
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...rest, headers: h, body });
  } catch (e) {
    throw new ApiError(0, "Network error — check your connection.");
  }
  if (res.status === 401) {
    clearToken();
    if (typeof window !== "undefined" && !window.location.pathname.startsWith("/auth")) {
      window.location.href = "/auth/login";
    }
    throw new ApiError(401, "Session expired. Please sign in again.");
  }
  const text = await res.text();
  const data = text ? safeJson(text) : null;
  if (!res.ok) {
    const msg = (data && (data.message || data.error)) || `Request failed (${res.status})`;
    throw new ApiError(res.status, msg);
  }
  // Unwrap the backend's { success: true, data: ... } envelope
  if (data && typeof data === "object" && "success" in data && "data" in data) {
    return (data as any).data as T;
  }
  return data as T;
}

function safeJson(t: string) {
  try { return JSON.parse(t); } catch { return t; }
}

// -------- Domain types --------
export type SiteStatus = "active" | "paused";
export interface Site {
  id: string;
  name: string;
  domain: string;
  category?: string;
  description?: string;
  status: SiteStatus;
  revenue: number;
  impressions: number;
  createdAt?: string;
}

export type ZoneStatus = "active" | "paused";
export type AdSize = "728x90" | "300x250" | "160x600" | "970x250" | "320x50";
export type AdFormat = "display" | "video" | "native";
export interface Zone {
  id: string;
  name: string;
  siteId: string;
  siteName?: string;
  size: AdSize;
  format: AdFormat;
  status: ZoneStatus;
  revenue: number;
  impressions: number;
  clicks: number;
}

export interface StatsSummary {
  impressions: number;
  clicks: number;
  ctr: number;
  ecpm: number;
  revenue: number;
  fillRate: number;
}

export interface StatsDaily {
  date: string;
  impressions: number;
  clicks: number;
  revenue: number;
}

export interface StatsResponse {
  summary: StatsSummary;
  daily: StatsDaily[];
  byCountry?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
  byDevice?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
  byBrowser?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
  byOS?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
  bySite?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
  byZone?: Array<{ key: string; impressions: number; clicks: number; revenue: number }>;
}

export interface Balance {
  available: number;
  pending: number;
  totalPaid: number;
}

export type PayoutStatus = "pending" | "processing" | "paid" | "rejected";
export interface Payout {
  id: string;
  amount: number;
  method: string;
  status: PayoutStatus;
  createdAt: string;
  reference?: string;
}

// -------- Endpoints --------
export const api = {
  // Generic fetch for admin endpoints
  fetch: <T = any>(path: string, opts?: RequestInit) => apiFetch<T>(`/api/v1${path}`, opts),

  // Auth
  login: (email: string, password: string) =>
    apiFetch<{ access_token: string; refresh_token: string; account_id: string; email: string; type: string }>(
      "/api/auth/login",
      { method: "POST", body: JSON.stringify({ email, password }), auth: false },
    ),
  register: (email: string, password: string, accountType: string, companyName?: string) =>
    apiFetch<{ access_token: string; refresh_token: string; account_id: string; email: string; type: string }>(
      "/api/auth/register",
      { method: "POST", body: JSON.stringify({ email, password, account_type: accountType, company_name: companyName }), auth: false },
    ),

  // Admin
  getStats: () => apiFetch<any>("/api/v1/admin/stats"),
  listUsers: () => apiFetch<any[]>("/api/v1/admin/users"),
  listCampaigns: () => apiFetch<any[]>("/api/v1/admin/campaigns"),
  updateCampaignStatus: (id: string, status: string) =>
    apiFetch<any>(`/api/v1/admin/campaigns/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) }),
  updateAccountStatus: (id: string, status: string) =>
    apiFetch<any>(`/api/v1/admin/accounts/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) }),

  // Sites
  listSites: () => apiFetch<Site[]>("/api/v1/sites"),
  createSite: (input: Omit<Site, "id" | "revenue" | "impressions" | "status">) =>
    apiFetch<Site>("/api/v1/sites", { method: "POST", body: JSON.stringify(input) }),
  updateSite: (id: string, patch: Partial<Site>) =>
    apiFetch<Site>(`/api/v1/sites/${id}`, { method: "PUT", body: JSON.stringify(patch) }),
  deleteSite: (id: string) =>
    apiFetch<void>(`/api/v1/sites/${id}`, { method: "DELETE" }),

  // Zones
  listZones: (siteId?: string) =>
    apiFetch<Zone[]>(`/api/v1/zones${siteId ? `?siteId=${siteId}` : ""}`),
  createZone: (input: Omit<Zone, "id" | "revenue" | "impressions" | "clicks" | "status" | "siteName">) =>
    apiFetch<Zone>("/api/v1/zones", { method: "POST", body: JSON.stringify(input) }),
  updateZone: (id: string, patch: Partial<Zone>) =>
    apiFetch<Zone>(`/api/v1/zones/${id}`, { method: "PUT", body: JSON.stringify(patch) }),
  deleteZone: (id: string) =>
    apiFetch<void>(`/api/v1/zones/${id}`, { method: "DELETE" }),

  // Stats - not implemented yet, return empty
  stats: (params: { from?: string; to?: string; groupBy?: string } = {}) => {
    return Promise.resolve({ summary: { impressions: 0, clicks: 0, ctr: 0, ecpm: 0, revenue: 0, fillRate: 0 }, daily: [] });
  },

  // Payouts
  balance: () => apiFetch<Balance>("/api/v1/payouts/balance"),
  listPayouts: () => apiFetch<Payout[]>("/api/v1/payouts"),
  requestPayout: (input: { amount: number; method: string; details?: Record<string, string> }) =>
    apiFetch<Payout>("/api/v1/payouts", { method: "POST", body: JSON.stringify(input) }),

  // Admin Financial endpoints
  getFinancials: () => apiFetch<{
    total_deposits: number;
    total_publisher_payouts: number;
    pending_payouts: number;
    total_platform_fees: number;
    total_ad_spend: number;
  }>("/api/v1/admin/financials"),

  getGatewayBalance: () => apiFetch<{
    available: boolean;
    balances: Array<{ currency: string; balance: number }>;
  }>("/api/v1/admin/financials/gateway-balance"),

  listDeposits: () => apiFetch<Array<{
    id: string;
    account_id: string;
    type: string;
    amount_cents: number;
    reference: string;
    created_at: string;
  }>>("/api/v1/admin/financials/deposits"),

  listAllPayouts: () => apiFetch<Array<{
    id: string;
    publisher_id: string;
    amount_cents: number;
    status: string;
    mpesa_receipt: string | null;
    requested_at: string;
    processed_at: string | null;
  }>>("/api/v1/admin/financials/payouts"),

  retryPayout: (id: string) =>
    apiFetch<{ status: string; message: string }>(`/api/v1/admin/financials/payouts/${id}/retry`, { method: "POST" }),
};
