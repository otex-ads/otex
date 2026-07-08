import { cn } from "@/lib/utils";

type Tone = "success" | "warning" | "danger" | "neutral" | "info";

const toneStyles: Record<Tone, string> = {
  success: "bg-success-soft text-success",
  warning: "bg-warning-soft text-warning",
  danger: "bg-danger-soft text-danger",
  neutral: "bg-muted text-muted-foreground",
  info: "bg-accent text-accent-foreground",
};

const mapping: Record<string, Tone> = {
  // KYC
  Verified: "success",
  Pending: "warning",
  Flagged: "danger",
  // Tx
  Completed: "success",
  Failed: "danger",
  // Card / account
  Active: "success",
  Frozen: "warning",
  Blocked: "danger",
  Dormant: "warning",
  Closed: "neutral",
  // Loan stages
  Applied: "neutral",
  "Committee Review": "info",
  Approved: "info",
  Disbursed: "info",
  Arrears: "danger",
  // Approval / disbursement
  Rejected: "danger",
  Confirmed: "success",
  "In Flight": "info",
  Accepted: "success",
  Declined: "danger",
  Posted: "success",
  Draft: "neutral",
};

export function StatusPill({ status, tone }: { status: string; tone?: Tone }) {
  const t = tone ?? mapping[status] ?? "neutral";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium",
        toneStyles[t],
      )}
    >
      <span className={cn("size-1.5 rounded-full", {
        "bg-success": t === "success",
        "bg-warning": t === "warning",
        "bg-danger": t === "danger",
        "bg-muted-foreground": t === "neutral",
        "bg-primary": t === "info",
      })} />
      {status}
    </span>
  );
}