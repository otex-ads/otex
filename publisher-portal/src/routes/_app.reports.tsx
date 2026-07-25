import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Download, Plus, CalendarClock } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { Modal, FormField, inputCls, selectCls, ModalActions } from "@/components/Modal";
import { api, type StatsResponse } from "@/lib/api";
import { USD, Num } from "@/lib/format";

export const Route = createFileRoute("/_app/reports")({
  component: ReportsPage,
});

const TABS = ["Performance", "Geographic", "Device", "Traffic"] as const;
type Tab = (typeof TABS)[number];

interface Scheduled {
  id: string;
  name: string;
  cadence: string;
  email: string;
}

function isoDaysAgo(d: number) {
  return new Date(Date.now() - d * 86400_000).toISOString().slice(0, 10);
}

function ReportsPage() {
  const [tab, setTab] = useState<Tab>("Performance");
  const [from, setFrom] = useState(isoDaysAgo(30));
  const [to, setTo] = useState(isoDaysAgo(0));
  const [scheduled, setScheduled] = useState<Scheduled[]>([]);
  const [creating, setCreating] = useState(false);

  const stats = useQuery<StatsResponse>({
    queryKey: ["stats", "reports", from, to],
    queryFn: () => api.stats({ from, to }),
    retry: false,
  });

  const rowsByTab = () => {
    const d = stats.data;
    if (!d) return [];
    if (tab === "Performance") return d.bySite ?? [];
    if (tab === "Geographic") return d.byCountry ?? [];
    if (tab === "Device") return d.byDevice ?? [];
    return d.byBrowser ?? []; // Traffic ~ browsers/referrers
  };
  const rows = rowsByTab();

  const exportCsv = () => {
    const header = ["Key", "Impressions", "Clicks", "Revenue"];
    const csv = [header, ...rows.map((r) => [r.key, r.impressions, r.clicks, r.revenue])]
      .map((r) => r.join(","))
      .join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${tab.toLowerCase()}-${from}-to-${to}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Reports"
        description="Custom reports and scheduled deliveries."
        actions={
          <button
            onClick={exportCsv}
            className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
          >
            <Download className="size-4" /> Export
          </button>
        }
      />

      <SurfaceCard>
        <div className="flex flex-wrap items-end gap-4">
          <div className="flex flex-wrap gap-1 rounded-full border border-border bg-muted/40 p-1">
            {TABS.map((t) => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`rounded-full px-3 py-1 text-xs font-medium ${tab === t ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
              >
                {t}
              </button>
            ))}
          </div>
          <div className="ml-auto flex flex-wrap items-end gap-3">
            <FormField label="From">
              <input
                type="date"
                className={inputCls}
                value={from}
                onChange={(e) => setFrom(e.target.value)}
              />
            </FormField>
            <FormField label="To">
              <input
                type="date"
                className={inputCls}
                value={to}
                onChange={(e) => setTo(e.target.value)}
              />
            </FormField>
          </div>
        </div>
      </SurfaceCard>

      <SurfaceCard className="p-0">
        <div className="px-6 py-5">
          <SectionLabel>{tab} report</SectionLabel>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">
                  {tab === "Geographic"
                    ? "Country"
                    : tab === "Device"
                      ? "Device"
                      : tab === "Traffic"
                        ? "Referrer"
                        : "Site"}
                </th>
                <th className="px-6 py-3 text-right font-medium">Impressions</th>
                <th className="px-6 py-3 text-right font-medium">Clicks</th>
                <th className="px-6 py-3 text-right font-medium">Revenue</th>
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-6 py-10 text-center text-muted-foreground">
                    {stats.isLoading ? "Loading…" : "No data for this range."}
                  </td>
                </tr>
              )}
              {rows.map((r) => (
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

      <SurfaceCard>
        <div className="flex items-start justify-between gap-3">
          <div>
            <SectionLabel>Scheduled reports</SectionLabel>
            <p className="mt-1 text-sm text-muted-foreground">
              Auto-email daily, weekly or monthly performance summaries.
            </p>
          </div>
          <button
            onClick={() => setCreating(true)}
            className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted/50"
          >
            <Plus className="size-3.5" /> New schedule
          </button>
        </div>
        <div className="mt-4 space-y-2">
          {scheduled.length === 0 && (
            <div className="rounded-lg border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
              <CalendarClock className="mx-auto mb-2 size-4" /> No scheduled reports yet
            </div>
          )}
          {scheduled.map((s) => (
            <div
              key={s.id}
              className="flex items-center justify-between gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm"
            >
              <div>
                <div className="font-medium text-foreground">{s.name}</div>
                <div className="text-xs text-muted-foreground">
                  {s.cadence} · to {s.email}
                </div>
              </div>
              <button
                onClick={() => setScheduled((all) => all.filter((x) => x.id !== s.id))}
                className="text-xs text-muted-foreground hover:text-danger"
              >
                Remove
              </button>
            </div>
          ))}
        </div>
      </SurfaceCard>

      {creating && (
        <ScheduleModal
          onClose={() => setCreating(false)}
          onSubmit={(s) => {
            setScheduled((all) => [{ ...s, id: Math.random().toString(36).slice(2, 8) }, ...all]);
            setCreating(false);
            toast.success("Schedule saved");
          }}
        />
      )}
    </div>
  );
}

function ScheduleModal({
  onClose,
  onSubmit,
}: {
  onClose: () => void;
  onSubmit: (v: Omit<Scheduled, "id">) => void;
}) {
  const [name, setName] = useState("");
  const [cadence, setCadence] = useState("Weekly");
  const [email, setEmail] = useState("");
  const submit = () => {
    if (!name || !email) return toast.error("Name and email required.");
    onSubmit({ name, cadence, email });
  };
  return (
    <Modal open onClose={onClose} title="Schedule report">
      <div className="flex flex-col gap-4">
        <FormField label="Report name" required>
          <input
            className={inputCls}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Weekly performance"
          />
        </FormField>
        <FormField label="Cadence">
          <select
            className={selectCls}
            value={cadence}
            onChange={(e) => setCadence(e.target.value)}
          >
            <option>Daily</option>
            <option>Weekly</option>
            <option>Monthly</option>
          </select>
        </FormField>
        <FormField label="Deliver to" required>
          <input
            type="email"
            className={inputCls}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
          />
        </FormField>
      </div>
      <ModalActions onCancel={onClose} onConfirm={submit} confirmLabel="Save schedule" />
    </Modal>
  );
}
