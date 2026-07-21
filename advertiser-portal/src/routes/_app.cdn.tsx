import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Upload, Check, AlertCircle, Globe, Cloud } from "lucide-react";

export const Route = createFileRoute("/_app/cdn")({
  component: CDNPage,
});

function CDNPage() {
  const [provider, setProvider] = useState<"cloudflare" | "aws" | "local">("cloudflare");
  const [apiKey, setApiKey] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [bucket, setBucket] = useState("");
  const [region, setRegion] = useState("");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // TODO: Save CDN configuration to API
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">CDN Management</h1>
        <p className="text-muted-foreground">Configure content delivery network for ad assets</p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Cloud className="size-5" />
              Cloudflare R2
            </CardTitle>
            <CardDescription>Low-cost object storage with global CDN</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Zero egress fees, S3-compatible API</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="size-5" />
              AWS S3
            </CardTitle>
            <CardDescription>Industry-standard object storage</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Full feature set, global edge locations</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Upload className="size-5" />
              Local Storage
            </CardTitle>
            <CardDescription>Store assets on your server</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">No external dependencies, simple setup</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>CDN Configuration</CardTitle>
          <CardDescription>Enter your CDN provider credentials</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="provider">Provider</Label>
            <Select value={provider} onValueChange={(v: "cloudflare" | "aws" | "local") => setProvider(v)}>
              <SelectTrigger id="provider">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="cloudflare">Cloudflare R2</SelectItem>
                <SelectItem value="aws">AWS S3</SelectItem>
                <SelectItem value="local">Local Storage</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {provider !== "local" && (
            <>
              <div className="space-y-2">
                <Label htmlFor="apiKey">API Key / Access Key</Label>
                <Input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter your API key"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="secretKey">Secret Key</Label>
                <Input
                  id="secretKey"
                  type="password"
                  value={secretKey}
                  onChange={(e) => setSecretKey(e.target.value)}
                  placeholder="Enter your secret key"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="bucket">Bucket Name</Label>
                <Input
                  id="bucket"
                  value={bucket}
                  onChange={(e) => setBucket(e.target.value)}
                  placeholder="my-ad-assets"
                />
              </div>

              {provider === "aws" && (
                <div className="space-y-2">
                  <Label htmlFor="region">Region</Label>
                  <Input
                    id="region"
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    placeholder="us-east-1"
                  />
                </div>
              )}
            </>
          )}

          <div className="flex items-center gap-2">
            <Button onClick={handleSave} className="w-full">
              {saved ? (
                <>
                  <Check className="mr-2 size-4" />
                  Saved
                </>
              ) : (
                "Save Configuration"
              )}
            </Button>
          </div>

          {saved && (
            <div className="flex items-center gap-2 text-sm text-green-600">
              <Check className="size-4" />
              CDN configuration saved successfully
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>CDN Status</CardTitle>
          <CardDescription>Current CDN configuration status</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-sm">
            <AlertCircle className="size-4 text-yellow-500" />
            <span className="text-muted-foreground">No CDN configured yet</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
