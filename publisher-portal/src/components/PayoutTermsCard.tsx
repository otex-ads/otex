import { Link } from "@tanstack/react-router";
import { SurfaceCard } from "@/components/Card";

const TERMS = [
  { label: "Payout method", value: "M-Pesa · Bank transfer" },
  { label: "Payout frequency", value: "Weekly · Every Friday" },
  { label: "Minimum threshold", value: "KES 500" },
  { label: "Revenue share", value: "70% of net ad revenue" },
  { label: "Onboarding fee", value: "None" },
  { label: "Withdrawal fee", value: "None on M-Pesa" },
];

export function PayoutTermsCard({ className }: { className?: string }) {
  return (
    <SurfaceCard className={className}>
      <div className="mb-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        Payout Terms
      </div>
      <h3 className="text-xl font-bold tracking-tight text-foreground">
        Get paid <span className="italic font-normal">weekly.</span> Straight to M-Pesa.
      </h3>
      <p className="mt-1.5 text-sm text-muted-foreground">
        Keep up to 80% of net ad revenue on your zones. No onboarding fee, no
        withdrawal fee on M-Pesa, no net-30 wait.
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

      <Link
        to="/payouts"
        className="mt-5 inline-flex items-center justify-center rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground shadow-sm transition hover:brightness-110"
      >
        Set up payout details
      </Link>
    </SurfaceCard>
  );
}
