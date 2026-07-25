import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { KeyRound, Plus, Trash2, Copy } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { FormField, inputCls, selectCls } from "@/components/Modal";
import { getUser, setUser } from "@/lib/api";

export const Route = createFileRoute("/_app/settings")({
  component: SettingsPage,
});

const TABS = ["Profile", "Security", "Notifications", "API keys"] as const;
type Tab = (typeof TABS)[number];

interface ApiKey {
  id: string;
  label: string;
  token: string;
  createdAt: string;
}

function SettingsPage() {
  const [tab, setTab] = useState<Tab>("Profile");
  const existing = getUser();

  const [profile, setProfile] = useState({
    name: existing?.name ?? "",
    email: existing?.email ?? "",
    phone: "",
    company: "",
  });
  const [pwd, setPwd] = useState({ current: "", next: "", confirm: "" });
  const [notif, setNotif] = useState({
    dailyReports: true,
    payoutAlerts: true,
    weeklyDigest: false,
  });
  const [twoFA, setTwoFA] = useState(false);
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [newKeyLabel, setNewKeyLabel] = useState("");

  const saveProfile = () => {
    if (!profile.email) return toast.error("Email required.");
    setUser({ email: profile.email, name: profile.name });
    toast.success("Profile updated");
  };

  const changePwd = () => {
    if (!pwd.current || !pwd.next) return toast.error("Fill all password fields.");
    if (pwd.next !== pwd.confirm) return toast.error("New passwords don't match.");
    if (pwd.next.length < 8) return toast.error("Password must be at least 8 characters.");
    toast.success("Password updated");
    setPwd({ current: "", next: "", confirm: "" });
  };

  const generateKey = () => {
    if (!newKeyLabel.trim()) return toast.error("Give the key a label.");
    const token =
      "pk_live_" + Math.random().toString(36).slice(2) + Math.random().toString(36).slice(2);
    setKeys((k) => [
      {
        id: Math.random().toString(36).slice(2, 8),
        label: newKeyLabel,
        token,
        createdAt: new Date().toISOString(),
      },
      ...k,
    ]);
    setNewKeyLabel("");
    toast.success("API key generated");
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Settings" description="Account, security, notifications and API access." />

      <div className="flex flex-wrap gap-1 rounded-full border border-border bg-muted/40 p-1 w-fit">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-full px-4 py-1.5 text-xs font-medium ${tab === t ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === "Profile" && (
        <SurfaceCard>
          <SectionLabel>Profile</SectionLabel>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <FormField label="Full name">
              <input
                className={inputCls}
                value={profile.name}
                onChange={(e) => setProfile({ ...profile, name: e.target.value })}
              />
            </FormField>
            <FormField label="Email">
              <input
                type="email"
                className={inputCls}
                value={profile.email}
                onChange={(e) => setProfile({ ...profile, email: e.target.value })}
              />
            </FormField>
            <FormField label="Phone">
              <input
                className={inputCls}
                value={profile.phone}
                onChange={(e) => setProfile({ ...profile, phone: e.target.value })}
                placeholder="+1 555 0100"
              />
            </FormField>
            <FormField label="Company">
              <input
                className={inputCls}
                value={profile.company}
                onChange={(e) => setProfile({ ...profile, company: e.target.value })}
              />
            </FormField>
          </div>
          <div className="mt-6 flex justify-end">
            <button
              onClick={saveProfile}
              className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
            >
              Save changes
            </button>
          </div>
        </SurfaceCard>
      )}

      {tab === "Security" && (
        <div className="space-y-6">
          <SurfaceCard>
            <SectionLabel>Change password</SectionLabel>
            <div className="mt-4 grid gap-4 sm:grid-cols-3">
              <FormField label="Current">
                <input
                  type="password"
                  className={inputCls}
                  value={pwd.current}
                  onChange={(e) => setPwd({ ...pwd, current: e.target.value })}
                />
              </FormField>
              <FormField label="New">
                <input
                  type="password"
                  className={inputCls}
                  value={pwd.next}
                  onChange={(e) => setPwd({ ...pwd, next: e.target.value })}
                />
              </FormField>
              <FormField label="Confirm">
                <input
                  type="password"
                  className={inputCls}
                  value={pwd.confirm}
                  onChange={(e) => setPwd({ ...pwd, confirm: e.target.value })}
                />
              </FormField>
            </div>
            <div className="mt-6 flex justify-end">
              <button
                onClick={changePwd}
                className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
              >
                Update password
              </button>
            </div>
          </SurfaceCard>
          <SurfaceCard>
            <div className="flex items-center justify-between">
              <div>
                <SectionLabel>Two-factor authentication</SectionLabel>
                <p className="mt-1 text-sm text-muted-foreground">
                  Require a one-time code at sign-in.
                </p>
              </div>
              <label className="inline-flex cursor-pointer items-center gap-2">
                <input
                  type="checkbox"
                  checked={twoFA}
                  onChange={(e) => {
                    setTwoFA(e.target.checked);
                    toast.success(e.target.checked ? "2FA enabled" : "2FA disabled");
                  }}
                  className="peer sr-only"
                />
                <span className="relative inline-block h-6 w-11 rounded-full bg-muted transition peer-checked:bg-primary">
                  <span className="absolute left-0.5 top-0.5 size-5 rounded-full bg-white shadow transition peer-checked:translate-x-5" />
                </span>
              </label>
            </div>
          </SurfaceCard>
        </div>
      )}

      {tab === "Notifications" && (
        <SurfaceCard>
          <SectionLabel>Email notifications</SectionLabel>
          <div className="mt-4 space-y-3">
            {(
              [
                ["dailyReports", "Daily performance report"],
                ["payoutAlerts", "Payout status alerts"],
                ["weeklyDigest", "Weekly digest"],
              ] as const
            ).map(([k, label]) => (
              <label
                key={k}
                className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 text-sm"
              >
                <span className="font-medium text-foreground">{label}</span>
                <input
                  type="checkbox"
                  checked={notif[k]}
                  onChange={(e) => setNotif({ ...notif, [k]: e.target.checked })}
                  className="size-4 accent-primary"
                />
              </label>
            ))}
          </div>
          <div className="mt-6 flex justify-end">
            <button
              onClick={() => toast.success("Preferences saved")}
              className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
            >
              Save preferences
            </button>
          </div>
        </SurfaceCard>
      )}

      {tab === "API keys" && (
        <SurfaceCard>
          <div className="flex items-start justify-between gap-3">
            <div>
              <SectionLabel>API keys</SectionLabel>
              <p className="mt-1 text-sm text-muted-foreground">
                Use these keys to access the Publisher API programmatically.
              </p>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap items-end gap-2">
            <div className="flex-1 min-w-[220px]">
              <FormField label="Label">
                <input
                  className={inputCls}
                  value={newKeyLabel}
                  onChange={(e) => setNewKeyLabel(e.target.value)}
                  placeholder="Production server"
                />
              </FormField>
            </div>
            <button
              onClick={generateKey}
              className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
            >
              <Plus className="size-4" /> Generate
            </button>
          </div>
          <div className="mt-6 space-y-2">
            {keys.length === 0 && (
              <div className="rounded-lg border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
                <KeyRound className="mx-auto mb-2 size-4" /> No API keys yet
              </div>
            )}
            {keys.map((k) => (
              <div
                key={k.id}
                className="flex items-center justify-between gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm"
              >
                <div className="min-w-0 flex-1">
                  <div className="font-medium text-foreground">{k.label}</div>
                  <div className="font-mono truncate text-xs text-muted-foreground">{k.token}</div>
                </div>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(k.token);
                    toast.success("Copied");
                  }}
                  className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                >
                  <Copy className="size-3.5" />
                </button>
                <button
                  onClick={() => {
                    setKeys((ks) => ks.filter((x) => x.id !== k.id));
                    toast.success("Key revoked");
                  }}
                  className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-danger-soft hover:text-danger"
                >
                  <Trash2 className="size-3.5" />
                </button>
              </div>
            ))}
          </div>
        </SurfaceCard>
      )}
    </div>
  );
}
