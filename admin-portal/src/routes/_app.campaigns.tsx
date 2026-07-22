import { createFileRoute } from "@tanstack/react-router";
import { SurfaceCard } from "@/components/Card";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";
import { KES } from "@/lib/format";

export const Route = createFileRoute("/_app/campaigns")({
  component: CampaignsPage,
});

function CampaignsPage() {
  const { data: campaigns, isLoading } = useQuery({
    queryKey: ["admin-campaigns"],
    queryFn: () => api.fetch("/admin/campaigns"),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Campaigns</h1>
        <p className="mt-2 text-sm text-muted-foreground">Manage all advertising campaigns</p>
      </div>
      <SurfaceCard className="p-0">
        {isLoading ? (
          <div className="p-8 text-center text-sm text-muted-foreground">Loading...</div>
        ) : campaigns && campaigns.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Campaign</th>
                  <th className="px-6 py-3 font-medium">Advertiser ID</th>
                  <th className="px-6 py-3 font-medium">Model</th>
                  <th className="px-6 py-3 font-medium">Bid</th>
                  <th className="px-6 py-3 font-medium">Daily Budget</th>
                  <th className="px-6 py-3 font-medium">Total Budget</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {campaigns.map((campaign: any) => (
                  <tr key={campaign.id} className="border-b border-border hover:bg-muted/30">
                    <td className="px-6 py-4 font-medium">{campaign.name}</td>
                    <td className="px-6 py-4 text-muted-foreground">{campaign.advertiser_id.slice(0, 8)}...</td>
                    <td className="px-6 py-4 uppercase">{campaign.pricing_model}</td>
                    <td className="px-6 py-4">{KES(campaign.bid_amount_cents)}</td>
                    <td className="px-6 py-4">{KES(campaign.daily_budget_cents)}</td>
                    <td className="px-6 py-4">{KES(campaign.total_budget_cents)}</td>
                    <td className="px-6 py-4">
                      <StatusPill status={campaign.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-sm text-muted-foreground">No campaigns found</div>
        )}
      </SurfaceCard>
    </div>
  );
}
