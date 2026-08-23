import { SurfaceCard } from "@/components/Card";

const TERMS = [
  { label: "Funding method", value: "M-Pesa · Card · Bank transfer" },
  { label: "Minimum spend", value: "None" },
  { label: "Pricing model", value: "CPC / CPM bidding" },
  { label: "Campaign review", value: "Within 24 hours" },
  { label: "Billing currency", value: "KES — no FX markup" },
  { label: "Commitment", value: "Pay-as-you-go, top up any amount" },
];

export function BillingTermsCard({ className }: { className?: string }) {
  return (
    <SurfaceCard className={className}>
      <div className="mb-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        Billing Terms
      </div>
      <h3 className="text-xl font-bold tracking-tight text-foreground">
        No minimum spend. <span className="italic font-normal">Pay as you go.</span>
      </h3>
      <p className="mt-1.5 text-sm text-muted-foreground">
        Top up any amount via M-Pesa, card or bank transfer. Launch a campaign
        the same day — no contracts, no FX surprises.
      </p>

      <div className="mt-5 grid gap-x-8 gap-y-3 sm:grid-cols-2">
        {TERMS.map((t) => (
          <div
            key={t.label}
            className="flex items-baseline justify-between gap-3 border-b border-border/70 pb-2 text-sm"
          >
            <span className="text-muted-foreground">{t.label}</span>
            <span className="font-semibold text-foreground">{t.value}</span>
          </div>
        ))}
      </div>
    </SurfaceCard>
  );
}
