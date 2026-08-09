// Store backed by API for the advertiser portal
import { useSyncExternalStore } from "react";
import { api, type Campaign, type CampaignStatus, type Creative, type WalletTx, type DailyStat } from "./api";

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
    const [campaigns, creatives, wallet] = await Promise.all([
      api.getCampaigns(),
      api.listCreatives(),
      api.getTransactions(),
    ]);
    const walletData = await api.getWallet();

    // Fetch stats for the first campaign (or all campaigns)
    let stats: DailyStat[] = [];
    if (campaigns.length > 0) {
      stats = await api.getStats(campaigns[0].id, 30);
    }

    set((s) => ({
      ...s,
      campaigns,
      creatives,
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
  subscribe: (fn: () => void) => {
    listeners.add(fn);
    return () => listeners.delete(fn);
  },
  refresh: fetchFromAPI,

  async addCampaign(
    input: Omit<
      Campaign,
      "id" | "spent" | "impressions" | "clicks" | "conversions" | "status" | "createdAt"
    > & { status?: CampaignStatus },
  ) {
    const c = await api.createCampaign(input);
    set((s) => ({ ...s, campaigns: [c, ...s.campaigns] }));
    return c;
  },
  async updateCampaign(id: string, patch: Partial<Campaign>) {
    const c = await api.updateCampaign(id, patch);
    set((s) => ({ ...s, campaigns: s.campaigns.map((x) => (x.id === id ? c : x)) }));
    return c;
  },
  removeCampaign(id: string) {
    set((s) => ({ ...s, campaigns: s.campaigns.filter((c) => c.id !== id) }));
  },
  async addCreative(input: Omit<Creative, "id" | "createdAt"> & { campaignId: string }) {
    const c = await api.createCreative({
      campaignId: input.campaignId,
      format: input.format,
      title: input.headline,
      body: input.description,
      imageUrl: input.imageUrl,
      clickUrl: input.landingUrl,
    });
    set((s) => ({ ...s, creatives: [c, ...s.creatives] }));
    return c;
  },
  async removeCreative(id: string) {
    await api.deleteCreative(id);
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
  async refreshStats(campaignId: string, days: number = 30) {
    const stats = await api.getStats(campaignId, days);
    set((s) => ({ ...s, stats }));
  },
};

export function useStore<T>(selector: (s: State) => T): T {
  return useSyncExternalStore(
    store.subscribe,
    () => selector(store.get()),
    () => selector(store.get()),
  );
}
