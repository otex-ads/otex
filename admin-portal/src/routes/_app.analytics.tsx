import { createFileRoute } from "@tanstack/react-router";
import { SurfaceCard } from "@/components/Card";

export const Route = createFileRoute("/_app/analytics")({
  component: AnalyticsPage,
});

function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Analytics</h1>
        <p className="mt-2 text-sm text-muted-foreground">Platform performance metrics</p>
      </div>
      <SurfaceCard className="p-8">
        <div className="text-center text-sm text-muted-foreground">
          Analytics dashboard coming soon. View platform-wide metrics, trends, and performance data.
        </div>
      </SurfaceCard>
    </div>
  );
}
