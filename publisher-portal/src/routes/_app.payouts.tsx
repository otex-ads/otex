import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Wallet as WalletIcon, Phone, Info, CheckCircle, Settings } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { EmptyState } from "@/components/EmptyState";
import { Modal, FormField, inputCls, ModalActions } from "@/components/Modal";
import { api, type Balance, type Payout, type RecipientStatus } from "@/lib/api";
import { USD, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/payouts")({
  component: PayoutsPage,
});

function toneFor(status: Payout["status"]) {
  return status === "paid" ? "success" : status === "rejected" || status === "failed" ? "danger" : status === "processing" ? "info" : "warning";
}

function PayoutsPage() {
  const balance = useQuery<Balance>({ queryKey: ["balance"], queryFn: api.balance, retry: false });
  const payouts = useQuery<Payout[]>({ queryKey: ["payouts"], queryFn: api.listPayouts, retry: false });
  const recipientQ = useQuery<RecipientStatus>({ queryKey: ["recipient"], queryFn: api.getRecipientStatus, retry: false });
  const qc = useQueryClient();
  const [requesting, setRequesting] = useState(false);
  const [settingUp, setSettingUp] = useState(false);

  const req = useMutation({
    mutationFn: (input: { amount: number; method: string }) => api.requestPayout(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["payouts"] });
      qc.invalidateQueries({ queryKey: ["balance"] });
      toast.success("Payout requested — transfer will be initiated automatically");
      setRequesting(false);
    },
    onError: (e: Error) => toast.error(e.message),
  });

  const list = payouts.data ?? [];
  const b = balance.data ?? { available: 0, pending: 0, totalPaid: 0 };
  const recipient = recipientQ.data;

  const handleRequestPayout = () => {
    if (!recipient?.configured) {
      toast.error("Please set up your payout details first");
      setSettingUp(true);
      return;
    }
    setRequesting(true);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Payouts"
        description="Track earnings, request payouts and manage payout details."
        actions={
          <div className="flex gap-2">
            <button onClick={() => setSettingUp(true)} className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-2 text-sm font-medium text-foreground shadow-sm hover:bg-muted">
              <Settings className="size-4" /> Payout details
            </button>
            <button onClick={handleRequestPayout} className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110">
              <Plus className="size-4" /> Request payout
            </button>
          </div>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <SurfaceCard>
          <div className="label-eyebrow">Available balance</div>
          <div className="num display mt-2 text-4xl font-bold tracking-tight text-foreground">{USD(b.available)}</div>
          <div className="mt-2 text-xs text-muted-foreground">Ready to withdraw</div>
        </SurfaceCard>
        <SurfaceCard delay={0.05}>
          <div className="label-eyebrow">Pending</div>
          <div className="num display mt-2 text-4xl font-bold tracking-tight text-foreground">{USD(b.pending)}</div>
          <div className="mt-2 text-xs text-muted-foreground">Held for verification</div>
        </SurfaceCard>
        <SurfaceCard delay={0.1}>
          <div className="label-eyebrow">Total paid</div>
          <div className="num display mt-2 text-4xl font-bold tracking-tight text-foreground">{USD(b.totalPaid)}</div>
          <div className="mt-2 text-xs text-muted-foreground">Lifetime</div>
        </SurfaceCard>
      </div>

      {recipient?.configured && (
        <div className="flex items-start gap-3 rounded-xl border border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950/30 p-4 text-sm">
          <CheckCircle className="mt-0.5 size-4 shrink-0 text-green-600 dark:text-green-400" />
          <div>
            <div className="font-medium text-foreground">Payout details configured</div>
            <div className="text-muted-foreground">
              M-Pesa · {recipient.recipient?.name} · {recipient.recipient?.phone}
            </div>
          </div>
        </div>
      )}

      {!recipient?.configured && (
        <div className="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/30 p-4 text-sm">
          <Info className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400" />
          <div>
            <div className="font-medium text-foreground">Set up payout details</div>
            <div className="text-muted-foreground">
              Register your M-Pesa number to receive automatic payouts.{" "}
              <button onClick={() => setSettingUp(true)} className="underline font-medium text-foreground">Set up now</button>
            </div>
          </div>
        </div>
      )}

      <SurfaceCard className="p-0">
        <div className="px-6 py-5"><SectionLabel>Payout history</SectionLabel></div>
        {list.length === 0 ? (
          <EmptyState icon={WalletIcon} title="No payouts yet" description="Once you request a payout it will appear here." />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Date</th>
                  <th className="px-6 py-3 text-right font-medium">Amount</th>
                  <th className="px-6 py-3 font-medium">Method</th>
                  <th className="px-6 py-3 font-medium">Reference</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {list.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0">
                    <td className="px-6 py-4 text-muted-foreground">{formatDate(p.createdAt)}</td>
                    <td className="num px-6 py-4 text-right font-semibold text-foreground">{USD(p.amount)}</td>
                    <td className="px-6 py-4 text-muted-foreground">{p.method}</td>
                    <td className="font-mono px-6 py-4 text-xs text-muted-foreground">{p.reference ?? "—"}</td>
                    <td className="px-6 py-4">
                      <StatusPill status={p.status[0].toUpperCase() + p.status.slice(1)} tone={toneFor(p.status)} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </SurfaceCard>

      {requesting && (
        <RequestPayoutModal
          available={b.available}
          onClose={() => setRequesting(false)}
          onSubmit={(v) => req.mutate(v)}
          loading={req.isPending}
        />
      )}

      {settingUp && (
        <SetupRecipientModal
          onClose={() => setSettingUp(false)}
          onSuccess={() => {
            qc.invalidateQueries({ queryKey: ["recipient"] });
            setSettingUp(false);
          }}
        />
      )}
    </div>
  );
}

function RequestPayoutModal({ available, onClose, onSubmit, loading }: {
  available: number; onClose: () => void; onSubmit: (v: { amount: number; method: string }) => void; loading?: boolean;
}) {
  const [amount, setAmount] = useState<number | "">(available > 0 ? available : "");

  const submit = () => {
    const n = Number(amount);
    if (!n || n <= 0) return toast.error("Enter an amount.");
    if (n > available) return toast.error("Amount exceeds available balance.");
    onSubmit({ amount: n, method: "M-Pesa" });
  };

  return (
    <Modal open onClose={onClose} title="Request payout" description={`Available balance: ${USD(available)}`}>
      <div className="flex flex-col gap-4">
        <FormField label="Amount (KES)" required>
          <input type="number" className={inputCls} value={amount} onChange={(e) => setAmount(e.target.value === "" ? "" : Number(e.target.value))} placeholder="0.00" min={0} step="0.01" />
        </FormField>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Phone className="size-3.5" /> Funds will be sent to your registered M-Pesa number
        </div>
      </div>
      <ModalActions onCancel={onClose} onConfirm={submit} loading={loading} confirmLabel="Submit request" />
    </Modal>
  );
}

function SetupRecipientModal({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [saving, setSaving] = useState(false);

  const submit = async () => {
    if (!name.trim()) return toast.error("Enter your full name");
    if (!phone.match(/^(?:\+?254|0)?7\d{8}$/)) return toast.error("Enter a valid Kenyan M-Pesa number");
    setSaving(true);
    try {
      await api.saveRecipient({ name: name.trim(), phone: phone.trim(), email: email.trim() || undefined });
      toast.success("Payout details saved successfully");
      onSuccess();
    } catch (e: any) {
      toast.error(e.message || "Failed to save payout details");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal open onClose={onClose} title="Set up payout details" description="Register your M-Pesa number to receive payouts">
      <div className="flex flex-col gap-4">
        <FormField label="Full name (as on M-Pesa)" required>
          <input type="text" className={inputCls} value={name} onChange={(e) => setName(e.target.value)} placeholder="John Doe" />
        </FormField>
        <FormField label="M-Pesa phone number" required>
          <input type="tel" className={inputCls} value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="07XXXXXXXX" />
        </FormField>
        <FormField label="Email (optional)">
          <input type="email" className={inputCls} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
        </FormField>
      </div>
      <ModalActions onCancel={onClose} onConfirm={submit} loading={saving} confirmLabel="Save details" />
    </Modal>
  );
}
