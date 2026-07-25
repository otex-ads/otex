import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { BarChart3, Zap, DollarSign, Settings } from "lucide-react";

export const Route = createFileRoute("/_app/rtb")({
  component: RTBPage,
});

function RTBPage() {
  const [enabled, setEnabled] = useState(false);
  const [floorPrice, setFloorPrice] = useState("0.01");
  const [bidStrategy, setBidStrategy] = useState("auto");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // TODO: Save RTB configuration to API
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">RTB/SSP Settings</h1>
        <p className="text-muted-foreground">
          Configure real-time bidding and supply-side platform integration
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="size-5" />
              Real-Time Bidding
            </CardTitle>
            <CardDescription>Auction-based ad delivery</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Maximize revenue through real-time auctions
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="size-5" />
              Floor Price
            </CardTitle>
            <CardDescription>Minimum bid acceptance</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Set minimum CPM for inventory</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="size-5" />
              Bid Strategy
            </CardTitle>
            <CardDescription>Optimize bidding behavior</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Auto or manual bid adjustments</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>RTB Configuration</CardTitle>
          <CardDescription>Configure real-time bidding settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Enable RTB</Label>
              <p className="text-sm text-muted-foreground">
                Allow real-time bidding for your campaigns
              </p>
            </div>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>

          {enabled && (
            <>
              <div className="space-y-2">
                <Label htmlFor="floorPrice">Floor Price (CPM)</Label>
                <Input
                  id="floorPrice"
                  type="number"
                  step="0.01"
                  value={floorPrice}
                  onChange={(e) => setFloorPrice(e.target.value)}
                  placeholder="0.01"
                />
                <p className="text-xs text-muted-foreground">
                  Minimum bid price per 1000 impressions
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="bidStrategy">Bid Strategy</Label>
                <Select value={bidStrategy} onValueChange={setBidStrategy}>
                  <SelectTrigger id="bidStrategy">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="auto">Automatic</SelectItem>
                    <SelectItem value="manual">Manual</SelectItem>
                    <SelectItem value="aggressive">Aggressive</SelectItem>
                    <SelectItem value="conservative">Conservative</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="sspEndpoint">SSP Endpoint</Label>
                <Input id="sspEndpoint" placeholder="https://ssp.example.com/bid" />
              </div>

              <div className="space-y-2">
                <Label htmlFor="sspApiKey">SSP API Key</Label>
                <Input id="sspApiKey" type="password" placeholder="Enter SSP API key" />
              </div>
            </>
          )}

          <Button onClick={handleSave} className="w-full" disabled={!enabled}>
            {saved ? (
              <>
                <Settings className="mr-2 size-4" />
                Saved
              </>
            ) : (
              "Save RTB Configuration"
            )}
          </Button>

          {saved && (
            <div className="flex items-center gap-2 text-sm text-green-600">
              <Settings className="size-4" />
              RTB configuration saved successfully
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>RTB Performance</CardTitle>
          <CardDescription>Real-time bidding performance metrics</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div>
              <p className="text-sm text-muted-foreground">Win Rate</p>
              <p className="text-2xl font-bold text-foreground">--</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Avg CPM</p>
              <p className="text-2xl font-bold text-foreground">--</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Impressions</p>
              <p className="text-2xl font-bold text-foreground">--</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Revenue</p>
              <p className="text-2xl font-bold text-foreground">--</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
