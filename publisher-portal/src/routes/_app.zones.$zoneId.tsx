import { createFileRoute, Link, useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Copy } from "lucide-react";
import { toast } from "sonner";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { api } from "@/lib/api";
import { USD, Num, Pct } from "@/lib/format";
import { buildAdCode } from "@/lib/adcode";

export const Route = createFileRoute("/_app/zones/$zoneId")({
  component: ZoneDetail,
});

function ZoneDetail() {
  const { zoneId } = useParams({ from: "/_app/zones/$zoneId" });
  const zones = useQuery({ queryKey: ["zones"], queryFn: () => api.listZones(), retry: false });
  const zone = zones.data?.find((z) => z.id === zoneId);
  const code = zone ? buildAdCode({ zoneId: zone.id, async: true, responsive: true }) : "";

  const copy = async () => {
    try { await navigator.clipboard.writeText(code); toast.success("Ad code copied"); } catch { toast.error("Copy failed"); }
  };

  return (
    <div className="space-y-6">
      <Link to="/zones" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="size-4" /> Back to zones
      </Link>

      <div>
        <div className="label-eyebrow mb-2">Zone details</div>
        <h1 className="display text-4xl font-normal tracking-tight text-foreground">{zone?.name ?? "Zone"}</h1>
        {zone && (
          <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
            <span>{zone.siteName ?? zone.siteId}</span><span>·</span>
            <span>{zone.size}</span><span>·</span>
            <span className="capitalize">{zone.format}</span>
            <StatusPill status={zone.status[0].toUpperCase() + zone.status.slice(1)} tone={zone.status === "active" ? "success" : "warning"} />
          </div>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {[
          { label: "Impressions", value: Num(zone?.impressions ?? 0) },
          { label: "Clicks", value: Num(zone?.clicks ?? 0) },
          { label: "CTR", value: Pct(zone && zone.impressions ? (zone.clicks / zone.impressions) * 100 : 0) },
          { label: "Revenue", value: USD(zone?.revenue ?? 0) },
        ].map((s, i) => (
          <SurfaceCard key={s.label} delay={0.04 * i}>
            <div className="label-eyebrow">{s.label}</div>
            <div className="num display mt-2 text-3xl font-bold tracking-tight text-foreground">{s.value}</div>
          </SurfaceCard>
        ))}
      </div>

      <SurfaceCard>
        <div className="flex items-start justify-between gap-3">
          <div>
            <SectionLabel>Embed code</SectionLabel>
            <p className="text-sm text-muted-foreground">Paste this snippet where you want the ad to render.</p>
          </div>
          <button onClick={copy} className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted/50">
            <Copy className="size-3.5" /> Copy
          </button>
        </div>
        <pre className="font-mono mt-4 max-h-64 overflow-auto rounded-lg border border-border bg-muted/40 p-4 text-[11px] leading-relaxed text-foreground">{code}</pre>
      </SurfaceCard>
    </div>
  );
}
