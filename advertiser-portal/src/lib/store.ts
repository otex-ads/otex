// Store backed by API for the advertiser portal
import { useSyncExternalStore } from "react";
import { api, type Campaign, type Creative, type WalletTx, type DailyStat } from "./api";

interface State {
  balance: number;
  campaigns: Campaign[];
  creatives: Creative[];
  wallet: WalletTx[];
  stats: DailyStat[];
}

let state: State = {
  balance: 0,
  campaigns: [],
  creatives: [],
  wallet: [],
  stats: [],
};
const listeners = new Set<() => void>();

function notify() {
  listeners.forEach((l) => l());
}

function set(mut: (s: State) => State) {
  state = mut(state);
  notify();
}

async function fetchFromAPI() {
  try {
    const [campaigns, wallet, stats] = await Promise.all([
      api.getCampaigns(),
      api.getTransactions(),
      api.getStats(),
    ]);
    const walletData = await api.getWallet();
    set((s) => ({
      ...s,
      campaigns,
      wallet,
      stats,
      balance: walletData.balance,
    }));
  } catch (error) {
    console.error("Failed to fetch from API:", error);
  }
}

export const store = {
  get: () => state,
  subscribe: (fn: () => void) => { listeners.add(fn); return () => listeners.delete(fn); },
  refresh: fetchFromAPI,

  async addCampaign(input: Omit<Campaign, "id" | "spent" | "impressions" | "clicks" | "conversions" | "status" | "createdAt"> & { status?: CampaignStatus }) {
    const c = await api.createCampaign(input);
    set((s) => ({ ...s, campaigns: [c, ...s.campaigns] }));
    return c;
  },
  async updateCampaign(id: string, patch: Partial<Campaign>) {
    const c = await api.updateCampaign(id, patch);
    set((s) => ({ ...s, campaigns: s.campaigns.map((x) => x.id === id ? c : x) }));
    return c;
  },
  removeCampaign(id: string) {
    set((s) => ({ ...s, campaigns: s.campaigns.filter((c) => c.id !== id) }));
  },
  addCreative(input: Omit<Creative, "id" | "createdAt">) {
    const c: Creative = { ...input, id: `cre_${Date.now()}`, createdAt: new Date().toISOString() };
    set((s) => ({ ...s, creatives: [c, ...s.creatives] }));
    return c;
  },
  removeCreative(id: string) {
    set((s) => ({ ...s, creatives: s.creatives.filter((c) => c.id !== id) }));
  },
  async topUp(amountKES: number, email: string, channel?: string) {
    const resp = await api.topUpWallet({
      amount_cents: Math.round(amountKES * 100),
      email,
      channel,
    });
    // Redirect user to Paystack checkout
    if (resp.authorization_url) {
      window.location.href = resp.authorization_url;
    }
    return resp;
  },
  async verifyTopUp(reference: string) {
    const resp = await api.verifyTopUp(reference);
    if (resp.status === "success") {
      set((s) => ({ ...s, balance: resp.new_balance / 100 }));
    }
    return resp;
  },
};

export function useStore<T>(selector: (s: State) => T): T {
  return useSyncExternalStore(store.subscribe, () => selector(store.get()), () => selector(store.get()));
}
