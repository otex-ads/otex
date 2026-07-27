const API_URL = import.meta.env.VITE_API_URL || "https://api.otexads.com";

export type PricingModel = "CPC" | "CPM" | "CPA";
export type CreativeFormat = "push" | "popunder" | "native" | "banner" | "interstitial";
export type CampaignStatus = "draft" | "active" | "paused" | "ended";

export interface Campaign {
  id: string;
  name: string;
  format: CreativeFormat;
  pricingModel: PricingModel;
  bid: number;
  dailyBudget: number;
  totalBudget: number;
  spent: number;
  impressions: number;
  clicks: number;
  conversions: number;
  targeting: {
    countries: string[];
    devices: string[];
    os: string[];
  };
  creativeId?: string;
  status: CampaignStatus;
  createdAt: string;
}

export interface Creative {
  id: string;
  name: string;
  format: CreativeFormat;
  headline?: string;
  description?: string;
  imageUrl?: string;
  landingUrl: string;
  createdAt: string;
}

export interface WalletTx {
  id: string;
  type: "topup" | "spend" | "payout" | "refund";
  amount: number;
  method?: string;
  phone?: string;
  reference: string;
  campaignId?: string;
  createdAt: string;
}

export interface TopUpResponse {
  authorization_url: string;
  reference: string;
  access_code: string;
}

export interface DailyStat {
  date: string;
  campaignId: string;
  impressions: number;
  clicks: number;
  conversions: number;
  spend: number;
}

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  account_id: string;
  email: string;
  type: string;
}

interface LoginRequest {
  email: string;
  password: string;
}

interface RegisterRequest {
  email: string;
  password: string;
  account_type: string;
  company_name?: string;
}

class ApiClient {
  private token: string | null = null;

  constructor() {
    if (typeof window !== "undefined" && window.localStorage) {
      this.token = localStorage.getItem("access_token");
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== "undefined" && window.localStorage) {
      localStorage.setItem("access_token", token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== "undefined" && window.localStorage) {
      localStorage.removeItem("access_token");
    }
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${API_URL}${endpoint}`;
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(options?.headers || {}),
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    let response: Response;
    try {
      response = await fetch(url, {
        ...options,
        headers,
      });
    } catch (networkErr) {
      // Network failure, CORS block, DNS, or server unreachable.
      console.error(`[api] Network error calling ${options?.method || "GET"} ${url}:`, networkErr);
      throw new Error(
        `Network error reaching ${url}. Check that the API is up and CORS/HTTPS are correct.`,
      );
    }

    // Read the body once as text so we can surface the real error message.
    const raw = await response.text();
    let parsed: unknown = null;
    if (raw) {
      try {
        parsed = JSON.parse(raw);
      } catch {
        parsed = raw; // Non-JSON body (e.g. HTML error page)
      }
    }

    if (!response.ok) {
      const backendMsg =
        parsed && typeof parsed === "object" && "error" in parsed
          ? (parsed as { error: string }).error
          : typeof parsed === "string" && parsed
            ? parsed
            : response.statusText;
      console.error(
        `[api] ${options?.method || "GET"} ${url} failed: ${response.status} — ${backendMsg}`,
      );

      // Expired/invalid session: clear token and send user to login.
      if (response.status === 401 && typeof window !== "undefined") {
        this.clearToken();
        if (!window.location.pathname.startsWith("/auth")) {
          window.location.href = "/auth/login";
        }
      }

      throw new Error(backendMsg || `Request failed with status ${response.status}`);
    }

    // Unwrap the backend's { success: true, data: ... } envelope
    if (parsed && typeof parsed === "object" && "success" in parsed && "data" in parsed) {
      return (parsed as { success: boolean; data: T }).data;
    }
    return parsed as T;
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    this.setToken(res.access_token);
    return res;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
    this.setToken(res.access_token);
    return res;
  }

  async getCampaigns(): Promise<Campaign[]> {
    const campaigns = await this.request<
      Array<{
        id: string;
        name: string;
        pricing_model: string;
        bid_amount_cents: number;
        daily_budget_cents: number;
        total_budget_cents: number;
        status: string;
        timezone: string;
        starts_at?: string;
        ends_at?: string;
        created_at: string;
      }>
    >("/api/v1/campaigns");
    // Transform backend format to frontend format
    return (campaigns || []).map((c) => ({
      id: c.id,
      name: c.name,
      format: "push" as CreativeFormat, // Default format since backend doesn't return it
      pricingModel: c.pricing_model.toUpperCase() as PricingModel,
      bid: c.bid_amount_cents / 100,
      dailyBudget: c.daily_budget_cents / 100,
      totalBudget: c.total_budget_cents / 100,
      spent: 0, // Backend doesn't return this
      impressions: 0,
      clicks: 0,
      conversions: 0,
      targeting: {
        countries: ["KE"], // Default targeting
        devices: ["mobile", "desktop"],
        os: ["android", "ios"],
      },
      status: c.status as CampaignStatus,
      createdAt: c.created_at,
    }));
  }

  async createCampaign(
    data: Omit<
      Campaign,
      "id" | "spent" | "impressions" | "clicks" | "conversions" | "status" | "createdAt"
    >,
  ): Promise<Campaign> {
    // Transform frontend format to backend format
    const backendData = {
      name: data.name,
      pricing_model: data.pricingModel.toLowerCase(),
      bid_amount_cents: Math.round(data.bid * 100),
      daily_budget_cents: Math.round(data.dailyBudget * 100),
      total_budget_cents: Math.round(data.totalBudget * 100),
      timezone: "Africa/Nairobi",
      format: data.format,
      targeting: data.targeting,
      creativeId: data.creativeId,
    };
    const backendCampaign = await this.request<{
      id: string;
      name: string;
      pricing_model: string;
      bid_amount_cents: number;
      daily_budget_cents: number;
      total_budget_cents: number;
      status: string;
      timezone: string;
      created_at: string;
    }>("/api/v1/campaigns", {
      method: "POST",
      body: JSON.stringify(backendData),
    });
    // Transform backend response to frontend format
    return {
      id: backendCampaign.id,
      name: backendCampaign.name,
      format: data.format,
      pricingModel: backendCampaign.pricing_model.toUpperCase() as PricingModel,
      bid: backendCampaign.bid_amount_cents / 100,
      dailyBudget: backendCampaign.daily_budget_cents / 100,
      totalBudget: backendCampaign.total_budget_cents / 100,
      spent: 0,
      impressions: 0,
      clicks: 0,
      conversions: 0,
      targeting: data.targeting,
      creativeId: data.creativeId,
      status: backendCampaign.status as CampaignStatus,
      createdAt: backendCampaign.created_at,
    };
  }

  async updateCampaign(id: string, data: Partial<Campaign>): Promise<Campaign> {
    return this.request<Campaign>(`/api/v1/campaigns/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async getWallet(): Promise<{ balance: number }> {
    const data = await this.request<{ balance_cents?: number; balance?: number }>("/api/v1/wallet");
    return { balance: (data.balance_cents ?? data.balance ?? 0) / 100 };
  }

  async topUpWallet(data: {
    amount_cents: number;
    email: string;
    channel?: string;
  }): Promise<TopUpResponse> {
    return this.request<TopUpResponse>("/api/v1/wallet/topup", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async verifyTopUp(
    reference: string,
  ): Promise<{ status: string; amount: number; new_balance: number }> {
    return this.request<{ status: string; amount: number; new_balance: number }>(
      `/api/v1/wallet/verify?reference=${encodeURIComponent(reference)}`,
    );
  }

  async getTransactions(): Promise<WalletTx[]> {
    const txs = await this.request<
      Array<{
        id: string;
        type: string;
        amount_cents?: number;
        reference?: string;
        created_at?: string;
        createdAt?: string;
      }>
    >("/api/v1/wallet/transactions");
    return (txs || []).map((t) => ({
      id: t.id,
      type: t.type,
      amount: (t.amount_cents ?? 0) / 100,
      reference: t.reference || "",
      createdAt: t.created_at || t.createdAt || "",
    }));
  }

  async getStats(): Promise<DailyStat[]> {
    // Backend may not have this endpoint yet, return empty for now
    return [];
  }
}

export const api = new ApiClient();

// Standalone helpers for auth pages
export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

export function setToken(token: string) {
  api.setToken(token);
}

export function clearToken() {
  api.clearToken();
}

export function getUser(): { email: string } | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem("user");
  return raw ? JSON.parse(raw) : null;
}

export function setUser(user: { email: string }) {
  if (typeof window !== "undefined") {
    localStorage.setItem("user", JSON.stringify(user));
  }
}
