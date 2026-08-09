import { createFileRoute, Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useState } from "react";
import { motion } from "framer-motion";
import { Loader2, ArrowLeft, CheckCircle2 } from "lucide-react";
import { toast } from "sonner";
import { api } from "@/lib/api";

export const Route = createFileRoute("/auth/reset-password")({
  ssr: false,
  component: ResetPasswordPage,
});

function ResetPasswordPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/auth/reset-password" });
  const token = search.token as string;
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPassword || !confirmPassword) {
      toast.error("Enter and confirm your new password.");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match.");
      return;
    }
    if (newPassword.length < 8) {
      toast.error("Password must be at least 8 characters.");
      return;
    }
    if (!token) {
      toast.error("Invalid reset token.");
      return;
    }
    setLoading(true);
    try {
      await api.resetPassword({ token, new_password: newPassword });
      setSuccess(true);
      toast.success("Password reset successfully.");
      setTimeout(() => {
        navigate({ to: "/auth/login" });
      }, 2000);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to reset password";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
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
          <h1 className="display text-3xl font-normal tracking-tight text-foreground">
            Invalid link
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            This password reset link is invalid or has expired.
          </p>
          <Link
            to="/auth/forgot-password"
            className="mt-6 inline-flex items-center gap-1 text-sm text-primary hover:underline"
          >
            Request a new reset link
          </Link>
        </div>
      </div>
    );
  }

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
        {success ? (
          <div className="py-8 text-center">
            <CheckCircle2 className="mx-auto mb-4 size-12 text-green-500" />
            <h1 className="display text-2xl font-normal tracking-tight text-foreground">
              Password reset
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Your password has been reset successfully. Redirecting to sign in...
            </p>
          </div>
        ) : (
          <>
            <h1 className="display text-3xl font-normal tracking-tight text-foreground">
              Reset password
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Enter your new password below.
            </p>

            <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-4">
              <div>
                <label className="text-xs font-medium text-foreground">New password</label>
                <input
                  type="password"
                  autoComplete="new-password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="mt-1.5 h-10 w-full rounded-lg border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring/40"
                  placeholder="••••••••"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-foreground">Confirm password</label>
                <input
                  type="password"
                  autoComplete="new-password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="mt-1.5 h-10 w-full rounded-lg border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring/40"
                  placeholder="••••••••"
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="mt-2 inline-flex h-10 items-center justify-center gap-2 rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground shadow-sm transition hover:brightness-110 disabled:opacity-50"
              >
                {loading && <Loader2 className="size-4 animate-spin" />}
                Reset password
              </button>
            </form>
          </>
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
