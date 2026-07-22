import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { DollarSign, ArrowDownLeft, ArrowUpRight, RefreshCw, Landmark, AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { api } from "@/lib/api";
import { USD, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/revenue")({
  component: RevenuePage,
});

function KES(cents: number) {
  return USD(cents / 100);
}

function toneFor(s: string) {
  return s === "paid" ? "success" : s === "failed" || s === "rejected" ? "danger" : s === "processing" ? "info" : "warning";
}

function RevenuePage() {
  const financials = useQuery({ queryKey: ["financials"], queryFn: api.getFinancials, retry: false });
  const gateway = useQuery({ queryKey: ["gateway-balance"], queryFn: api.getGatewayBalance, retry: false });
  const deposits = useQuery({ queryKey: ["deposits"], queryFn: api.listDeposits, retry: false });
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

  const f = financials.data;
  const g = gateway.data;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Financials</h1>
        <p className="mt-2 text-sm text-muted-foreground">Platform treasury, gateway balance, deposits and payouts</p>
      </div>

      {/* Treasury overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <SurfaceCard>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground"><ArrowDownLeft className="size-3.5 text-green-500" /> Total Deposits</div>
          <div className="num mt-2 text-2xl font-bold">{f ? KES(f.total_deposits) : "—"}</div>
        </SurfaceCard>
        <SurfaceCard>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground"><ArrowUpRight className="size-3.5 text-red-500" /> Ad Spend</div>
          <div className="num mt-2 text-2xl font-bold">{f ? KES(f.total_ad_spend) : "—"}</div>
        </SurfaceCard>
        <SurfaceCard>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground"><DollarSign className="size-3.5 text-blue-500" /> Platform Fees</div>
          <div className="num mt-2 text-2xl font-bold">{f ? KES(f.total_platform_fees) : "—"}</div>
        </SurfaceCard>
        <SurfaceCard>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground"><AlertTriangle className="size-3.5 text-amber-500" /> Pending Payouts</div>
          <div className="num mt-2 text-2xl font-bold">{f ? KES(f.pending_payouts) : "—"}</div>
        </SurfaceCard>
      </div>

      {/* Gateway balance */}
      <SurfaceCard>
        <SectionLabel>Paystack Gateway Balance</SectionLabel>
        {g?.available ? (
          <div className="mt-3 flex flex-wrap gap-6">
            {g.balances.map((b, i) => (
              <div key={i} className="flex items-center gap-3">
                <Landmark className="size-5 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">{b.currency}</div>
                  <div className="num text-xl font-bold">{USD(b.balance / 100)}</div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-muted-foreground">Gateway balance not available — check PAYSTACK_SECRET_KEY config.</p>
        )}
      </SurfaceCard>

      {/* Recent deposits */}
      <SurfaceCard className="p-0">
        <div className="px-6 py-5"><SectionLabel>Recent Deposits</SectionLabel></div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Date</th>
                <th className="px-6 py-3 font-medium">Account</th>
                <th className="px-6 py-3 font-medium">Reference</th>
                <th className="px-6 py-3 text-right font-medium">Amount</th>
              </tr>
            </thead>
            <tbody>
              {(deposits.data ?? []).length === 0 && (
                <tr><td colSpan={4} className="px-6 py-8 text-center text-sm text-muted-foreground">No deposits yet</td></tr>
              )}
              {(deposits.data ?? []).map((d) => (
                <tr key={d.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-3 text-xs text-muted-foreground">{formatDate(d.created_at, true)}</td>
                  <td className="px-6 py-3 font-mono text-xs text-muted-foreground">{d.account_id.slice(0, 8)}…</td>
                  <td className="px-6 py-3 font-mono text-xs text-muted-foreground">{d.reference ?? "—"}</td>
                  <td className="num px-6 py-3 text-right font-semibold text-green-600">{KES(d.amount_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SurfaceCard>

      {/* All payouts */}
      <SurfaceCard className="p-0">
        <div className="px-6 py-5"><SectionLabel>All Payouts</SectionLabel></div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Date</th>
                <th className="px-6 py-3 font-medium">Publisher</th>
                <th className="px-6 py-3 text-right font-medium">Amount</th>
                <th className="px-6 py-3 font-medium">Status</th>
                <th className="px-6 py-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {(payouts.data ?? []).length === 0 && (
                <tr><td colSpan={5} className="px-6 py-8 text-center text-sm text-muted-foreground">No payouts yet</td></tr>
              )}
              {(payouts.data ?? []).map((p) => (
                <tr key={p.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-3 text-xs text-muted-foreground">{formatDate(p.requested_at, true)}</td>
                  <td className="px-6 py-3 font-mono text-xs text-muted-foreground">{p.publisher_id.slice(0, 8)}…</td>
                  <td className="num px-6 py-3 text-right font-semibold text-foreground">{KES(p.amount_cents)}</td>
                  <td className="px-6 py-3">
                    <StatusPill status={p.status[0].toUpperCase() + p.status.slice(1)} tone={toneFor(p.status)} />
                  </td>
                  <td className="px-6 py-3">
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
      </SurfaceCard>
    </div>
  );
}
