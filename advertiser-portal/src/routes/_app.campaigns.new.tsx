import { createFileRoute, useNavigate, Link } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { ArrowLeft } from "lucide-react";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { store, useStore } from "@/lib/store";
import { type CreativeFormat, type PricingModel } from "@/lib/api";

export const Route = createFileRoute("/_app/campaigns/new")({
  component: NewCampaignPage,
});

const FORMATS: { value: CreativeFormat; label: string; desc: string }[] = [
  { value: "push", label: "Push", desc: "Native OS-style notifications" },
  { value: "popunder", label: "Popunder", desc: "Full-page background window" },
  { value: "native", label: "Native", desc: "Blends with publisher content" },
  { value: "banner", label: "Banner", desc: "Standard IAB display sizes" },
  { value: "interstitial", label: "Interstitial", desc: "Full-screen mobile overlay" },
];

const COUNTRIES = ["KE", "TZ", "UG", "NG", "ZA", "US", "GB", "IN", "BR"];
const DEVICES = ["desktop", "mobile", "tablet"] as const;
const OSES = ["android", "ios", "windows", "macos", "linux"] as const;

function Chip({ selected, onClick, children }: { selected: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full border px-3 py-1 text-xs font-medium capitalize transition-colors ${
        selected ? "border-primary bg-primary text-primary-foreground" : "border-border bg-card text-muted-foreground hover:text-foreground"
      }`}
    >
      {children}
    </button>
  );
}

function NewCampaignPage() {
  const nav = useNavigate();
  const creatives = useStore((s) => s.creatives) || [];

  const [name, setName] = useState("");
  const [format, setFormat] = useState<CreativeFormat>("push");
  const [pricingModel, setPricingModel] = useState<PricingModel>("CPC");
  const [bid, setBid] = useState<number>(3);
  const [dailyBudget, setDailyBudget] = useState<number>(1000);
  const [totalBudget, setTotalBudget] = useState<number>(20000);
  const [countries, setCountries] = useState<string[]>(["KE"]);
  const [devices, setDevices] = useState<string[]>(["mobile", "desktop"]);
  const [os, setOs] = useState<string[]>(["android", "ios"]);
  const [creativeId, setCreativeId] = useState<string>("");

  const toggle = <T,>(arr: T[], v: T, setter: (a: T[]) => void) => {
    setter(arr.includes(v) ? arr.filter((x) => x !== v) : [...arr, v]);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return toast.error("Name is required");
    if (countries.length === 0) return toast.error("Pick at least one country");
    try {
      await store.addCampaign({
        name: name.trim(),
        format, pricingModel, bid, dailyBudget, totalBudget,
        targeting: { countries, devices, os },
        creativeId: creativeId || undefined,
      });
      toast.success("Campaign launched");
      nav({ to: "/campaigns" });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to create campaign";
      toast.error(msg);
    }
  };

  return (
    <div className="mx-auto max-w-3xl">
      <Link to="/campaigns" className="mb-6 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="size-4" /> Back to campaigns
      </Link>
      <h1 className="display text-4xl font-normal tracking-tight text-foreground">New campaign</h1>
      <p className="mt-2 text-sm text-muted-foreground">Configure targeting, pricing, and budget. You can pause or edit anytime.</p>

      <form onSubmit={submit} className="mt-8 space-y-6">
        <SurfaceCard>
          <SectionLabel>Basics</SectionLabel>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium text-foreground">Campaign name</div>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Kenya · Fintech Push Q3"
              className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring/40"
            />
          </label>

          <div className="mt-5">
            <div className="mb-2 text-xs font-medium text-foreground">Ad format</div>
            <div className="grid gap-2 md:grid-cols-5">
              {FORMATS.map((f) => (
                <button
                  key={f.value}
                  type="button"
                  onClick={() => setFormat(f.value)}
                  className={`rounded-xl border p-3 text-left transition-all ${
                    format === f.value ? "border-primary bg-primary/5" : "border-border bg-card hover:border-primary/40"
                  }`}
                >
                  <div className="text-sm font-medium text-foreground">{f.label}</div>
                  <div className="mt-1 text-[11px] leading-snug text-muted-foreground">{f.desc}</div>
                </button>
              ))}
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <SectionLabel>Pricing & budget</SectionLabel>
          <div className="grid gap-4 md:grid-cols-3">
            <label className="block">
              <div className="mb-1.5 text-xs font-medium">Pricing model</div>
              <select
                value={pricingModel}
                onChange={(e) => setPricingModel(e.target.value as PricingModel)}
                className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm"
              >
                <option value="CPC">CPC · Cost per click</option>
                <option value="CPM">CPM · Cost per 1K impressions</option>
                <option value="CPA">CPA · Cost per action</option>
              </select>
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium">Bid amount (KES)</div>
              <input type="number" min={0.1} step={0.1} value={bid} onChange={(e) => setBid(+e.target.value)}
                className="num h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" />
            </label>
            <div />
            <label className="block">
              <div className="mb-1.5 text-xs font-medium">Daily budget (KES)</div>
              <input type="number" min={0} step={100} value={dailyBudget} onChange={(e) => setDailyBudget(+e.target.value)}
                className="num h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" />
            </label>
            <label className="block md:col-span-2">
              <div className="mb-1.5 text-xs font-medium">Total budget (KES)</div>
              <input type="number" min={0} step={500} value={totalBudget} onChange={(e) => setTotalBudget(+e.target.value)}
                className="num h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" />
            </label>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <SectionLabel>Targeting</SectionLabel>
          <div className="space-y-4">
            <div>
              <div className="mb-2 text-xs font-medium">Countries</div>
              <div className="flex flex-wrap gap-2">
                {COUNTRIES.map((c) => (
                  <Chip key={c} selected={countries.includes(c)} onClick={() => toggle(countries, c, setCountries)}>{c}</Chip>
                ))}
              </div>
            </div>
            <div>
              <div className="mb-2 text-xs font-medium">Devices</div>
              <div className="flex flex-wrap gap-2">
                {DEVICES.map((d) => (
                  <Chip key={d} selected={devices.includes(d)} onClick={() => toggle(devices, d, setDevices)}>{d}</Chip>
                ))}
              </div>
            </div>
            <div>
              <div className="mb-2 text-xs font-medium">Operating systems</div>
              <div className="flex flex-wrap gap-2">
                {OSES.map((o) => (
                  <Chip key={o} selected={os.includes(o)} onClick={() => toggle(os, o, setOs)}>{o}</Chip>
                ))}
              </div>
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <SectionLabel>Creative</SectionLabel>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Attach a creative (optional)</div>
            <select value={creativeId} onChange={(e) => setCreativeId(e.target.value)}
              className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm">
              <option value="">— none —</option>
              {creatives.filter((c) => c.format === format).map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
            <div className="mt-2 text-xs text-muted-foreground">
              Only creatives matching the selected format are shown. <Link to="/creatives" className="text-primary hover:underline">Manage creatives →</Link>
            </div>
          </label>
        </SurfaceCard>

        <div className="flex justify-end gap-3">
          <Link to="/campaigns" className="rounded-full border border-border bg-card px-5 py-2.5 text-sm font-medium hover:bg-muted">Cancel</Link>
          <button type="submit" className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground hover:brightness-110">
            Launch campaign
          </button>
        </div>
      </form>
    </div>
  );
}
