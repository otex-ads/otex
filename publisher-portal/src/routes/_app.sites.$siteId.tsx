import { createFileRoute, Link, useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Globe, Plus } from "lucide-react";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { api } from "@/lib/api";
import { USD, Num, Pct } from "@/lib/format";

export const Route = createFileRoute("/_app/sites/$siteId")({
  component: SiteDetail,
});

function isoDaysAgo(d: number) { return new Date(Date.now() - d * 86400_000).toISOString().slice(0, 10); }

function SiteDetail() {
  const { siteId } = useParams({ from: "/_app/sites/$siteId" });
  const sites = useQuery({ queryKey: ["sites"], queryFn: api.listSites, retry: false });
  const zones = useQuery({ queryKey: ["zones", siteId], queryFn: () => api.listZones(siteId), retry: false });
  const stats = useQuery({
    queryKey: ["stats", "site", siteId],
    queryFn: () => api.stats({ from: isoDaysAgo(30), to: isoDaysAgo(0), groupBy: `site:${siteId}` }),
    retry: false,
  });
  const site = sites.data?.find((s) => s.id === siteId);

  return (
    <div className="space-y-6">
      <Link to="/sites" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="size-4" /> Back to sites
      </Link>

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="label-eyebrow mb-2">Site details</div>
          <h1 className="display text-4xl font-normal tracking-tight text-foreground">{site?.name ?? "Site"}</h1>
          <div className="mt-2 flex items-center gap-3 text-sm text-muted-foreground">
            <Globe className="size-4" /> {site?.domain}
            {site && <StatusPill status={site.status[0].toUpperCase() + site.status.slice(1)} tone={site.status === "active" ? "success" : "warning"} />}
          </div>
        </div>
        <Link to="/zones" className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground">
          <Plus className="size-4" /> Add zone
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {[
          { label: "Impressions", value: Num(stats.data?.summary.impressions ?? site?.impressions ?? 0) },
          { label: "Clicks", value: Num(stats.data?.summary.clicks ?? 0) },
          { label: "CTR", value: Pct(stats.data?.summary.ctr ?? 0) },
          { label: "Revenue", value: USD(stats.data?.summary.revenue ?? site?.revenue ?? 0) },
        ].map((s, i) => (
          <SurfaceCard key={s.label} delay={0.04 * i}>
            <div className="label-eyebrow">{s.label}</div>
            <div className="num display mt-2 text-3xl font-bold tracking-tight text-foreground">{s.value}</div>
          </SurfaceCard>
        ))}
      </div>

      <SurfaceCard>
        <SectionLabel>Performance · 30 days</SectionLabel>
        <div className="mt-4 h-64">
          {(stats.data?.daily?.length ?? 0) === 0 ? (
            <div className="grid h-full place-items-center text-sm text-muted-foreground">No data yet</div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={stats.data!.daily}>
                <CartesianGrid strokeDasharray="2 4" stroke="oklch(0 0 0 / 0.06)" vertical={false} />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 11, fill: "oklch(0.48 0.012 150)" }} axisLine={false} tickLine={false} width={40} />
                <Tooltip contentStyle={{ borderRadius: 12, border: "1px solid oklch(0 0 0 / 0.06)", fontSize: 12 }} />
                <Line type="monotone" dataKey="revenue" stroke="oklch(0.305 0.052 160)" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>
      </SurfaceCard>

      <SurfaceCard className="p-0">
        <div className="px-6 py-5">
          <SectionLabel>Zones on this site</SectionLabel>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Zone</th>
                <th className="px-6 py-3 font-medium">Size</th>
                <th className="px-6 py-3 font-medium">Format</th>
                <th className="px-6 py-3 text-right font-medium">Impr.</th>
                <th className="px-6 py-3 text-right font-medium">Revenue</th>
              </tr>
            </thead>
            <tbody>
              {(zones.data ?? []).length === 0 && (
                <tr><td colSpan={5} className="px-6 py-10 text-center text-muted-foreground">No zones on this site</td></tr>
              )}
              {(zones.data ?? []).map((z) => (
                <tr key={z.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-4"><Link to="/zones/$zoneId" params={{ zoneId: z.id }} className="font-medium text-foreground hover:text-primary">{z.name}</Link></td>
                  <td className="px-6 py-4 text-muted-foreground">{z.size}</td>
                  <td className="px-6 py-4 capitalize text-muted-foreground">{z.format}</td>
                  <td className="num px-6 py-4 text-right">{Num(z.impressions)}</td>
                  <td className="num px-6 py-4 text-right">{USD(z.revenue)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>
    </div>
  );
}
