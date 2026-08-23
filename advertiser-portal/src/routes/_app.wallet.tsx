import { createFileRoute, useSearch } from "@tanstack/react-router";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import { Wallet as WalletIcon, ArrowDownLeft, ArrowUpRight, CreditCard, Mail } from "lucide-react";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { PageHeader } from "@/components/PageHeader";
import { BillingTermsCard } from "@/components/BillingTermsCard";
import { store, useStore } from "@/lib/store";
import { KES, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/wallet")({
  component: WalletPage,
});

const QUICK = [500, 1000, 5000, 10000, 25000];

const TX_LABELS: Record<string, string> = {
  topup: "Top-up",
  spend: "Spend",
  payout: "Payout",
  refund: "Refund",
};

function WalletPage() {
  const balance = useStore((s) => s.balance) || 0;
  const wallet = useStore((s) => s.wallet) || [];
  const [amount, setAmount] = useState<number>(1000);
  const [email, setEmail] = useState<string>("");
  const [pending, setPending] = useState(false);
  const [verified, setVerified] = useState(false);

  // Auto-verify on return from Paystack
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const ref = params.get("reference") || params.get("trxref");
    if (ref && !verified) {
      setVerified(true);
      toast.loading("Verifying payment…", { id: "verify" });
      store
        .verifyTopUp(ref)
        .then((resp) => {
          if (resp.status === "success") {
            toast.success(`Payment verified · ${KES(resp.amount / 100)} added`, { id: "verify" });
            store.refresh();
          } else {
            toast.error("Payment verification failed", { id: "verify" });
          }
          // Clean URL
          window.history.replaceState({}, "", window.location.pathname);
        })
        .catch(() => {
          toast.error("Failed to verify payment", { id: "verify" });
        });
    }
  }, [verified]);

  // Pre-fill email from localStorage if available
  useEffect(() => {
    try {
      const raw = localStorage.getItem("user");
      if (raw) {
        const u = JSON.parse(raw);
        if (u.email) setEmail(u.email);
      }
    } catch {
      // Ignore localStorage errors
    }
  }, []);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/))
      return toast.error("Enter a valid email address");
    if (amount < 10) return toast.error("Minimum top-up is KES 10");
    setPending(true);
    try {
      toast.loading("Redirecting to payment…", { id: "topup" });
      await store.topUp(amount, email);
      // If no redirect happened (shouldn't reach here normally)
      setPending(false);
    } catch (err) {
      setPending(false);
      toast.error((err as Error).message || "Failed to initiate payment", { id: "topup" });
    }
  };

  return (
    <div>
      <PageHeader title="Wallet" description="Fund your account and track every transaction." />

      <BillingTermsCard className="mb-6" />

      <div className="grid gap-6 lg:grid-cols-3">
        <SurfaceCard className="lg:col-span-1 relative overflow-hidden bg-anchor text-anchor-foreground">
          <div className="absolute -right-10 -top-10 size-40 rounded-full bg-white/5 blur-2xl" />
          <div className="label-eyebrow text-anchor-foreground/60">Available balance</div>
          <div className="num mt-4 text-4xl font-semibold">{KES(balance)}</div>
          <div className="mt-2 text-xs text-anchor-foreground/60">
            Spendable across all active campaigns
          </div>
          <div className="mt-6 flex items-center gap-2 border-t border-white/10 pt-4 text-xs text-anchor-foreground/70">
            <WalletIcon className="size-4" /> Real-time balance
          </div>
        </SurfaceCard>

        <SurfaceCard className="lg:col-span-2">
          <SectionLabel>Top up via Paystack</SectionLabel>
          <form onSubmit={submit} className="space-y-4">
            <div>
              <div className="mb-2 text-xs font-medium">Quick amount (KES)</div>
              <div className="flex flex-wrap gap-2">
                {QUICK.map((v) => (
                  <button
                    key={v}
                    type="button"
                    onClick={() => setAmount(v)}
                    className={`num rounded-full border px-3.5 py-1.5 text-sm font-medium transition-colors ${
                      amount === v
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border bg-card text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {v.toLocaleString()}
                  </button>
                ))}
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block">
                <div className="mb-1.5 text-xs font-medium">Amount (KES)</div>
                <input
                  type="number"
                  min={10}
                  step={10}
                  value={amount}
                  onChange={(e) => setAmount(+e.target.value)}
                  className="num h-11 w-full rounded-lg border border-border bg-card px-3 text-sm"
                />
              </label>
              <label className="block">
                <div className="mb-1.5 text-xs font-medium">Email address</div>
                <div className="relative">
                  <Mail className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <input
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                    type="email"
                    className="h-11 w-full rounded-lg border border-border bg-card pl-9 pr-3 text-sm"
                  />
                </div>
              </label>
            </div>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <CreditCard className="size-4" />
              Supports M-Pesa, cards, bank transfers, and USSD
            </div>
            <button
              type="submit"
              disabled={pending}
              className="inline-flex items-center gap-2 rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground disabled:opacity-60"
            >
              {pending ? "Redirecting…" : `Top up ${KES(amount)}`}
            </button>
          </form>
        </SurfaceCard>
      </div>

      <SurfaceCard className="mt-6 p-0">
        <div className="px-6 py-5">
          <SectionLabel>Wallet activity</SectionLabel>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="px-6 py-3 font-medium">Type</th>
                <th className="px-6 py-3 font-medium">Reference</th>
                <th className="px-6 py-3 font-medium">When</th>
                <th className="px-6 py-3 text-right font-medium">Amount</th>
              </tr>
            </thead>
            <tbody>
              {wallet.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-muted-foreground text-sm">
                    No transactions yet
                  </td>
                </tr>
              )}
              {wallet.map((tx) => (
                <tr key={tx.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-4">
                    <div
                      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        tx.type === "topup" || tx.type === "refund"
                          ? "bg-success-soft text-success"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      {tx.type === "topup" || tx.type === "refund" ? (
                        <ArrowDownLeft className="size-3" />
                      ) : (
                        <ArrowUpRight className="size-3" />
                      )}
                      {TX_LABELS[tx.type] || tx.type}
                    </div>
                  </td>
                  <td className="num px-6 py-4 text-xs text-muted-foreground">{tx.reference}</td>
                  <td className="px-6 py-4 text-xs text-muted-foreground">
                    {formatDate(tx.createdAt, true)}
                  </td>
                  <td
                    className={`num px-6 py-4 text-right font-medium ${tx.type === "topup" || tx.type === "refund" ? "text-success" : "text-foreground"}`}
                  >
                    {tx.amount > 0 ? "+" : "−"}
                    {KES(Math.abs(tx.amount))}
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
