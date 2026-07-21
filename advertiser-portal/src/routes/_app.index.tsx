import { createFileRoute, Link } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { MousePointerClick, Eye, TrendingUp, DollarSign, Plus, ArrowUpRight } from "lucide-react";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { useStore } from "@/lib/store";
import { KES, Num } from "@/lib/format";

export const Route = createFileRoute("/_app/")({
  component: Dashboard,
});

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

function Dashboard() {
  const campaigns = useStore((s) => s.campaigns) || [];
  const stats = useStore((s) => s.stats) || [];

  const totals = stats.reduce(
    (acc, s) => {
      acc.impressions += s.impressions;
      acc.clicks += s.clicks;
      acc.conversions += s.conversions;
      acc.spend += s.spend;
      return acc;
    },
    { impressions: 0, clicks: 0, conversions: 0, spend: 0 },
  );
  const ctr = totals.impressions ? (totals.clicks / totals.impressions) * 100 : 0;

  // Trend by day
  const byDay = new Map<string, { date: string; impressions: number; clicks: number; spend: number }>();
  for (const s of stats) {
    const cur = byDay.get(s.date) ?? { date: s.date, impressions: 0, clicks: 0, spend: 0 };
    cur.impressions += s.impressions;
    cur.clicks += s.clicks;
    cur.spend += s.spend;
    byDay.set(s.date, cur);
  }
  const trend = [...byDay.values()].sort((a, b) => a.date.localeCompare(b.date));

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
              <span>Overview</span><span className="text-anchor-foreground/30">/</span><span className="text-anchor-foreground/90">Last 14 days</span>
            </div>
            <h1 className="display mt-4 text-5xl font-bold tracking-tight md:text-6xl">
              Your ads, <span className="text-anchor-foreground/60">performing.</span>
            </h1>
            <p className="mt-3 max-w-xl text-sm text-anchor-foreground/60">
              Live snapshot across every campaign — impressions, clicks, and spend rolled up in real time.
            </p>
          </div>
          <Link
            to="/campaigns/new"
            className="inline-flex items-center gap-2 rounded-full bg-white/95 px-5 py-2.5 text-sm font-medium text-anchor hover:bg-white"
          >
            <Plus className="size-4" /> New campaign
          </Link>
        </div>
      </motion.section>

      <div className="grid gap-4 lg:grid-cols-4">
        <StatCard delay={0.04} label="Impressions" value={Num(totals.impressions)} sub="Last 14 days" icon={Eye} />
        <StatCard delay={0.08} label="Clicks" value={Num(totals.clicks)} sub={`CTR ${ctr.toFixed(2)}%`} icon={MousePointerClick} />
        <StatCard delay={0.12} label="Conversions" value={Num(totals.conversions)} sub="Attributed" icon={TrendingUp} />
        <StatCard delay={0.16} label="Spend" value={KES(totals.spend, { compact: true })} sub="Across all campaigns" icon={DollarSign} />
      </div>

      <SurfaceCard delay={0.2}>
        <div className="flex items-start justify-between">
          <SectionLabel>Traffic trend · 14 days</SectionLabel>
        </div>
        <div className="mt-3 h-72">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={trend}>
              <defs>
                <linearGradient id="imp" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="oklch(0.305 0.052 160)" stopOpacity={0.4} />
                  <stop offset="100%" stopColor="oklch(0.305 0.052 160)" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="clk" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="oklch(0.62 0.06 160)" stopOpacity={0.5} />
                  <stop offset="100%" stopColor="oklch(0.62 0.06 160)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
              <XAxis dataKey="date" tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false}
                tickFormatter={(d) => new Date(d).toLocaleDateString("en-GB", { day: "2-digit", month: "short" })} />
              <YAxis tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} tickFormatter={(v) => `${v / 1000}K`} width={40} />
              <Tooltip contentStyle={{ borderRadius: 12, border: "1px solid oklch(0 0 0 / 0.06)", fontSize: 12 }} />
              <Area type="monotone" dataKey="impressions" stroke="oklch(0.305 0.052 160)" fill="url(#imp)" strokeWidth={2} />
              <Area type="monotone" dataKey="clicks" stroke="oklch(0.62 0.06 160)" fill="url(#clk)" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </SurfaceCard>

      <SurfaceCard delay={0.24} className="p-0">
        <div className="flex items-center justify-between px-6 py-5">
          <div>
            <SectionLabel>Active campaigns</SectionLabel>
            <div className="text-base font-semibold">{campaigns.length} total · {campaigns.filter(c => c.status === "active").length} active</div>
          </div>
          <Link to="/campaigns" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
            View all <ArrowUpRight className="size-3" />
          </Link>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Campaign</th>
                <th className="px-6 py-3 font-medium">Format</th>
                <th className="px-6 py-3 font-medium">Model</th>
                <th className="px-6 py-3 text-right font-medium">Impr.</th>
                <th className="px-6 py-3 text-right font-medium">Clicks</th>
                <th className="px-6 py-3 text-right font-medium">Spend</th>
                <th className="px-6 py-3 font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {campaigns.slice(0, 5).map((c) => (
                <tr key={c.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-4 text-foreground">{c.name}</td>
                  <td className="px-6 py-4 capitalize text-muted-foreground">{c.format}</td>
                  <td className="px-6 py-4 text-muted-foreground">{c.pricingModel}</td>
                  <td className="num px-6 py-4 text-right">{Num(c.impressions)}</td>
                  <td className="num px-6 py-4 text-right">{Num(c.clicks)}</td>
                  <td className="num px-6 py-4 text-right">{KES(c.spent)}</td>
                  <td className="px-6 py-4">
                    <StatusPill
                      status={c.status[0].toUpperCase() + c.status.slice(1)}
                      tone={c.status === "active" ? "success" : c.status === "paused" ? "warning" : "neutral"}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>
    </div>
  );
}
