import { createFileRoute } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { Users, TrendingUp, DollarSign, Activity, ArrowUpRight } from "lucide-react";
import { SurfaceCard } from "@/components/Card";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { KES } from "@/lib/format";

export const Route = createFileRoute("/_app/")({
  component: Dashboard,
});

function StatCard({ label, value, sub, icon: Icon, delay = 0 }: {
  label: string; value: string; sub?: string; icon: typeof Users; delay?: number;
}) {
  return (
    <SurfaceCard delay={delay} className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="label-eyebrow">{label}</div>
        <div className="grid size-8 place-items-center rounded-lg bg-muted text-muted-foreground">
          <Icon className="size-4" />
        </div>
      </div>
      <div>
        <div className="num display text-4xl font-bold tracking-tight text-foreground">{value}</div>
        {sub && <div className="mt-2 text-xs text-muted-foreground">{sub}</div>}
      </div>
    </SurfaceCard>
  );
}

function Dashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ["admin-stats"],
    queryFn: () => api.fetch("/admin/stats"),
  });

  if (isLoading) {
    return (
      <div className="space-y-8">
        <motion.section
          initial={{ opacity: 0, y: -6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
          className="card-anchor relative overflow-hidden p-8 md:p-10"
        >
          <div className="relative">
            <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-anchor-foreground/60">
              <span>Admin</span><span className="text-anchor-foreground/30">/</span><span className="text-anchor-foreground/90">Dashboard</span>
            </div>
            <h1 className="display mt-4 text-5xl font-bold tracking-tight md:text-6xl">
              Platform <span className="text-anchor-foreground/60">Overview</span>
            </h1>
          </div>
        </motion.section>
        <div className="text-center text-sm text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <motion.section
        initial={{ opacity: 0, y: -6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
        className="card-anchor relative overflow-hidden p-8 md:p-10"
      >
        <div className="absolute -right-24 -top-24 size-72 rounded-full bg-white/[0.04] blur-3xl" />
        <div className="absolute -bottom-24 left-1/3 size-80 rounded-full bg-white/[0.03] blur-3xl" />
        <div className="relative">
          <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-anchor-foreground/60">
            <span>Admin</span><span className="text-anchor-foreground/30">/</span><span className="text-anchor-foreground/90">Dashboard</span>
          </div>
          <h1 className="display mt-4 text-5xl font-bold tracking-tight md:text-6xl">
            Platform <span className="text-anchor-foreground/60">Overview</span>
          </h1>
          <p className="mt-3 max-w-xl text-sm text-anchor-foreground/60">
            Monitor users, revenue, and platform performance across OtexAds.
          </p>
        </div>
      </motion.section>

      <div className="grid gap-4 lg:grid-cols-4">
        <StatCard delay={0.04} label="Total Users" value={stats?.total_users?.toString() || "0"} sub={`${stats?.total_advertisers || 0} advertisers · ${stats?.total_publishers || 0} publishers`} icon={Users} />
        <StatCard delay={0.08} label="Active Campaigns" value={stats?.active_campaigns?.toString() || "0"} sub={`${stats?.total_campaigns || 0} total campaigns`} icon={Activity} />
        <StatCard delay={0.12} label="Total Revenue" value={KES(stats?.total_revenue_cents || 0)} sub="All time" icon={DollarSign} />
        <StatCard delay={0.16} label="Total Campaigns" value={stats?.total_campaigns?.toString() || "0"} sub="Across all advertisers" icon={TrendingUp} />
      </div>

      <SurfaceCard delay={0.2} className="p-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="label-eyebrow mb-3">Quick Actions</div>
            <div className="text-sm text-muted-foreground">Manage users, campaigns, and view detailed analytics.</div>
          </div>
          <div className="flex gap-2">
            <a href="/users" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              Users <ArrowUpRight className="size-3" />
            </a>
            <a href="/campaigns" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              Campaigns <ArrowUpRight className="size-3" />
            </a>
            <a href="/revenue" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
              Revenue <ArrowUpRight className="size-3" />
            </a>
          </div>
        </div>
      </SurfaceCard>
    </div>
  );
}
