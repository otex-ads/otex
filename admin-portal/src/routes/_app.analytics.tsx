import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Download, Users, TrendingUp, DollarSign, Activity } from "lucide-react";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { api, type DetailedAnalyticsResponse } from "@/lib/api";

export const Route = createFileRoute("/_app/analytics")({
  component: AnalyticsPage,
});

const RANGES = [
  { key: "7d", label: "Last 7 days", days: 7 },
  { key: "30d", label: "Last 30 days", days: 30 },
  { key: "90d", label: "Last 90 days", days: 90 },
] as const;

const BREAKDOWNS = ["Country", "Device", "Site", "Zone", "Campaign"] as const;
type Breakdown = (typeof BREAKDOWNS)[number];

function isoDaysAgo(d: number) {
  return new Date(Date.now() - d * 86400_000).toISOString().slice(0, 10);
}

function KES(n: number) {
  return `KES ${n.toLocaleString(undefined, { maximumFractionDigits: 0 })}`;
}

function Num(n: number) {
  return n.toLocaleString();
}

function Pct(n: number) {
  return `${n.toFixed(2)}%`;
}

function AnalyticsPage() {
  const [range, setRange] = useState<(typeof RANGES)[number]["key"]>("30d");
  const [tab, setTab] = useState<Breakdown>("Country");

  const days = RANGES.find((r) => r.key === range)!.days;
  const from = isoDaysAgo(days);
  const to = isoDaysAgo(0);

  const platform = useQuery({ queryKey: ["admin-stats"], queryFn: () => api.getStats() });
  const analytics = useQuery<DetailedAnalyticsResponse>({
    queryKey: ["admin-analytics", range],
    queryFn: () => api.getDetailedAnalytics({ from, to }),
    retry: false,
  });

  const summary = analytics.data?.summary ?? {
    impressions: 0, clicks: 0, conversions: 0, ctr: 0, ecpm: 0, revenue: 0, grossSpend: 0, fillRate: 0,
  };
  const daily = analytics.data?.daily ?? [];

  const breakdownRows = useMemo(() => {
    const data = analytics.data;
    if (!data) return [];
    const map: Record<Breakdown, typeof data.byCountry> = {
      Country: data.byCountry,
      Device: data.byDevice,
      Site: data.bySite,
      Zone: data.byZone,
      Campaign: data.byCampaign,
    };
    return map[tab] ?? [];
  }, [analytics.data, tab]);

  const exportCsv = () => {
    const rows = [
      ["Key", "Impressions", "Clicks", "Revenue"],
      ...breakdownRows.map((r) => [r.key, r.impressions, r.clicks, r.revenue]),
    ];
    const csv = rows.map((r) => r.join(",")).join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `analytics-${tab.toLowerCase()}-${range}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Analytics"
        description="Platform-wide performance, revenue and traffic breakdowns."
        actions={
          <div className="flex flex-wrap gap-1 rounded-full border border-border bg-card p-1">
            {RANGES.map((r) => (
              <button
                key={r.key}
                onClick={() => setRange(r.key)}
                className={`rounded-full px-3 py-1 text-xs font-medium ${
                  range === r.key
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
        }
      />

      {/* Platform-wide counts (all-time, independent of range) */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <SurfaceCard className="p-6">
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-muted-foreground">Total Users</div>
            <Users className="size-4 text-muted-foreground" />
          </div>
          <div className="num mt-2 text-3xl font-bold text-foreground">
            {platform.data?.total_users ?? 0}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {platform.data?.total_advertisers ?? 0} advertisers · {platform.data?.total_publishers ?? 0} publishers
          </div>
        </SurfaceCard>
        <SurfaceCard className="p-6" delay={0.03}>
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-muted-foreground">Active Campaigns</div>
            <TrendingUp className="size-4 text-muted-foreground" />
          </div>
          <div className="num mt-2 text-3xl font-bold text-foreground">
            {platform.data?.active_campaigns ?? 0}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            of {platform.data?.total_campaigns ?? 0} total campaigns
          </div>
        </SurfaceCard>
        <SurfaceCard className="p-6" delay={0.06}>
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-muted-foreground">Lifetime Revenue</div>
            <DollarSign className="size-4 text-muted-foreground" />
          </div>
          <div className="num mt-2 text-3xl font-bold text-foreground">
            {KES((platform.data?.total_revenue_cents ?? 0) / 100)}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">Platform fee revenue, all-time</div>
        </SurfaceCard>
        <SurfaceCard className="p-6" delay={0.09}>
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-muted-foreground">Platform Status</div>
            <Activity className="size-4 text-green-600" />
          </div>
          <div className="mt-2 text-3xl font-bold text-green-600">Active</div>
          <div className="mt-1 text-xs text-muted-foreground">All systems operational</div>
        </SurfaceCard>
      </div>

      {/* Range-scoped performance summary */}
      <div className="grid gap-4 lg:grid-cols-6">
        {[
          { label: "Impressions", value: Num(summary.impressions) },
          { label: "Clicks", value: Num(summary.clicks) },
          { label: "Conversions", value: Num(summary.conversions) },
          { label: "CTR", value: Pct(summary.ctr) },
          { label: "eCPM", value: KES(summary.ecpm) },
          { label: "Platform Revenue", value: KES(summary.revenue) },
        ].map((s, i) => (
          <SurfaceCard key={s.label} delay={0.03 * i}>
            <div className="label-eyebrow">{s.label}</div>
            <div className="num display mt-2 text-2xl font-bold tracking-tight text-foreground">
              {s.value}
            </div>
          </SurfaceCard>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <SurfaceCard>
          <SectionLabel>Platform revenue over time</SectionLabel>
          <ChartLine data={daily} dataKey="revenue" formatValue={KES} />
        </SurfaceCard>
        <SurfaceCard>
          <SectionLabel>Impressions over time</SectionLabel>
          <ChartLine data={daily} dataKey="impressions" formatValue={Num} />
        </SurfaceCard>
      </div>

      <SurfaceCard className="p-0">
        <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-5">
          <div className="flex flex-wrap gap-1 rounded-full border border-border bg-muted/40 p-1">
            {BREAKDOWNS.map((b) => (
              <button
                key={b}
                onClick={() => setTab(b)}
                className={`rounded-full px-3 py-1 text-xs font-medium ${
                  tab === b ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                By {b}
              </button>
            ))}
          </div>
          <button
            onClick={exportCsv}
            className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted/50"
          >
            <Download className="size-3.5" /> Export CSV
          </button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">{tab}</th>
                <th className="px-6 py-3 text-right font-medium">Impressions</th>
                <th className="px-6 py-3 text-right font-medium">Clicks</th>
                <th className="px-6 py-3 text-right font-medium">
                  {tab === "Campaign" ? "Ad Spend" : "Revenue"}
                </th>
              </tr>
            </thead>
            <tbody>
              {analytics.isLoading && (
                <tr>
                  <td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">
                    Loading…
                  </td>
                </tr>
              )}
              {!analytics.isLoading && breakdownRows.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">
                    No data for the selected range.
                  </td>
                </tr>
              )}
              {breakdownRows.map((r) => (
                <tr key={r.key} className="border-b border-border last:border-0">
                  <td className="px-6 py-4 text-foreground">{r.key}</td>
                  <td className="num px-6 py-4 text-right">{Num(r.impressions)}</td>
                  <td className="num px-6 py-4 text-right">{Num(r.clicks)}</td>
                  <td className="num px-6 py-4 text-right">{KES(r.revenue)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>
    </div>
  );
}

function ChartLine({
  data,
  dataKey,
  formatValue,
}: {
  data: Array<{ date: string; revenue: number; impressions: number; clicks: number }>;
  dataKey: "revenue" | "impressions" | "clicks";
  formatValue: (n: number) => string;
}) {
  return (
    <div className="mt-4 h-64">
      {data.length === 0 ? (
        <div className="grid h-full place-items-center text-sm text-muted-foreground">No data</div>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} width={60} />
            <Tooltip
              formatter={(v: number) => formatValue(v)}
              contentStyle={{ borderRadius: 12, border: "1px solid oklch(0 0 0 / 0.06)", fontSize: 12 }}
            />
            <Line type="monotone" dataKey={dataKey} stroke="oklch(0.305 0.052 160)" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
