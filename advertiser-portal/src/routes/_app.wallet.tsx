import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { Wallet as WalletIcon, ArrowDownLeft, ArrowUpRight, Smartphone } from "lucide-react";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { PageHeader } from "@/components/PageHeader";
import { store, useStore } from "@/lib/store";
import { KES, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/wallet")({
  component: WalletPage,
});

const QUICK = [500, 1000, 5000, 10000, 25000];

function WalletPage() {
  const balance = useStore((s) => s.balance);
  const wallet = useStore((s) => s.wallet);
  const [amount, setAmount] = useState<number>(1000);
  const [phone, setPhone] = useState<string>("");
  const [pending, setPending] = useState(false);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!phone.match(/^(?:\+?254|0)?7\d{8}$/)) return toast.error("Enter a valid Kenyan mobile number");
    if (amount < 10) return toast.error("Minimum top-up is KES 10");
    setPending(true);
    toast.loading("STK push sent — approve on your phone", { id: "mpesa" });
    setTimeout(() => {
      const tx = store.topUp(amount, phone);
      setPending(false);
      toast.success(`Top-up successful · ${tx.reference}`, { id: "mpesa" });
    }, 1600);
  };

  return (
    <div>
      <PageHeader title="Wallet" description="Fund your account via M-Pesa and track every debit." />

      <div className="grid gap-6 lg:grid-cols-3">
        <SurfaceCard className="lg:col-span-1 relative overflow-hidden bg-anchor text-anchor-foreground">
          <div className="absolute -right-10 -top-10 size-40 rounded-full bg-white/5 blur-2xl" />
          <div className="label-eyebrow text-anchor-foreground/60">Available balance</div>
          <div className="num mt-4 text-4xl font-semibold">{KES(balance)}</div>
          <div className="mt-2 text-xs text-anchor-foreground/60">Spendable across all active campaigns</div>
          <div className="mt-6 flex items-center gap-2 border-t border-white/10 pt-4 text-xs text-anchor-foreground/70">
            <WalletIcon className="size-4" /> Real-time balance
          </div>
        </SurfaceCard>

        <SurfaceCard className="lg:col-span-2">
          <SectionLabel>Top up via M-Pesa</SectionLabel>
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
                      amount === v ? "border-primary bg-primary text-primary-foreground" : "border-border bg-card text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {v.toLocaleString()}
                  </button>
                ))}
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block">
                <div className="mb-1.5 text-xs font-medium">Amount</div>
                <input type="number" min={10} step={10} value={amount} onChange={(e) => setAmount(+e.target.value)}
                  className="num h-11 w-full rounded-lg border border-border bg-card px-3 text-sm" />
              </label>
              <label className="block">
                <div className="mb-1.5 text-xs font-medium">M-Pesa phone</div>
                <div className="relative">
                  <Smartphone className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="07XXXXXXXX"
                    className="h-11 w-full rounded-lg border border-border bg-card pl-9 pr-3 text-sm" />
                </div>
              </label>
            </div>
            <button
              type="submit"
              disabled={pending}
              className="inline-flex items-center gap-2 rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground disabled:opacity-60"
            >
              {pending ? "Waiting for STK approval…" : `Top up ${KES(amount)}`}
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
                <th className="px-6 py-3 font-medium">Details</th>
                <th className="px-6 py-3 font-medium">When</th>
                <th className="px-6 py-3 text-right font-medium">Amount</th>
              </tr>
            </thead>
            <tbody>
              {wallet.map((tx) => (
                <tr key={tx.id} className="border-b border-border last:border-0">
                  <td className="px-6 py-4">
                    <div className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      tx.type === "topup" ? "bg-success-soft text-success" : "bg-muted text-muted-foreground"
                    }`}>
                      {tx.type === "topup" ? <ArrowDownLeft className="size-3" /> : <ArrowUpRight className="size-3" />}
                      {tx.type === "topup" ? "Top-up" : "Spend"}
                    </div>
                  </td>
                  <td className="num px-6 py-4 text-xs text-muted-foreground">{tx.reference}</td>
                  <td className="px-6 py-4 text-xs text-muted-foreground">
                    {tx.type === "topup" ? `M-Pesa · ${tx.phone ?? ""}` : `Campaign spend`}
                  </td>
                  <td className="px-6 py-4 text-xs text-muted-foreground">{formatDate(tx.createdAt, true)}</td>
                  <td className={`num px-6 py-4 text-right font-medium ${tx.type === "topup" ? "text-success" : "text-foreground"}`}>
                    {tx.type === "topup" ? "+" : "−"}{KES(tx.amount)}
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
