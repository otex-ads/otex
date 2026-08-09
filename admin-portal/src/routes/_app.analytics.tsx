import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { SurfaceCard } from "@/components/Card";
import { api } from "@/lib/api";

export const Route = createFileRoute("/_app/analytics")({
  component: AnalyticsPage,
});

function AnalyticsPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ["admin-stats"],
    queryFn: () => api.getStats(),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Analytics</h1>
        <p className="mt-2 text-sm text-muted-foreground">Platform performance metrics</p>
      </div>

      {isLoading ? (
        <SurfaceCard className="p-8">
          <div className="text-center text-sm text-muted-foreground">Loading...</div>
        </SurfaceCard>
      ) : stats ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <SurfaceCard className="p-6">
            <div className="text-sm font-medium text-muted-foreground">Total Users</div>
            <div className="mt-2 text-3xl font-bold text-foreground">{stats.total_users || 0}</div>
            <div className="mt-1 text-xs text-muted-foreground">
              {stats.total_advertisers || 0} advertisers · {stats.total_publishers || 0} publishers
            </div>
          </SurfaceCard>
          <SurfaceCard className="p-6">
            <div className="text-sm font-medium text-muted-foreground">Active Campaigns</div>
            <div className="mt-2 text-3xl font-bold text-foreground">{stats.active_campaigns || 0}</div>
            <div className="mt-1 text-xs text-muted-foreground">
              of {stats.total_campaigns || 0} total campaigns
            </div>
          </SurfaceCard>
          <SurfaceCard className="p-6">
            <div className="text-sm font-medium text-muted-foreground">Total Revenue</div>
            <div className="mt-2 text-3xl font-bold text-foreground">
              KES {((stats.total_revenue_cents || 0) / 100).toLocaleString()}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">Lifetime platform revenue</div>
          </SurfaceCard>
          <SurfaceCard className="p-6">
            <div className="text-sm font-medium text-muted-foreground">Platform Status</div>
            <div className="mt-2 text-3xl font-bold text-green-600">Active</div>
            <div className="mt-1 text-xs text-muted-foreground">All systems operational</div>
          </SurfaceCard>
        </div>
      ) : (
        <SurfaceCard className="p-8">
          <div className="text-center text-sm text-muted-foreground">
            Failed to load analytics data.
          </div>
        </SurfaceCard>
      )}
    </div>
  );
}
