// Publisher API client. All requests attach the JWT from localStorage.
// Base URL is a plain-HTTP dev endpoint; browsers on HTTPS may block mixed
// content — errors surface via TanStack Query as network errors.
export const API_BASE = "https://api.otexads.com";
export const TOKEN_KEY = "pub_token";
export const USER_KEY = "pub_user";

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

async function apiFetch<T>(path: string, opts: RequestInit & { auth?: boolean } = {}): Promise<T> {
  const { auth = true, headers, body, ...rest } = opts;
  const h: Record<string, string> = {
    Accept: "application/json",
    ...(headers as Record<string, string>),
  };
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
    return (data as { success: boolean; data: T }).data;
  }
  return data as T;
}

function safeJson(t: string) {
  try {
    return JSON.parse(t);
  } catch {
    return t;
  }
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
export type AdFormat =
  | "banner"
  | "native"
  | "push"
  | "popunder"
  | "interstitial"
  | "in_page_push";
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
  countries?: string[];
  deviceTypes?: string[];
  os?: string[];
  browsers?: string[];
  carriers?: string[];
  connectionTypes?: string[];
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

export type PayoutStatus = "pending" | "processing" | "paid" | "rejected" | "failed";
export interface Payout {
  id: string;
  amount: number;
  method: string;
  status: PayoutStatus;
  createdAt: string;
  reference?: string;
}

export interface TransferRecipient {
  id: string;
  account_id: string;
  recipient_code: string;
  type: string;
  name: string;
  phone?: string;
  email?: string;
  bank_code?: string;
  currency: string;
  is_default: boolean;
  created_at: string;
}

export interface RecipientStatus {
  configured: boolean;
  recipient?: TransferRecipient;
}

// -------- Endpoints --------
export const api = {
  // Auth
  login: (email: string, password: string) =>
    apiFetch<{
      access_token: string;
      refresh_token: string;
      account_id: string;
      email: string;
      type: string;
    }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
      auth: false,
    }),
  register: (
    email: string,
    password: string,
    accountType: string,
    companyName?: string,
  ) =>
    apiFetch<{
      access_token: string;
      refresh_token: string;
      account_id: string;
      email: string;
      type: string;
    }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({
        email,
        password,
        account_type: accountType,
        company_name: companyName,
      }),
      auth: false,
    }),
  requestPasswordReset: (data: { email: string }) =>
    apiFetch<void>("/api/v1/auth/password-reset/request", {
      method: "POST",
      body: JSON.stringify(data),
      auth: false,
    }),
  resetPassword: (data: { token: string; new_password: string }) =>
    apiFetch<void>("/api/v1/auth/password-reset/confirm", {
      method: "POST",
      body: JSON.stringify(data),
      auth: false,
    }),

  // Sites
  listSites: () => apiFetch<Site[]>("/api/v1/sites"),
  createSite: (input: Omit<Site, "id" | "revenue" | "impressions" | "status">) =>
    apiFetch<Site>("/api/v1/sites", { method: "POST", body: JSON.stringify(input) }),
  updateSite: (id: string, patch: Partial<Site>) =>
    apiFetch<Site>(`/api/v1/sites/${id}`, { method: "PUT", body: JSON.stringify(patch) }),
  deleteSite: (id: string) => apiFetch<void>(`/api/v1/sites/${id}`, { method: "DELETE" }),

  // Zones
  listZones: (siteId?: string) =>
    apiFetch<Zone[]>(`/api/v1/zones${siteId ? `?siteId=${siteId}` : ""}`),
  createZone: (
    input: Omit<Zone, "id" | "revenue" | "impressions" | "clicks" | "status" | "siteName">,
  ) => apiFetch<Zone>("/api/v1/zones", { method: "POST", body: JSON.stringify(input) }),
  updateZone: (id: string, patch: Partial<Zone>) =>
    apiFetch<Zone>(`/api/v1/zones/${id}`, { method: "PUT", body: JSON.stringify(patch) }),
  deleteZone: (id: string) => apiFetch<void>(`/api/v1/zones/${id}`, { method: "DELETE" }),

  // Stats - not implemented yet, return empty
  stats: (params: { from?: string; to?: string; groupBy?: string } = {}) => {
    return Promise.resolve({
      summary: { impressions: 0, clicks: 0, ctr: 0, ecpm: 0, revenue: 0, fillRate: 0 },
      daily: [],
    });
  },

  // Payouts
  balance: async () => {
    const data = await apiFetch<{
      available?: number;
      pending?: number;
      totalPaid?: number;
      total_paid?: number;
    }>("/api/v1/payouts/balance");
    return {
      available: (data.available ?? 0) / 100,
      pending: (data.pending ?? 0) / 100,
      totalPaid: (data.totalPaid ?? data.total_paid ?? 0) / 100,
    } as Balance;
  },
  listPayouts: async () => {
    const data = await apiFetch<
      Array<{
        id: string;
        amount_cents?: number;
        status: string;
        requested_at?: string;
        created_at?: string;
        paystack_reference?: string;
        mpesa_receipt?: string;
      }>
    >("/api/v1/payouts");
    return (data || []).map((p) => ({
      id: p.id,
      amount: (p.amount_cents ?? 0) / 100,
      method: "M-Pesa",
      status: p.status as PayoutStatus,
      createdAt: p.requested_at || p.created_at || "",
      reference: p.paystack_reference || p.mpesa_receipt || undefined,
    })) as Payout[];
  },
  requestPayout: (input: { amount: number; method: string; details?: Record<string, string> }) =>
    apiFetch<Payout>("/api/v1/payouts", {
      method: "POST",
      body: JSON.stringify({ amount_cents: Math.round(input.amount * 100) }),
    }),

  // Transfer Recipients
  getRecipientStatus: () => apiFetch<RecipientStatus>("/api/v1/recipients/default"),
  listRecipients: () => apiFetch<TransferRecipient[]>("/api/v1/recipients"),
  saveRecipient: (input: { name: string; phone: string; email?: string; bank_code?: string }) =>
    apiFetch<TransferRecipient>("/api/v1/recipients", {
      method: "POST",
      body: JSON.stringify(input),
    }),
};
