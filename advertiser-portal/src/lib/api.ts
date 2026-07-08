const API_URL = import.meta.env.VITE_API_URL || "http://167.233.171.202:8080";

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
  type: "topup" | "spend";
  amount: number;
  method?: "mpesa";
  phone?: string;
  reference: string;
  campaignId?: string;
  createdAt: string;
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
  accessToken: string;
  refreshToken: string;
}

interface LoginRequest {
  email: string;
  password: string;
}

interface RegisterRequest {
  email: string;
  password: string;
  name: string;
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

    return response.json();
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    this.setToken(res.accessToken);
    return res;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
    this.setToken(res.accessToken);
    return res;
  }

  async getCampaigns(): Promise<Campaign[]> {
    return this.request<Campaign[]>("/api/v1/campaigns");
  }

  async createCampaign(data: Omit<Campaign, "id" | "spent" | "impressions" | "clicks" | "conversions" | "status" | "createdAt">): Promise<Campaign> {
    return this.request<Campaign>("/api/v1/campaigns", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateCampaign(id: string, data: Partial<Campaign>): Promise<Campaign> {
    return this.request<Campaign>(`/api/v1/campaigns/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async getWallet(): Promise<{ balance: number }> {
    return this.request<{ balance: number }>("/api/v1/wallet");
  }

  async topUpWallet(data: { amount: number; phone: string }): Promise<WalletTx> {
    return this.request<WalletTx>("/api/v1/wallet/topup", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getTransactions(): Promise<WalletTx[]> {
    return this.request<WalletTx[]>("/api/v1/wallet/transactions");
  }

  async getStats(): Promise<DailyStat[]> {
    // Backend may not have this endpoint yet, return empty for now
    return [];
  }
}

export const api = new ApiClient();
