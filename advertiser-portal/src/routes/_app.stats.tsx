import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
} from "recharts";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { PageHeader } from "@/components/PageHeader";
import { useStore } from "@/lib/store";
import { KES, Num } from "@/lib/format";

export const Route = createFileRoute("/_app/stats")({
  component: StatsPage,
});

const RANGES = [
  { label: "7 days", days: 7 },
  { label: "14 days", days: 14 },
  { label: "30 days", days: 30 },
];

function StatsPage() {
  const stats = useStore((s) => s.stats) || [];
  const campaigns = useStore((s) => s.campaigns) || [];
  const [days, setDays] = useState(14);
  const [campaignId, setCampaignId] = useState<string>("all");

  const cutoff = useMemo(() => {
    const d = new Date();
    d.setDate(d.getDate() - days + 1);
    return d.toISOString().slice(0, 10);
  }, [days]);

  const filtered = stats.filter(
    (s) => s.date >= cutoff && (campaignId === "all" || s.campaignId === campaignId),
  );

  const totals = filtered.reduce(
    (a, s) => ({
      impressions: a.impressions + s.impressions,
      clicks: a.clicks + s.clicks,
      conversions: a.conversions + s.conversions,
      spend: a.spend + s.spend,
    }),
    { impressions: 0, clicks: 0, conversions: 0, spend: 0 },
  );
  const ctr = totals.impressions ? (totals.clicks / totals.impressions) * 100 : 0;
  const cpc = totals.clicks ? totals.spend / totals.clicks : 0;

  const byDay = new Map<
    string,
    { date: string; impressions: number; clicks: number; spend: number }
  >();
  for (const s of filtered) {
    const cur = byDay.get(s.date) ?? { date: s.date, impressions: 0, clicks: 0, spend: 0 };
    cur.impressions += s.impressions;
    cur.clicks += s.clicks;
    cur.spend += s.spend;
    byDay.set(s.date, cur);
  }
  const daily = [...byDay.values()].sort((a, b) => a.date.localeCompare(b.date));

  const byCampaign = campaigns.map((c) => {
    const rows = filtered.filter((s) => s.campaignId === c.id);
    const t = rows.reduce(
      (a, s) => ({
        impressions: a.impressions + s.impressions,
        clicks: a.clicks + s.clicks,
        conversions: a.conversions + s.conversions,
        spend: a.spend + s.spend,
      }),
      { impressions: 0, clicks: 0, conversions: 0, spend: 0 },
    );
    return { campaign: c, ...t, ctr: t.impressions ? (t.clicks / t.impressions) * 100 : 0 };
  });

  return (
    <div>
      <PageHeader
        title="Statistics"
        description="Impressions, clicks, spend & CTR — filter by campaign and date range."
        actions={
          <div className="flex items-center gap-2">
            <select
              value={campaignId}
              onChange={(e) => setCampaignId(e.target.value)}
              className="h-10 rounded-full border border-border bg-card px-4 text-sm"
            >
              <option value="all">All campaigns</option>
              {campaigns.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
            <div className="flex overflow-hidden rounded-full border border-border bg-card">
              {RANGES.map((r) => (
                <button
                  key={r.days}
                  onClick={() => setDays(r.days)}
                  className={`px-4 py-2 text-xs font-medium transition-colors ${
                    days === r.days
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {r.label}
                </button>
              ))}
            </div>
          </div>
        }
      />

      <div className="mb-6 grid gap-4 md:grid-cols-4">
        {[
          { label: "Impressions", value: Num(totals.impressions) },
          { label: "Clicks", value: Num(totals.clicks), sub: `CTR ${ctr.toFixed(2)}%` },
          { label: "Conversions", value: Num(totals.conversions) },
          { label: "Spend", value: KES(totals.spend), sub: `Avg CPC ${KES(cpc, { decimals: 2 })}` },
        ].map((s, i) => (
          <SurfaceCard key={s.label} delay={0.04 * i}>
            <div className="label-eyebrow">{s.label}</div>
            <div className="num mt-3 text-3xl font-bold">{s.value}</div>
            {s.sub && <div className="mt-1 text-xs text-muted-foreground">{s.sub}</div>}
          </SurfaceCard>
        ))}
      </div>

      <SurfaceCard>
        <SectionLabel>Daily performance</SectionLabel>
        <div className="mt-3 h-72">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={daily} barGap={6}>
              <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }}
                axisLine={false}
                tickLine={false}
                tickFormatter={(d) =>
                  new Date(d).toLocaleDateString("en-GB", { day: "2-digit", month: "short" })
                }
              />
              <YAxis
                yAxisId="l"
                tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }}
                axisLine={false}
                tickLine={false}
                tickFormatter={(v) => `${v / 1000}K`}
                width={40}
              />
              <YAxis
                yAxisId="r"
                orientation="right"
                tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }}
                axisLine={false}
                tickLine={false}
                tickFormatter={(v) => `${v / 1000}K`}
                width={40}
              />
              <Tooltip
                contentStyle={{
                  borderRadius: 12,
                  border: "1px solid oklch(0 0 0 / 0.06)",
                  fontSize: 12,
                }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Bar
                yAxisId="l"
                dataKey="impressions"
                fill="oklch(0.305 0.052 160)"
                radius={[6, 6, 0, 0]}
              />
              <Bar yAxisId="l" dataKey="clicks" fill="oklch(0.62 0.06 160)" radius={[6, 6, 0, 0]} />
              <Bar yAxisId="r" dataKey="spend" fill="oklch(0.82 0.03 150)" radius={[6, 6, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </SurfaceCard>

      <SurfaceCard className="mt-6 p-0">
        <div className="px-6 py-5">
          <SectionLabel>By campaign</SectionLabel>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Campaign</th>
                <th className="px-6 py-3 font-medium">Format</th>
                <th className="px-6 py-3 text-right font-medium">Impressions</th>
                <th className="px-6 py-3 text-right font-medium">Clicks</th>
                <th className="px-6 py-3 text-right font-medium">CTR</th>
                <th className="px-6 py-3 text-right font-medium">Conv.</th>
                <th className="px-6 py-3 text-right font-medium">Spend</th>
              </tr>
            </thead>
            <tbody>
              {byCampaign.map((r) => (
                <tr key={r.campaign.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-4 text-foreground">{r.campaign.name}</td>
                  <td className="px-6 py-4 capitalize text-muted-foreground">
                    {r.campaign.format}
                  </td>
                  <td className="num px-6 py-4 text-right">{Num(r.impressions)}</td>
                  <td className="num px-6 py-4 text-right">{Num(r.clicks)}</td>
                  <td className="num px-6 py-4 text-right">{r.ctr.toFixed(2)}%</td>
                  <td className="num px-6 py-4 text-right">{Num(r.conversions)}</td>
                  <td className="num px-6 py-4 text-right font-medium">{KES(r.spend)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>
    </div>
  );
}
