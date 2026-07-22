import { createFileRoute, Link } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { MousePointerClick, Eye, TrendingUp, DollarSign, Plus, ArrowUpRight, Code2, Wallet as WalletIcon, Percent, Gauge } from "lucide-react";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { api, type StatsResponse, type Balance, type Site, type Zone } from "@/lib/api";
import { USD, Num, Pct } from "@/lib/format";

export const Route = createFileRoute("/_app/")({
  component: Dashboard,
});

const RANGES = [
  { label: "7 days", days: 7 },
  { label: "30 days", days: 30 },
  { label: "90 days", days: 90 },
] as const;

function StatCard({ label, value, sub, icon: Icon, delay = 0 }: {
  label: string; value: string; sub?: string; icon: typeof Eye; delay?: number;
}) {
  return (
    <SurfaceCard delay={delay} className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="label-eyebrow">{label}</div>
        <div className="grid size-8 place-items-center rounded-lg bg-muted text-muted-foreground">
          <Icon className="size-4" />
        </div>
      </div>
      <div>
        <div className="num display text-4xl font-bold tracking-tight text-foreground">{value}</div>
        {sub && <div className="mt-2 text-xs text-muted-foreground">{sub}</div>}
      </div>
    </SurfaceCard>
  );
}

function isoDaysAgo(d: number) {
  return new Date(Date.now() - d * 86400_000).toISOString().slice(0, 10);
}

function Dashboard() {
  const [days, setDays] = useState<number>(30);
  const from = isoDaysAgo(days);
  const to = isoDaysAgo(0);

  const stats = useQuery<StatsResponse>({
    queryKey: ["stats", "dashboard", days],
    queryFn: () => api.stats({ from, to }),
    retry: false,
  });
  const balance = useQuery<Balance>({ queryKey: ["balance"], queryFn: api.balance, retry: false });
  const sites = useQuery<Site[]>({ queryKey: ["sites"], queryFn: api.listSites, retry: false });
  const zones = useQuery<Zone[]>({ queryKey: ["zones"], queryFn: () => api.listZones(), retry: false });

  const summary = stats.data?.summary ?? { impressions: 0, clicks: 0, ctr: 0, ecpm: 0, revenue: 0, fillRate: 0 };
  const daily = stats.data?.daily ?? [];

  const today = daily[daily.length - 1]?.revenue ?? 0;
  const thisMonth = daily
    .filter((d) => d.date.startsWith(to.slice(0, 7)))
    .reduce((sum, d) => sum + d.revenue, 0);

  const topSites = [...(sites.data ?? [])]
    .sort((a, b) => b.revenue - a.revenue)
    .slice(0, 5);
  const topZones = [...(zones.data ?? [])]
    .sort((a, b) => b.revenue - a.revenue)
    .slice(0, 5);

  return (
    <div className="space-y-8">
      <motion.section
        initial={{ opacity: 0, y: -6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
        className="card-anchor relative overflow-hidden p-8 md:p-10"
      >
        <div className="absolute -right-24 -top-24 size-72 rounded-full bg-white/[0.04] blur-3xl" />
        <div className="absolute -bottom-24 left-1/3 size-80 rounded-full bg-white/[0.03] blur-3xl" />
        <div className="relative flex flex-wrap items-end justify-between gap-6">
          <div>
            <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-anchor-foreground/60">
              <span>Overview</span><span className="text-anchor-foreground/30">/</span><span className="text-anchor-foreground/90">Last {days} days</span>
            </div>
            <h1 className="display mt-4 text-5xl font-bold tracking-tight md:text-6xl">
              Your traffic, <span className="text-anchor-foreground/60">earning.</span>
            </h1>
            <p className="mt-3 max-w-xl text-sm text-anchor-foreground/60">
              Live earnings snapshot across every site and zone — updated in real time.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Link to="/sites" className="inline-flex items-center gap-2 rounded-full bg-white/95 px-4 py-2 text-sm font-medium text-anchor hover:bg-white">
              <Plus className="size-4" /> Add site
            </Link>
            <Link to="/adcode" className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-4 py-2 text-sm font-medium text-anchor-foreground hover:bg-white/15">
              <Code2 className="size-4" /> Get ad code
            </Link>
            <Link to="/payouts" className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-4 py-2 text-sm font-medium text-anchor-foreground hover:bg-white/15">
              <WalletIcon className="size-4" /> Request payout
            </Link>
          </div>
        </div>
      </motion.section>

      <div className="grid gap-4 lg:grid-cols-3">
        <StatCard delay={0.04} label="Total earnings" value={USD(summary.revenue, { compact: true })} sub={`Last ${days} days`} icon={DollarSign} />
        <StatCard delay={0.08} label="Today" value={USD(today, { compact: true })} sub="Rolling last day" icon={TrendingUp} />
        <StatCard delay={0.12} label="This month" value={USD(thisMonth, { compact: true })} sub={`Available: ${USD(balance.data?.available ?? 0, { compact: true })}`} icon={WalletIcon} />
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        <StatCard delay={0.16} label="Impressions" value={Num(summary.impressions)} icon={Eye} />
        <StatCard delay={0.18} label="Clicks" value={Num(summary.clicks)} icon={MousePointerClick} />
        <StatCard delay={0.2} label="CTR" value={Pct(summary.ctr)} icon={Percent} />
        <StatCard delay={0.22} label="eCPM" value={USD(summary.ecpm)} icon={TrendingUp} />
        <StatCard delay={0.24} label="Fill rate" value={Pct(summary.fillRate)} icon={Gauge} />
      </div>

      <SurfaceCard delay={0.28}>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <SectionLabel>Revenue trend · {days} days</SectionLabel>
          <div className="flex gap-1 rounded-full border border-border bg-muted/40 p-1">
            {RANGES.map((r) => (
              <button
                key={r.days}
                onClick={() => setDays(r.days)}
                className={`rounded-full px-3 py-1 text-xs font-medium transition ${
                  days === r.days ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
        </div>
        <div className="mt-4 h-72">
          {daily.length === 0 ? (
            <div className="grid h-full place-items-center text-sm text-muted-foreground">
              {stats.isLoading ? "Loading…" : "No data yet — add a site and start earning."}
            </div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={daily}>
                <defs>
                  <linearGradient id="rev" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="oklch(0.305 0.052 160)" stopOpacity={0.4} />
                    <stop offset="100%" stopColor="oklch(0.305 0.052 160)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false}
                  tickFormatter={(d) => new Date(d).toLocaleDateString("en-GB", { day: "2-digit", month: "short" })} />
                <YAxis tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${v}`} width={50} />
                <Tooltip contentStyle={{ borderRadius: 12, border: "1px solid oklch(0 0 0 / 0.06)", fontSize: 12 }} />
                <Area type="monotone" dataKey="revenue" stroke="oklch(0.305 0.052 160)" fill="url(#rev)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>
      </SurfaceCard>

      <div className="grid gap-6 lg:grid-cols-2">
        <SurfaceCard delay={0.3} className="p-0">
          <div className="flex items-center justify-between px-6 py-5">
            <div>
              <SectionLabel>Top sites</SectionLabel>
              <div className="text-base font-semibold">By revenue</div>
            </div>
            <Link to="/sites" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              View all <ArrowUpRight className="size-3" />
            </Link>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Site</th>
                  <th className="px-6 py-3 text-right font-medium">Impr.</th>
                  <th className="px-6 py-3 text-right font-medium">Revenue</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {topSites.length === 0 && (
                  <tr><td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">No sites yet</td></tr>
                )}
                {topSites.map((s) => (
                  <tr key={s.id} className="border-b border-border last:border-0">
                    <td className="px-6 py-4">
                      <div className="font-medium text-foreground">{s.name}</div>
                      <div className="text-xs text-muted-foreground">{s.domain}</div>
                    </td>
                    <td className="num px-6 py-4 text-right">{Num(s.impressions)}</td>
                    <td className="num px-6 py-4 text-right">{USD(s.revenue)}</td>
                    <td className="px-6 py-4">
                      <StatusPill status={s.status[0].toUpperCase() + s.status.slice(1)} tone={s.status === "active" ? "success" : "warning"} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </SurfaceCard>

        <SurfaceCard delay={0.32} className="p-0">
          <div className="flex items-center justify-between px-6 py-5">
            <div>
              <SectionLabel>Top zones</SectionLabel>
              <div className="text-base font-semibold">Best performers</div>
            </div>
            <Link to="/zones" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              View all <ArrowUpRight className="size-3" />
            </Link>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Zone</th>
                  <th className="px-6 py-3 font-medium">Size</th>
                  <th className="px-6 py-3 text-right font-medium">Clicks</th>
                  <th className="px-6 py-3 text-right font-medium">Revenue</th>
                </tr>
              </thead>
              <tbody>
                {topZones.length === 0 && (
                  <tr><td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">No zones yet</td></tr>
                )}
                {topZones.map((z) => (
                  <tr key={z.id} className="border-b border-border last:border-0">
                    <td className="px-6 py-4">
                      <div className="font-medium text-foreground">{z.name}</div>
                      <div className="text-xs text-muted-foreground">{z.siteName ?? z.siteId}</div>
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">{z.size}</td>
                    <td className="num px-6 py-4 text-right">{Num(z.clicks)}</td>
                    <td className="num px-6 py-4 text-right">{USD(z.revenue)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </SurfaceCard>
      </div>
    </div>
  );
}
