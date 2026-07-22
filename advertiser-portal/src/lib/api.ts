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

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }

    const json = await response.json();
    // Unwrap the backend's { success: true, data: ... } envelope
    if (json && typeof json === "object" && "success" in json && "data" in json) {
      return (json as any).data as T;
    }
    return json as T;
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
    const campaigns = await this.request<any[]>("/api/v1/campaigns");
    // Transform backend format to frontend format
    return campaigns.map((c) => ({
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

  async createCampaign(data: Omit<Campaign, "id" | "spent" | "impressions" | "clicks" | "conversions" | "status" | "createdAt">): Promise<Campaign> {
    // Transform frontend format to backend format
    const backendData = {
      name: data.name,
      pricing_model: data.pricingModel.toLowerCase(),
      bid_amount_cents: Math.round(data.bid * 100),
      daily_budget_cents: Math.round(data.dailyBudget * 100),
      total_budget_cents: Math.round(data.totalBudget * 100),
      timezone: "Africa/Nairobi",
    };
    return this.request<Campaign>("/api/v1/campaigns", {
      method: "POST",
      body: JSON.stringify(backendData),
    });
  }

  async updateCampaign(id: string, data: Partial<Campaign>): Promise<Campaign> {
    return this.request<Campaign>(`/api/v1/campaigns/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async getWallet(): Promise<{ balance: number }> {
    const data = await this.request<any>("/api/v1/wallet");
    return { balance: (data.balance_cents ?? data.balance ?? 0) / 100 };
  }

  async topUpWallet(data: { amount_cents: number; email: string; channel?: string }): Promise<TopUpResponse> {
    return this.request<TopUpResponse>("/api/v1/wallet/topup", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async verifyTopUp(reference: string): Promise<{ status: string; amount: number; new_balance: number }> {
    return this.request<{ status: string; amount: number; new_balance: number }>(
      `/api/v1/wallet/verify?reference=${encodeURIComponent(reference)}`
    );
  }

  async getTransactions(): Promise<WalletTx[]> {
    const txs = await this.request<any[]>("/api/v1/wallet/transactions");
    return (txs || []).map((t: any) => ({
      id: t.id,
      type: t.type,
      amount: (t.amount_cents ?? 0) / 100,
      reference: t.reference || '',
      createdAt: t.created_at || t.createdAt || '',
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
