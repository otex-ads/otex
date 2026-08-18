import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Pencil, Trash2, LayoutGrid } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { EmptyState } from "@/components/EmptyState";
import { Modal, FormField, inputCls, selectCls, ModalActions } from "@/components/Modal";
import { api, type Zone, type AdSize, type AdFormat, type Site } from "@/lib/api";
import { USD, Num } from "@/lib/format";

export const Route = createFileRoute("/_app/zones")({
  component: ZonesPage,
});

const SIZES: { value: AdSize; label: string }[] = [
  { value: "728x90", label: "Leaderboard 728×90" },
  { value: "300x250", label: "Rectangle 300×250" },
  { value: "160x600", label: "Skyscraper 160×600" },
  { value: "970x250", label: "Billboard 970×250" },
  { value: "320x50", label: "Mobile 320×50" },
];
const FORMATS: { value: AdFormat; label: string }[] = [
  { value: "banner", label: "Banner" },
  { value: "native", label: "Native" },
  { value: "push", label: "Push" },
  { value: "popunder", label: "Popunder" },
  { value: "interstitial", label: "Interstitial" },
  { value: "in_page_push", label: "In-Page Push" },
];

interface ZoneForm {
  name: string;
  siteId: string;
  size: AdSize;
  format: AdFormat;
  countries: string[];
  deviceTypes: string[];
  os: string[];
  browsers: string[];
  carriers: string[];
  connectionTypes: string[];
}

function ZonesPage() {
  const zones = useQuery<Zone[]>({
    queryKey: ["zones"],
    queryFn: () => api.listZones(),
    retry: false,
  });
  const sites = useQuery<Site[]>({ queryKey: ["sites"], queryFn: api.listSites, retry: false });
  const targetingOptions = useQuery({
    queryKey: ["targeting-options"],
    queryFn: () => api.getTargetingOptions(),
    retry: false,
  });
  const campaigns = useQuery({
    queryKey: ["campaigns"],
    queryFn: () => api.listCampaigns(),
    retry: false,
  });
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: ["zones"] });

  const create = useMutation({
    mutationFn: (input: ZoneForm) => api.createZone(input),
    onSuccess: () => {
      invalidate();
      toast.success("Zone created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
  const update = useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<Zone> }) => api.updateZone(id, patch),
    onSuccess: () => {
      invalidate();
      toast.success("Zone updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
  const remove = useMutation({
    mutationFn: (id: string) => api.deleteZone(id),
    onSuccess: () => {
      invalidate();
      toast.success("Zone deleted");
    },
    onError: (e: Error) => toast.error(e.message),
  });

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Zone | null>(null);
  const [deleting, setDeleting] = useState<Zone | null>(null);

  const list = zones.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Zones"
        description="Ad placements within your sites."
        actions={
          <button
            onClick={() => setCreating(true)}
            className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
          >
            <Plus className="size-4" /> Add zone
          </button>
        }
      />

      <SurfaceCard className="p-0">
        {list.length === 0 ? (
          <EmptyState
            icon={LayoutGrid}
            title={zones.isLoading ? "Loading zones…" : "No zones yet"}
            description="Create a zone to generate ad code for one of your sites."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Zone</th>
                  <th className="px-6 py-3 font-medium">Site</th>
                  <th className="px-6 py-3 font-medium">Size</th>
                  <th className="px-6 py-3 font-medium">Format</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 text-right font-medium">Revenue</th>
                  <th className="px-6 py-3 text-right font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {list.map((z) => (
                  <tr key={z.id} className="border-b border-border last:border-0">
                    <td className="px-6 py-4">
                      <Link
                        to="/zones/$zoneId"
                        params={{ zoneId: z.id }}
                        className="font-medium text-foreground hover:text-primary"
                      >
                        {z.name}
                      </Link>
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">{z.siteName ?? z.siteId}</td>
                    <td className="px-6 py-4 text-muted-foreground">{z.size}</td>
                    <td className="px-6 py-4 capitalize text-muted-foreground">{z.format}</td>
                    <td className="px-6 py-4">
                      <StatusPill
                        status={z.status[0].toUpperCase() + z.status.slice(1)}
                        tone={z.status === "active" ? "success" : "warning"}
                      />
                    </td>
                    <td className="num px-6 py-4 text-right">{USD(z.revenue)}</td>
                    <td className="px-6 py-4">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => setEditing(z)}
                          className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                        >
                          <Pencil className="size-3.5" />
                        </button>
                        <button
                          onClick={() => setDeleting(z)}
                          className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-danger-soft hover:text-danger"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="border-t border-border px-6 py-3 text-right text-xs text-muted-foreground">
          {list.length} total · {Num(list.reduce((a, z) => a + z.impressions, 0))} impressions
        </div>
      </SurfaceCard>

      {creating && (
        <ZoneFormModal
          sites={sites.data ?? []}
          targetingOptions={targetingOptions.data}
          campaigns={campaigns.data ?? []}
          onClose={() => setCreating(false)}
          onSubmit={(v) => create.mutate(v, { onSuccess: () => setCreating(false) })}
          loading={create.isPending}
          title="Add zone"
        />
      )}
      {editing && (
        <ZoneFormModal
          zone={editing}
          sites={sites.data ?? []}
          targetingOptions={targetingOptions.data}
          campaigns={campaigns.data ?? []}
          onClose={() => setEditing(null)}
          onSubmit={(v) =>
            update.mutate({ id: editing.id, patch: v }, { onSuccess: () => setEditing(null) })
          }
          loading={update.isPending}
          title="Edit zone"
          onToggle={() =>
            update.mutate(
              {
                id: editing.id,
                patch: { status: editing.status === "active" ? "paused" : "active" },
              },
              { onSuccess: () => setEditing(null) },
            )
          }
        />
      )}
      <Modal
        open={!!deleting}
        onClose={() => setDeleting(null)}
        title="Delete zone?"
        description={deleting ? `"${deleting.name}" will be removed.` : ""}
      >
        <ModalActions
          onCancel={() => setDeleting(null)}
          onConfirm={() =>
            deleting && remove.mutate(deleting.id, { onSuccess: () => setDeleting(null) })
          }
          loading={remove.isPending}
          confirmLabel="Delete"
          confirmVariant="danger"
        />
      </Modal>
    </div>
  );
}

function ZoneFormModal({
  zone,
  sites,
  targetingOptions,
  campaigns,
  onClose,
  onSubmit,
  loading,
  title,
  onToggle,
}: {
  zone?: Zone;
  sites: Site[];
  targetingOptions?: {
    countries?: string[];
    device_types?: string[];
    os?: string[];
    browsers?: string[];
    carriers?: string[];
    connection_types?: string[];
  };
  campaigns: Array<{
    id: string;
    name: string;
    format: string;
    status: string;
    targeting: {
      countries: string[];
      devices: string[];
      os: string[];
    };
  }>;
  onClose: () => void;
  onSubmit: (v: ZoneForm) => void;
  loading?: boolean;
  title: string;
  onToggle?: () => void;
}) {
  const [form, setForm] = useState<ZoneForm>({
    name: zone?.name ?? "",
    siteId: zone?.siteId ?? sites[0]?.id ?? "",
    size: zone?.size ?? "300x250",
    format: zone?.format ?? "banner",
    countries: zone?.countries ?? [],
    deviceTypes: zone?.deviceTypes ?? [],
    os: zone?.os ?? [],
    browsers: zone?.browsers ?? [],
    carriers: zone?.carriers ?? [],
    connectionTypes: zone?.connectionTypes ?? [],
  });
  const [selectedCampaignId, setSelectedCampaignId] = useState<string>("");

  // Check if campaign targeting matches zone targeting
  const getMatchingCampaigns = () => {
    if (!campaigns || campaigns.length === 0) return [];
    return campaigns.filter((campaign) => {
      if (campaign.status !== "active") return false;
      // Check format match
      if (campaign.format !== form.format) return false;
      // Check countries overlap
      if (form.countries.length > 0 && campaign.targeting.countries.length > 0) {
        const hasOverlap = form.countries.some((c) => campaign.targeting.countries.includes(c));
        if (!hasOverlap) return false;
      }
      // Check devices overlap
      if (form.deviceTypes.length > 0 && campaign.targeting.devices.length > 0) {
        const hasOverlap = form.deviceTypes.some((d) => campaign.targeting.devices.includes(d));
        if (!hasOverlap) return false;
      }
      // Check OS overlap
      if (form.os.length > 0 && campaign.targeting.os.length > 0) {
        const hasOverlap = form.os.some((o) => campaign.targeting.os.includes(o));
        if (!hasOverlap) return false;
      }
      return true;
    });
  };

  const matchingCampaigns = getMatchingCampaigns();

  const submit = () => {
    if (!form.name.trim()) return toast.error("Zone name required.");
    if (!form.siteId) return toast.error("Select a site.");
    onSubmit(form);
  };

  return (
    <Modal open onClose={onClose} title={title}>
      <div className="flex flex-col gap-4">
        <FormField label="Zone name" required>
          <input
            className={inputCls}
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="Homepage top banner"
          />
        </FormField>
        <FormField label="Site" required>
          <select
            className={selectCls}
            value={form.siteId}
            onChange={(e) => setForm({ ...form, siteId: e.target.value })}
          >
            {sites.length === 0 && <option value="">No sites — add one first</option>}
            {sites.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name} · {s.domain}
              </option>
            ))}
          </select>
        </FormField>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Ad size">
            <select
              className={selectCls}
              value={form.size}
              onChange={(e) => setForm({ ...form, size: e.target.value as AdSize })}
            >
              {SIZES.map((s) => (
                <option key={s.value} value={s.value}>
                  {s.label}
                </option>
              ))}
            </select>
          </FormField>
          <FormField label="Format">
            <select
              className={selectCls}
              value={form.format}
              onChange={(e) => setForm({ ...form, format: e.target.value as AdFormat })}
            >
              {FORMATS.map((f) => (
                <option key={f.value} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
          </FormField>
        </div>
        <FormField label="Targeting (optional)">
          <div className="space-y-3">
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Countries (ISO codes, e.g., KE, UG, ZA)</label>
              <input
                className={inputCls}
                value={form.countries.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const countries = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, countries });
                }}
                placeholder="KE, UG, ZA"
                list="countries-list"
              />
              <datalist id="countries-list">
                {targetingOptions?.countries?.map((c) => (
                  <option key={c} value={c} />
                ))}
              </datalist>
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Device types</label>
              <input
                className={inputCls}
                value={form.deviceTypes.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const deviceTypes = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, deviceTypes });
                }}
                placeholder="mobile, desktop, tablet"
                list="device-types-list"
              />
              <datalist id="device-types-list">
                {targetingOptions?.device_types?.map((d) => (
                  <option key={d} value={d} />
                ))}
              </datalist>
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Operating systems</label>
              <input
                className={inputCls}
                value={form.os.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const os = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, os });
                }}
                placeholder="android, ios"
                list="os-list"
              />
              <datalist id="os-list">
                {targetingOptions?.os?.map((o) => (
                  <option key={o} value={o} />
                ))}
              </datalist>
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Browsers</label>
              <input
                className={inputCls}
                value={form.browsers.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const browsers = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, browsers });
                }}
                placeholder="chrome, firefox, safari"
                list="browsers-list"
              />
              <datalist id="browsers-list">
                {targetingOptions?.browsers?.map((b) => (
                  <option key={b} value={b} />
                ))}
              </datalist>
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Carriers</label>
              <input
                className={inputCls}
                value={form.carriers.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const carriers = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, carriers });
                }}
                placeholder="safaricom, airtel"
                list="carriers-list"
              />
              <datalist id="carriers-list">
                {targetingOptions?.carriers?.map((c) => (
                  <option key={c} value={c} />
                ))}
              </datalist>
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Connection types</label>
              <input
                className={inputCls}
                value={form.connectionTypes.join(", ")}
                onChange={(e) => {
                  const raw = e.target.value;
                  const connectionTypes = raw.split(",").map(s => s.trim()).filter(Boolean);
                  setForm({ ...form, connectionTypes });
                }}
                placeholder="wifi, 4g, 3g"
                list="connection-types-list"
              />
              <datalist id="connection-types-list">
                {targetingOptions?.connection_types?.map((c) => (
                  <option key={c} value={c} />
                ))}
              </datalist>
            </div>
          </div>
        </FormField>
        <FormField label="Matching Campaigns">
          <div className="space-y-2">
            {matchingCampaigns.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                No active campaigns match this zone's targeting. Adjust targeting or wait for campaigns to be created.
              </p>
            ) : (
              <>
                <select
                  className={selectCls}
                  value={selectedCampaignId}
                  onChange={(e) => setSelectedCampaignId(e.target.value)}
                >
                  <option value="">Select a campaign to view details</option>
                  {matchingCampaigns.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} ({c.format}) · {c.status}
                    </option>
                  ))}
                </select>
                {selectedCampaignId && (
                  <div className="rounded-md bg-muted/50 p-3 text-xs">
                    {(() => {
                      const selected = matchingCampaigns.find((c) => c.id === selectedCampaignId);
                      if (!selected) return null;
                      return (
                        <div className="space-y-1">
                          <p className="font-medium text-foreground">{selected.name}</p>
                          <p className="text-muted-foreground">Format: {selected.format}</p>
                          <p className="text-muted-foreground">Status: {selected.status}</p>
                          <p className="text-muted-foreground">
                            Countries: {selected.targeting.countries.join(", ") || "All"}
                          </p>
                          <p className="text-muted-foreground">
                            Devices: {selected.targeting.devices.join(", ") || "All"}
                          </p>
                          <p className="text-muted-foreground">
                            OS: {selected.targeting.os.join(", ") || "All"}
                          </p>
                        </div>
                      );
                    })()}
                  </div>
                )}
              </>
            )}
          </div>
        </FormField>
        {zone && onToggle && (
          <button
            type="button"
            onClick={onToggle}
            className="self-start rounded-full border border-border bg-card px-4 py-2 text-xs font-medium text-foreground hover:bg-muted/50"
          >
            {zone.status === "active" ? "Pause zone" : "Resume zone"}
          </button>
        )}
      </div>
      <ModalActions
        onCancel={onClose}
        onConfirm={submit}
        loading={loading}
        confirmLabel={zone ? "Save changes" : "Create zone"}
      />
    </Modal>
  );
}
