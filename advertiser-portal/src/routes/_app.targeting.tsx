import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Target, Users, Globe, Brain, TrendingUp } from "lucide-react";

export const Route = createFileRoute("/_app/targeting")({
  component: TargetingPage,
});

function TargetingPage() {
  const [behavioral, setBehavioral] = useState(false);
  const [contextual, setContextual] = useState(false);
  const [lookalike, setLookalike] = useState(false);
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // TODO: Save targeting configuration to API
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Advanced Targeting</h1>
        <p className="text-muted-foreground">Configure advanced targeting options for your campaigns</p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="size-5" />
              Behavioral
            </CardTitle>
            <CardDescription>Target based on user behavior</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Past interactions, browsing history, purchase intent</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="size-5" />
              Contextual
            </CardTitle>
            <CardDescription>Target based on content context</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Page content, keywords, categories, topics</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Brain className="size-5" />
              Lookalike
            </CardTitle>
            <CardDescription>Find users similar to your best customers</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">AI-powered audience expansion</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Targeting Configuration</CardTitle>
          <CardDescription>Select targeting methods to enable</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="behavioral"
                checked={behavioral}
                onCheckedChange={(checked) => setBehavioral(checked as boolean)}
              />
              <Label htmlFor="behavioral" className="flex items-center gap-2">
                <Users className="size-4" />
                Behavioral Targeting
              </Label>
            </div>
            {behavioral && (
              <div className="ml-6 space-y-3 border-l-2 border-border pl-4">
                <div className="space-y-2">
                  <Label>Interest Categories</Label>
                  <Select>
                    <SelectTrigger>
                      <SelectValue placeholder="Select interests" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="tech">Technology</SelectItem>
                      <SelectItem value="finance">Finance</SelectItem>
                      <SelectItem value="health">Health & Wellness</SelectItem>
                      <SelectItem value="shopping">Shopping</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Purchase Intent</Label>
                  <Select>
                    <SelectTrigger>
                      <SelectValue placeholder="Select intent level" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="high">High</SelectItem>
                      <SelectItem value="medium">Medium</SelectItem>
                      <SelectItem value="low">Low</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="contextual"
                checked={contextual}
                onCheckedChange={(checked) => setContextual(checked as boolean)}
              />
              <Label htmlFor="contextual" className="flex items-center gap-2">
                <Globe className="size-4" />
                Contextual Targeting
              </Label>
            </div>
            {contextual && (
              <div className="ml-6 space-y-3 border-l-2 border-border pl-4">
                <div className="space-y-2">
                  <Label>Keywords</Label>
                  <Input placeholder="Enter keywords (comma separated)" />
                </div>
                <div className="space-y-2">
                  <Label>Content Categories</Label>
                  <Select>
                    <SelectTrigger>
                      <SelectValue placeholder="Select categories" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="news">News</SelectItem>
                      <SelectItem value="entertainment">Entertainment</SelectItem>
                      <SelectItem value="sports">Sports</SelectItem>
                      <SelectItem value="business">Business</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="lookalike"
                checked={lookalike}
                onCheckedChange={(checked) => setLookalike(checked as boolean)}
              />
              <Label htmlFor="lookalike" className="flex items-center gap-2">
                <Brain className="size-4" />
                Lookalike Audiences
              </Label>
            </div>
            {lookalike && (
              <div className="ml-6 space-y-3 border-l-2 border-border pl-4">
                <div className="space-y-2">
                  <Label>Source Audience</Label>
                  <Select>
                    <SelectTrigger>
                      <SelectValue placeholder="Select source audience" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="customers">Existing Customers</SelectItem>
                      <SelectItem value="converters">Past Converters</SelectItem>
                      <SelectItem value="engaged">Highly Engaged Users</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Similarity Threshold</Label>
                  <Select>
                    <SelectTrigger>
                      <SelectValue placeholder="Select threshold" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="high">High (1-3% similarity)</SelectItem>
                      <SelectItem value="medium">Medium (3-5% similarity)</SelectItem>
                      <SelectItem value="low">Low (5-10% similarity)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}
          </div>

          <Button onClick={handleSave} className="w-full">
            {saved ? (
              <>
                <Target className="mr-2 size-4" />
                Saved
              </>
            ) : (
              "Save Targeting Configuration"
            )}
          </Button>

          {saved && (
            <div className="flex items-center gap-2 text-sm text-green-600">
              <Target className="size-4" />
              Targeting configuration saved successfully
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
