import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Download } from "lucide-react";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { api, type StatsResponse } from "@/lib/api";
import { USD, Num, Pct } from "@/lib/format";

export const Route = createFileRoute("/_app/stats")({
  component: StatsPage,
});

const RANGES = [
  { key: "today", label: "Today", days: 0 },
  { key: "yesterday", label: "Yesterday", days: 1 },
  { key: "7d", label: "Last 7 days", days: 7 },
  { key: "30d", label: "Last 30 days", days: 30 },
  { key: "90d", label: "Last 90 days", days: 90 },
] as const;

const BREAKDOWNS = ["Site", "Zone", "Country", "Device", "Browser", "OS"] as const;
type Breakdown = typeof BREAKDOWNS[number];

function isoDaysAgo(d: number) { return new Date(Date.now() - d * 86400_000).toISOString().slice(0, 10); }

function StatsPage() {
  const [range, setRange] = useState<typeof RANGES[number]["key"]>("30d");
  const [tab, setTab] = useState<Breakdown>("Site");

  const days = RANGES.find((r) => r.key === range)!.days;
  const from = isoDaysAgo(Math.max(days, 1));
  const to = isoDaysAgo(0);

  const stats = useQuery<StatsResponse>({
    queryKey: ["stats", "analytics", range],
    queryFn: () => api.stats({ from, to }),
    retry: false,
  });

  const summary = stats.data?.summary ?? { impressions: 0, clicks: 0, ctr: 0, ecpm: 0, revenue: 0, fillRate: 0 };
  const daily = stats.data?.daily ?? [];

  const breakdownRows = useMemo(() => {
    const data = stats.data;
    if (!data) return [];
    const map: Record<Breakdown, StatsResponse["byCountry"]> = {
      Site: data.bySite, Zone: data.byZone, Country: data.byCountry, Device: data.byDevice, Browser: data.byBrowser, OS: data.byOS,
    };
    return map[tab] ?? [];
  }, [stats.data, tab]);

  const exportCsv = () => {
    const rows = [["Key", "Impressions", "Clicks", "Revenue"], ...breakdownRows.map((r) => [r.key, r.impressions, r.clicks, r.revenue])];
    const csv = rows.map((r) => r.join(",")).join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url; a.download = `stats-${tab.toLowerCase()}-${range}.csv`; a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Statistics"
        description="Analyze earnings and traffic across every dimension."
        actions={
          <div className="flex flex-wrap gap-1 rounded-full border border-border bg-card p-1">
            {RANGES.map((r) => (
              <button key={r.key} onClick={() => setRange(r.key)} className={`rounded-full px-3 py-1 text-xs font-medium ${range === r.key ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground"}`}>
                {r.label}
              </button>
            ))}
          </div>
        }
      />

      <div className="grid gap-4 lg:grid-cols-6">
        {[
          { label: "Impressions", value: Num(summary.impressions) },
          { label: "Clicks", value: Num(summary.clicks) },
          { label: "CTR", value: Pct(summary.ctr) },
          { label: "eCPM", value: USD(summary.ecpm) },
          { label: "Revenue", value: USD(summary.revenue) },
          { label: "Fill rate", value: Pct(summary.fillRate) },
        ].map((s, i) => (
          <SurfaceCard key={s.label} delay={0.03 * i}>
            <div className="label-eyebrow">{s.label}</div>
            <div className="num display mt-2 text-2xl font-bold tracking-tight text-foreground">{s.value}</div>
          </SurfaceCard>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <SurfaceCard>
          <SectionLabel>Revenue over time</SectionLabel>
          <ChartLine data={daily} dataKey="revenue" />
        </SurfaceCard>
        <SurfaceCard>
          <SectionLabel>Impressions over time</SectionLabel>
          <ChartLine data={daily} dataKey="impressions" />
        </SurfaceCard>
      </div>

      <SurfaceCard className="p-0">
        <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-5">
          <div className="flex flex-wrap gap-1 rounded-full border border-border bg-muted/40 p-1">
            {BREAKDOWNS.map((b) => (
              <button key={b} onClick={() => setTab(b)} className={`rounded-full px-3 py-1 text-xs font-medium ${tab === b ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}>
                By {b}
              </button>
            ))}
          </div>
          <button onClick={exportCsv} className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted/50">
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
                <th className="px-6 py-3 text-right font-medium">Revenue</th>
              </tr>
            </thead>
            <tbody>
              {(breakdownRows ?? []).length === 0 && (
                <tr><td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">No data for the selected range.</td></tr>
              )}
              {(breakdownRows ?? []).map((r) => (
                <tr key={r.key} className="border-b border-border last:border-0">
                  <td className="px-6 py-4 text-foreground">{r.key}</td>
                  <td className="num px-6 py-4 text-right">{Num(r.impressions)}</td>
                  <td className="num px-6 py-4 text-right">{Num(r.clicks)}</td>
                  <td className="num px-6 py-4 text-right">{USD(r.revenue)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>
    </div>
  );
}

function ChartLine({ data, dataKey }: { data: Array<{ date: string; revenue: number; impressions: number; clicks: number }>; dataKey: "revenue" | "impressions" | "clicks" }) {
  return (
    <div className="mt-4 h-64">
      {data.length === 0 ? (
        <div className="grid h-full place-items-center text-sm text-muted-foreground">No data</div>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} width={50} />
            <Tooltip contentStyle={{ borderRadius: 12, border: "1px solid oklch(0 0 0 / 0.06)", fontSize: 12 }} />
            <Line type="monotone" dataKey={dataKey} stroke="oklch(0.305 0.052 160)" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
