import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Copy, Code2 } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { FormField, selectCls } from "@/components/Modal";
import { api } from "@/lib/api";
import { buildAdCode } from "@/lib/adcode";

export const Route = createFileRoute("/_app/adcode")({
  component: AdCodePage,
});

function AdCodePage() {
  const zones = useQuery({ queryKey: ["zones"], queryFn: () => api.listZones(), retry: false });
  const [zoneId, setZoneId] = useState<string>("");
  const [asyncLoad, setAsyncLoad] = useState(true);
  const [responsive, setResponsive] = useState(true);
  const [customStyle, setCustomStyle] = useState("");

  const list = zones.data ?? [];
  const selected = list.find((z) => z.id === zoneId) ?? list[0];
  const activeId = selected?.id ?? "";

  const code = useMemo(
    () => (activeId ? buildAdCode({ zoneId: activeId, async: asyncLoad, responsive, customStyle }) : "// Select a zone to generate ad code."),
    [activeId, asyncLoad, responsive, customStyle],
  );

  const [w, h] = (selected?.size ?? "300x250").split("x").map(Number);

  const copy = async () => {
    try { await navigator.clipboard.writeText(code); toast.success("Ad code copied"); } catch { toast.error("Copy failed"); }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Ad Code Generator"
        description="Generate the embed snippet for any of your zones."
      />

      <div className="grid gap-6 lg:grid-cols-2">
        <SurfaceCard>
          <SectionLabel>Configuration</SectionLabel>
          <div className="mt-4 flex flex-col gap-4">
            <FormField label="Zone">
              <select className={selectCls} value={activeId} onChange={(e) => setZoneId(e.target.value)}>
                {list.length === 0 && <option value="">No zones yet — create one first</option>}
                {list.map((z) => <option key={z.id} value={z.id}>{z.name} · {z.size} · {z.siteName ?? z.siteId}</option>)}
              </select>
            </FormField>
            <label className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 text-sm">
              <div>
                <div className="font-medium text-foreground">Async loading</div>
                <div className="text-xs text-muted-foreground">Non-blocking script tag; recommended.</div>
              </div>
              <input type="checkbox" checked={asyncLoad} onChange={(e) => setAsyncLoad(e.target.checked)} className="size-4 accent-primary" />
            </label>
            <label className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 text-sm">
              <div>
                <div className="font-medium text-foreground">Responsive container</div>
                <div className="text-xs text-muted-foreground">Fills available width up to the zone size.</div>
              </div>
              <input type="checkbox" checked={responsive} onChange={(e) => setResponsive(e.target.checked)} className="size-4 accent-primary" />
            </label>
            <FormField label="Custom container CSS" hint="Optional overrides for the wrapper div.">
              <input value={customStyle} onChange={(e) => setCustomStyle(e.target.value)} className="h-9 w-full rounded-lg border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring/40" placeholder="margin:16px 0;" />
            </FormField>
          </div>
        </SurfaceCard>

        <div className="space-y-6">
          <SurfaceCard>
            <div className="flex items-start justify-between gap-3">
              <SectionLabel>Embed snippet</SectionLabel>
              <button onClick={copy} disabled={!activeId} className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted/50 disabled:opacity-50">
                <Copy className="size-3.5" /> Copy
              </button>
            </div>
            <pre className="font-mono mt-4 max-h-64 overflow-auto rounded-lg border border-border bg-muted/40 p-4 text-[11px] leading-relaxed text-foreground">{code}</pre>
          </SurfaceCard>

          <SurfaceCard>
            <SectionLabel>Live preview</SectionLabel>
            <div className="mt-4 grid place-items-center rounded-lg border border-dashed border-border bg-muted/40 p-6">
              {selected ? (
                <div
                  className="grid place-items-center border border-border bg-card text-center text-xs text-muted-foreground"
                  style={{ width: responsive ? "100%" : w, maxWidth: w, height: h }}
                >
                  <div>
                    <Code2 className="mx-auto mb-1 size-4" />
                    Ad placeholder<br />{selected.size} · {selected.format}
                  </div>
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">Select a zone to see the preview.</div>
              )}
            </div>
          </SurfaceCard>
        </div>
      </div>
    </div>
  );
}
