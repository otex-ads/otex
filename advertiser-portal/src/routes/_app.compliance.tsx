import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import { Shield, CheckCircle, AlertTriangle } from "lucide-react";

export const Route = createFileRoute("/_app/compliance")({
  component: CompliancePage,
});

function CompliancePage() {
  const [gdprEnabled, setGdprEnabled] = useState(true);
  const [ccpaEnabled, setCcpaEnabled] = useState(true);
  const [adVerification, setAdVerification] = useState(true);
  const [consentGiven, setConsentGiven] = useState(false);
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // TODO: Save compliance configuration to API
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Compliance Settings</h1>
        <p className="text-muted-foreground">Configure GDPR, CCPA, and ad verification settings</p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="size-5" />
              GDPR
            </CardTitle>
            <CardDescription>General Data Protection Regulation</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">EU data protection compliance</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle className="size-5" />
              CCPA
            </CardTitle>
            <CardDescription>California Consumer Privacy Act</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">California privacy law compliance</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="size-5" />
              Ad Verification
            </CardTitle>
            <CardDescription>Ad fraud prevention</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Verify ad delivery and prevent fraud</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>GDPR Compliance</CardTitle>
          <CardDescription>Configure GDPR consent management</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Enable GDPR Consent</Label>
              <p className="text-sm text-muted-foreground">
                Require user consent for data processing
              </p>
            </div>
            <Switch checked={gdprEnabled} onCheckedChange={setGdprEnabled} />
          </div>

          {gdprEnabled && (
            <div className="space-y-3 border-l-2 border-border pl-4">
              <div className="flex items-center space-x-2">
                <Checkbox id="consentDisplay" defaultChecked />
                <Label htmlFor="consentDisplay">Display consent banner</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="consentGranular" defaultChecked />
                <Label htmlFor="consentGranular">Granular consent options</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="consentRecord" defaultChecked />
                <Label htmlFor="consentRecord">Record consent timestamps</Label>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>CCPA Compliance</CardTitle>
          <CardDescription>Configure CCPA opt-out settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Enable CCPA Opt-Out</Label>
              <p className="text-sm text-muted-foreground">
                Allow California users to opt out of data sale
              </p>
            </div>
            <Switch checked={ccpaEnabled} onCheckedChange={setCcpaEnabled} />
          </div>

          {ccpaEnabled && (
            <div className="space-y-3 border-l-2 border-border pl-4">
              <div className="flex items-center space-x-2">
                <Checkbox id="ccpaLink" defaultChecked />
                <Label htmlFor="ccpaLink">Display "Do Not Sell" link</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="ccpaRecord" defaultChecked />
                <Label htmlFor="ccpaRecord">Record opt-out requests</Label>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Ad Verification</CardTitle>
          <CardDescription>Configure ad fraud prevention settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Enable Ad Verification</Label>
              <p className="text-sm text-muted-foreground">Verify ad delivery and prevent fraud</p>
            </div>
            <Switch checked={adVerification} onCheckedChange={setAdVerification} />
          </div>

          {adVerification && (
            <div className="space-y-3 border-l-2 border-border pl-4">
              <div className="flex items-center space-x-2">
                <Checkbox id="verifyViewability" defaultChecked />
                <Label htmlFor="verifyViewability">Verify viewability</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="verifyDomain" defaultChecked />
                <Label htmlFor="verifyDomain">Verify domain safety</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="verifyBot" defaultChecked />
                <Label htmlFor="verifyBot">Detect bot traffic</Label>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Button onClick={handleSave} className="w-full">
        {saved ? (
          <>
            <Shield className="mr-2 size-4" />
            Saved
          </>
        ) : (
          "Save Compliance Settings"
        )}
      </Button>

      {saved && (
        <div className="flex items-center gap-2 text-sm text-green-600">
          <Shield className="size-4" />
          Compliance settings saved successfully
        </div>
      )}
    </div>
  );
}
