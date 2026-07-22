import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Wallet as WalletIcon, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { EmptyState } from "@/components/EmptyState";
import { api } from "@/lib/api";
import { USD, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/payouts")({
  component: AdminPayoutsPage,
});

function KES(cents: number) {
  return USD(cents / 100);
}

function toneFor(s: string) {
  return s === "paid" ? "success" : s === "failed" || s === "rejected" ? "danger" : s === "processing" ? "info" : "warning";
}

function AdminPayoutsPage() {
  const payouts = useQuery({ queryKey: ["all-payouts"], queryFn: api.listAllPayouts, retry: false });
  const qc = useQueryClient();

  const retryMut = useMutation({
    mutationFn: (id: string) => api.retryPayout(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["all-payouts"] });
      toast.success("Payout queued for retry");
    },
    onError: (e: Error) => toast.error(e.message),
  });

  const list = payouts.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Publisher Payouts"
        description="View and manage all publisher payout requests across the platform."
      />

      <SurfaceCard className="p-0">
        <div className="px-6 py-5"><SectionLabel>All payouts</SectionLabel></div>
        {list.length === 0 ? (
          <EmptyState icon={WalletIcon} title="No payouts yet" description="Publisher payout requests will appear here." />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Date</th>
                  <th className="px-6 py-3 font-medium">Publisher</th>
                  <th className="px-6 py-3 text-right font-medium">Amount</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 font-medium">Processed</th>
                  <th className="px-6 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {list.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0">
                    <td className="px-6 py-4 text-xs text-muted-foreground">{formatDate(p.requested_at, true)}</td>
                    <td className="px-6 py-4 font-mono text-xs text-muted-foreground">{p.publisher_id.slice(0, 8)}…</td>
                    <td className="num px-6 py-4 text-right font-semibold text-foreground">{KES(p.amount_cents)}</td>
                    <td className="px-6 py-4">
                      <StatusPill status={p.status[0].toUpperCase() + p.status.slice(1)} tone={toneFor(p.status)} />
                    </td>
                    <td className="px-6 py-4 text-xs text-muted-foreground">{p.processed_at ? formatDate(p.processed_at, true) : "—"}</td>
                    <td className="px-6 py-4">
                      {p.status === "failed" && (
                        <button
                          onClick={() => retryMut.mutate(p.id)}
                          disabled={retryMut.isPending}
                          className="inline-flex items-center gap-1 rounded-md bg-amber-100 px-2 py-1 text-xs font-medium text-amber-800 hover:bg-amber-200 dark:bg-amber-900/30 dark:text-amber-300"
                        >
                          <RefreshCw className="size-3" /> Retry
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </SurfaceCard>
    </div>
  );
}
