import { createFileRoute, redirect, useNavigate, Link } from "@tanstack/react-router";
import { useState } from "react";
import { motion } from "framer-motion";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { api, getToken, setToken, setUser } from "@/lib/api";

export const Route = createFileRoute("/auth/login")({
  ssr: false,
  beforeLoad: () => {
    if (getToken()) throw redirect({ to: "/" });
  },
  component: LoginPage,
});

function LoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      toast.error("Enter your email and password.");
      return;
    }
    setLoading(true);
    try {
      const res = await api.login(email.trim(), password);
      if (!res?.access_token) throw new Error("No token returned");
      setToken(res.access_token);
      setUser({ email: res.email ?? email.trim() });
      toast.success("Signed in");
      navigate({ to: "/" });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Sign in failed";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid min-h-screen place-items-center bg-background p-4">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
        className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-xl"
      >
        <div className="mb-6 flex items-center gap-2.5">
          <img src="/logo.png" alt="OtexAds" className="size-10 object-contain" />
          <div className="leading-tight">
            <div className="text-base font-bold text-foreground">OtexAds</div>
            <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Publisher Portal</div>
          </div>
        </div>
        <h1 className="display text-3xl font-normal tracking-tight text-foreground">Welcome back.</h1>
        <p className="mt-1 text-sm text-muted-foreground">Sign in to monetize your traffic.</p>

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
          <div>
            <label className="text-xs font-medium text-foreground">Password</label>
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
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
            Sign in
          </button>
          <p className="mt-2 text-center text-xs text-muted-foreground">
            No account yet? <Link to="/auth/register" className="text-primary hover:underline">Create account</Link>
          </p>
        </form>
      </motion.div>
    </div>
  );
}
