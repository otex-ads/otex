import { createFileRoute, Link } from "@tanstack/react-router";
import { Plus, Pause, Play, Trash2, Megaphone } from "lucide-react";
import { SurfaceCard } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { useStore, store } from "@/lib/store";
import { KES, Num, formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/campaigns")({
  component: CampaignsPage,
});

function CampaignsPage() {
  const campaigns = useStore((s) => s.campaigns) || [];

  return (
    <div>
      <PageHeader
        title="Campaigns"
        description="Create, pause, and monitor your ad campaigns across every format."
        actions={
          <Link
            to="/campaigns/new"
            className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:brightness-110"
          >
            <Plus className="size-4" /> New campaign
          </Link>
        }
      />

      {campaigns.length === 0 ? (
        <EmptyState
          icon={Megaphone}
          title="No campaigns yet"
          description="Launch your first campaign to start reaching users."
          action={
            <Link
              to="/campaigns/new"
              className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
            >
              <Plus className="size-4" /> Create campaign
            </Link>
          }
        />
      ) : (
        <SurfaceCard className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Campaign</th>
                  <th className="px-6 py-3 font-medium">Format</th>
                  <th className="px-6 py-3 font-medium">Model / Bid</th>
                  <th className="px-6 py-3 text-right font-medium">Daily / Total</th>
                  <th className="px-6 py-3 text-right font-medium">Spent</th>
                  <th className="px-6 py-3 text-right font-medium">Clicks · CTR</th>
                  <th className="px-6 py-3 font-medium">Targeting</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 text-right font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {campaigns.map((c) => {
                  const ctr = c.impressions ? (c.clicks / c.impressions) * 100 : 0;
                  return (
                    <tr
                      key={c.id}
                      className="border-b border-border last:border-0 hover:bg-muted/20"
                    >
                      <td className="px-6 py-4">
                        <div className="font-medium text-foreground">{c.name}</div>
                        <div className="text-xs text-muted-foreground">
                          Created {formatDate(c.createdAt)}
                        </div>
                      </td>
                      <td className="px-6 py-4 capitalize text-muted-foreground">{c.format}</td>
                      <td className="px-6 py-4">
                        <div className="text-foreground">{c.pricingModel}</div>
                        <div className="num text-xs text-muted-foreground">
                          {KES(c.bid)} /{" "}
                          {c.pricingModel === "CPM"
                            ? "1K impr"
                            : c.pricingModel === "CPC"
                              ? "click"
                              : "conv"}
                        </div>
                      </td>
                      <td className="num px-6 py-4 text-right text-muted-foreground">
                        {KES(c.dailyBudget)} <span className="text-muted-foreground/60">/</span>{" "}
                        {KES(c.totalBudget)}
                      </td>
                      <td className="num px-6 py-4 text-right font-medium">{KES(c.spent)}</td>
                      <td className="num px-6 py-4 text-right">
                        <div>{Num(c.clicks)}</div>
                        <div className="text-xs text-muted-foreground">{ctr.toFixed(2)}%</div>
                      </td>
                      <td className="px-6 py-4 text-xs text-muted-foreground">
                        {c.targeting.countries.join(", ")} · {c.targeting.devices.join("/")}
                      </td>
                      <td className="px-6 py-4">
                        <StatusPill
                          status={c.status[0].toUpperCase() + c.status.slice(1)}
                          tone={
                            c.status === "active"
                              ? "success"
                              : c.status === "paused"
                                ? "warning"
                                : "neutral"
                          }
                        />
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex justify-end gap-1">
                          {c.status === "active" ? (
                            <button
                              onClick={() => store.updateCampaign(c.id, { status: "paused" })}
                              className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted"
                              title="Pause"
                            >
                              <Pause className="size-3.5" />
                            </button>
                          ) : (
                            <button
                              onClick={() => store.updateCampaign(c.id, { status: "active" })}
                              className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted"
                              title="Resume"
                            >
                              <Play className="size-3.5" />
                            </button>
                          )}
                          <button
                            onClick={() => {
                              if (confirm(`Delete "${c.name}"?`)) store.removeCampaign(c.id);
                            }}
                            className="grid size-8 place-items-center rounded-md text-danger hover:bg-danger-soft"
                            title="Delete"
                          >
                            <Trash2 className="size-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </SurfaceCard>
      )}
    </div>
  );
}
