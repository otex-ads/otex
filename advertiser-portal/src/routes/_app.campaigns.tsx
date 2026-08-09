import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_app/campaigns")({
  component: CampaignsLayout,
});

function CampaignsLayout() {
  return <Outlet />;
}
