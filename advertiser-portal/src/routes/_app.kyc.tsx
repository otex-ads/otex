import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FileText, Upload, CheckCircle, AlertCircle, Clock } from "lucide-react";

export const Route = createFileRoute("/_app/kyc")({
  component: KYCPage,
});

function KYCPage() {
  const [status, setStatus] = useState<"pending" | "submitted" | "approved" | "rejected">("pending");
  const [businessName, setBusinessName] = useState("");
  const [taxId, setTaxId] = useState("");
  const [documentType, setDocumentType] = useState("");
  const [saved, setSaved] = useState(false);

  const handleSubmit = () => {
    // TODO: Submit KYC documents to API
    setStatus("submitted");
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">KYC Verification</h1>
        <p className="text-muted-foreground">Complete identity verification to unlock full platform features</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {status === "pending" && <Clock className="size-5 text-yellow-500" />}
            {status === "submitted" && <AlertCircle className="size-5 text-blue-500" />}
            {status === "approved" && <CheckCircle className="size-5 text-green-500" />}
            {status === "rejected" && <AlertCircle className="size-5 text-red-500" />}
            Verification Status
          </CardTitle>
          <CardDescription>
            {status === "pending" && "Verification not started"}
            {status === "submitted" && "Documents submitted - under review"}
            {status === "approved" && "Verification complete"}
            {status === "rejected" && "Verification failed - please resubmit"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <div className={`h-2 flex-1 rounded-full ${
              status === "pending" ? "bg-muted" : 
              status === "submitted" ? "bg-blue-500" : 
              status === "approved" ? "bg-green-500" : "bg-red-500"
            }`} />
          </div>
        </CardContent>
      </Card>

      {status !== "approved" && (
        <Card>
          <CardHeader>
            <CardTitle>Business Information</CardTitle>
            <CardDescription>Provide your business details for verification</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="businessName">Business Name</Label>
              <Input
                id="businessName"
                value={businessName}
                onChange={(e) => setBusinessName(e.target.value)}
                placeholder="Your Business Ltd"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="taxId">Tax ID / VAT Number</Label>
              <Input
                id="taxId"
                value={taxId}
                onChange={(e) => setTaxId(e.target.value)}
                placeholder="Enter tax identification number"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="documentType">Document Type</Label>
              <Select value={documentType} onValueChange={setDocumentType}>
                <SelectTrigger id="documentType">
                  <SelectValue placeholder="Select document type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="passport">Passport</SelectItem>
                  <SelectItem value="drivers_license">Driver's License</SelectItem>
                  <SelectItem value="national_id">National ID</SelectItem>
                  <SelectItem value="business_license">Business License</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="document">Upload Document</Label>
              <div className="flex items-center gap-4">
                <Input
                  id="document"
                  type="file"
                  className="flex-1"
                />
                <Button variant="outline" size="icon">
                  <Upload className="size-4" />
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Accepted formats: PDF, JPG, PNG. Max file size: 5MB
              </p>
            </div>

            <Button onClick={handleSubmit} className="w-full">
              {saved ? (
                <>
                  <CheckCircle className="mr-2 size-4" />
                  Submitted
                </>
              ) : (
                "Submit for Verification"
              )}
            </Button>

            {saved && (
              <div className="flex items-center gap-2 text-sm text-green-600">
                <CheckCircle className="size-4" />
                Documents submitted successfully
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {status === "approved" && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-green-600">
              <CheckCircle className="size-5" />
              Verification Complete
            </CardTitle>
            <CardDescription>Your account has been fully verified</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm">
                <CheckCircle className="size-4 text-green-500" />
                <span>Business information verified</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <CheckCircle className="size-4 text-green-500" />
                <span>Documents approved</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <CheckCircle className="size-4 text-green-500" />
                <span>Full platform access enabled</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Why KYC?</CardTitle>
          <CardDescription>Benefits of completing verification</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="space-y-2 text-sm text-muted-foreground">
            <li className="flex items-start gap-2">
              <FileText className="size-4 mt-0.5" />
              <span>Higher spending limits</span>
            </li>
            <li className="flex items-start gap-2">
              <FileText className="size-4 mt-0.5" />
              <span>Premium ad placements</span>
            </li>
            <li className="flex items-start gap-2">
              <FileText className="size-4 mt-0.5" />
              <span>Faster payment processing</span>
            </li>
            <li className="flex items-start gap-2">
              <FileText className="size-4 mt-0.5" />
              <span>Priority support access</span>
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
