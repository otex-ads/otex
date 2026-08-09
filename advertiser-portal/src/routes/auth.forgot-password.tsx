import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { motion } from "framer-motion";
import { Loader2, ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { api } from "@/lib/api";

export const Route = createFileRoute("/auth/forgot-password")({
  ssr: false,
  component: ForgotPasswordPage,
});

function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) {
      toast.error("Enter your email address.");
      return;
    }
    setLoading(true);
    try {
      await api.requestPasswordReset({ email: email.trim() });
      setSent(true);
      toast.success("Password reset email sent. Check your inbox.");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to send reset email";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid min-h-screen place-items-center bg-background p-4">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-xl">
        <div className="mb-6 flex items-center gap-2.5">
          <img src="/logo.png" alt="OtexAds" className="size-10 object-contain" />
          <div className="leading-tight">
            <div className="text-base font-bold text-foreground">OtexAds</div>
            <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              Advertiser Portal
            </div>
          </div>
        </div>
        <Link to="/auth/login" className="mb-4 inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-3" />
          Back to sign in
        </Link>
        <h1 className="display text-3xl font-normal tracking-tight text-foreground">
          Forgot password?
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {sent
            ? "We've sent a password reset link to your email."
            : "Enter your email and we'll send you a reset link."}
        </p>

        {!sent && (
          <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-4">
            <div>
              <label className="text-xs font-medium text-foreground">Email</label>
              <input
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1.5 h-10 w-full rounded-lg border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring/40"
                placeholder="you@example.com"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="mt-2 inline-flex h-10 items-center justify-center gap-2 rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground shadow-sm transition hover:brightness-110 disabled:opacity-50"
            >
              {loading && <Loader2 className="size-4 animate-spin" />}
              Send reset link
            </button>
          </form>
        )}
        <div className="mt-6 flex items-center justify-center gap-2 border-t border-border pt-4">
          <img src="/sio.png" alt="Siohioma" className="h-5 w-auto object-contain" />
          <span className="text-xs text-muted-foreground">Powered by</span>
          <span className="text-xs font-medium text-foreground">Siohioma</span>
        </div>
      </div>
    </div>
  );
}
