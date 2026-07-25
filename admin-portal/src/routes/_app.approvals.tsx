import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect } from "react";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { Button } from "@/components/ui/button";
import { Check, X, Globe, Megaphone, Image as ImageIcon } from "lucide-react";
import { getToken } from "@/lib/api";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/approvals")({
  component: ApprovalsPage,
});

function ApprovalsPage() {
  const [activeTab, setActiveTab] = useState<"sites" | "campaigns" | "creatives">("sites");
  const [pendingSites, setPendingSites] = useState<any[]>([]);
  const [pendingCampaigns, setPendingCampaigns] = useState<any[]>([]);
  const [pendingCreatives, setPendingCreatives] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPendingItems();
  }, []);

  const fetchPendingItems = async () => {
    setLoading(true);
    const token = getToken();
    try {
      const [sitesRes, campaignsRes, creativesRes] = await Promise.all([
        fetch("https://api.otexads.com/api/v1/admin/sites/pending", {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch("https://api.otexads.com/api/v1/admin/campaigns/pending", {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch("https://api.otexads.com/api/v1/admin/creatives/pending", {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ]);

      const [sitesData, campaignsData, creativesData] = await Promise.all([
        sitesRes.json(),
        campaignsRes.json(),
        creativesRes.json(),
      ]);

      // Unwrap the backend's { success: true, data: ... } envelope
      const sites = sitesData.success ? sitesData.data : sitesData;
      const campaigns = campaignsData.success ? campaignsData.data : campaignsData;
      const creatives = creativesData.success ? creativesData.data : creativesData;

      setPendingSites(Array.isArray(sites) ? sites : []);
      setPendingCampaigns(Array.isArray(campaigns) ? campaigns : []);
      setPendingCreatives(Array.isArray(creatives) ? creatives : []);
    } catch (error) {
      console.error("Failed to fetch pending items:", error);
      toast.error("Failed to load pending items");
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (type: string, id: string) => {
    const token = getToken();
    const endpoint = type === "sites" 
      ? "https://api.otexads.com/api/v1/admin/sites/moderate"
      : type === "campaigns"
      ? "https://api.otexads.com/api/v1/admin/campaigns/moderate"
      : "https://api.otexads.com/api/v1/admin/creatives/moderate";

    const body = type === "sites" 
      ? { site_id: id, action: "approve" }
      : type === "campaigns"
      ? { campaign_id: id, action: "approve" }
      : { creative_id: id, action: "approve" };

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      if (res.ok) {
        toast.success(`${type.slice(0, -1)} approved`);
        fetchPendingItems();
      } else {
        toast.error("Failed to approve");
      }
    } catch (error) {
      toast.error("Failed to approve");
    }
  };

  const handleReject = async (type: string, id: string) => {
    const token = getToken();
    const endpoint = type === "sites" 
      ? "https://api.otexads.com/api/v1/admin/sites/moderate"
      : type === "campaigns"
      ? "https://api.otexads.com/api/v1/admin/campaigns/moderate"
      : "https://api.otexads.com/api/v1/admin/creatives/moderate";

    const body = type === "sites" 
      ? { site_id: id, action: "reject" }
      : type === "campaigns"
      ? { campaign_id: id, action: "reject" }
      : { creative_id: id, action: "reject" };

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      if (res.ok) {
        toast.success(`${type.slice(0, -1)} rejected`);
        fetchPendingItems();
      } else {
        toast.error("Failed to reject");
      }
    } catch (error) {
      toast.error("Failed to reject");
    }
  };

  return (
    <div>
      <PageHeader
        title="Approvals"
        description="Review and approve pending sites, campaigns, and creatives."
      />

      <div className="mb-6 flex gap-2">
        <Button
          variant={activeTab === "sites" ? "default" : "outline"}
          onClick={() => setActiveTab("sites")}
        >
          <Globe className="mr-2 size-4" />
          Sites ({pendingSites.length})
        </Button>
        <Button
          variant={activeTab === "campaigns" ? "default" : "outline"}
          onClick={() => setActiveTab("campaigns")}
        >
          <Megaphone className="mr-2 size-4" />
          Campaigns ({pendingCampaigns.length})
        </Button>
        <Button
          variant={activeTab === "creatives" ? "default" : "outline"}
          onClick={() => setActiveTab("creatives")}
        >
          <ImageIcon className="mr-2 size-4" />
          Creatives ({pendingCreatives.length})
        </Button>
      </div>

      {loading ? (
        <div className="text-center text-muted-foreground">Loading...</div>
      ) : (
        <SurfaceCard className="p-0">
          {activeTab === "sites" && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                    <th className="px-6 py-3 font-medium">Domain</th>
                    <th className="px-6 py-3 font-medium">Publisher ID</th>
                    <th className="px-6 py-3 font-medium">Status</th>
                    <th className="px-6 py-3 font-medium">Created</th>
                    <th className="px-6 py-3 text-right font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingSites.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-6 py-8 text-center text-muted-foreground">
                        No pending sites
                      </td>
                    </tr>
                  ) : (
                    pendingSites.map((site) => (
                      <tr key={site.id} className="border-b border-border last:border-0 hover:bg-muted/20">
                        <td className="px-6 py-4 font-medium">{site.domain}</td>
                        <td className="px-6 py-4 text-muted-foreground">{site.publisher_id}</td>
                        <td className="px-6 py-4">
                          <StatusPill status={site.status} tone="warning" />
                        </td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {new Date(site.created_at).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 text-right">
                          <div className="flex justify-end gap-2">
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleApprove("sites", site.id)}
                            >
                              <Check className="size-4 text-green-600" />
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleReject("sites", site.id)}
                            >
                              <X className="size-4 text-red-600" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}

          {activeTab === "campaigns" && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                    <th className="px-6 py-3 font-medium">Name</th>
                    <th className="px-6 py-3 font-medium">Advertiser ID</th>
                    <th className="px-6 py-3 font-medium">Model</th>
                    <th className="px-6 py-3 font-medium">Bid</th>
                    <th className="px-6 py-3 font-medium">Status</th>
                    <th className="px-6 py-3 text-right font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingCampaigns.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="px-6 py-8 text-center text-muted-foreground">
                        No pending campaigns
                      </td>
                    </tr>
                  ) : (
                    pendingCampaigns.map((campaign) => (
                      <tr key={campaign.id} className="border-b border-border last:border-0 hover:bg-muted/20">
                        <td className="px-6 py-4 font-medium">{campaign.name}</td>
                        <td className="px-6 py-4 text-muted-foreground">{campaign.advertiser_id}</td>
                        <td className="px-6 py-4 capitalize">{campaign.pricing_model}</td>
                        <td className="px-6 py-4">KES {campaign.bid_amount_cents / 100}</td>
                        <td className="px-6 py-4">
                          <StatusPill status={campaign.status} tone="warning" />
                        </td>
                        <td className="px-6 py-4 text-right">
                          <div className="flex justify-end gap-2">
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleApprove("campaigns", campaign.id)}
                            >
                              <Check className="size-4 text-green-600" />
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleReject("campaigns", campaign.id)}
                            >
                              <X className="size-4 text-red-600" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}

          {activeTab === "creatives" && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                    <th className="px-6 py-3 font-medium">ID</th>
                    <th className="px-6 py-3 font-medium">Campaign ID</th>
                    <th className="px-6 py-3 font-medium">Type</th>
                    <th className="px-6 py-3 font-medium">Status</th>
                    <th className="px-6 py-3 font-medium">Created</th>
                    <th className="px-6 py-3 text-right font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingCreatives.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="px-6 py-8 text-center text-muted-foreground">
                        No pending creatives
                      </td>
                    </tr>
                  ) : (
                    pendingCreatives.map((creative) => (
                      <tr key={creative.id} className="border-b border-border last:border-0 hover:bg-muted/20">
                        <td className="px-6 py-4 font-medium">{creative.id}</td>
                        <td className="px-6 py-4 text-muted-foreground">{creative.campaign_id}</td>
                        <td className="px-6 py-4 capitalize">{creative.type}</td>
                        <td className="px-6 py-4">
                          <StatusPill status={creative.status} tone="warning" />
                        </td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {new Date(creative.created_at).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 text-right">
                          <div className="flex justify-end gap-2">
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleApprove("creatives", creative.id)}
                            >
                              <Check className="size-4 text-green-600" />
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleReject("creatives", creative.id)}
                            >
                              <X className="size-4 text-red-600" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}
        </SurfaceCard>
      )}
    </div>
  );
}
