import { createFileRoute, Outlet, Link, useRouterState } from "@tanstack/react-router";
import { motion } from "framer-motion";
import {
  LayoutDashboard,
  Megaphone,
  ImageIcon,
  Wallet,
  BarChart3,
  Menu,
  X,
  Plus,
  Globe,
  Target,
  Shield,
  MessageSquare,
  FileText,
  LogOut,
} from "lucide-react";
import { useState } from "react";
import { useStore } from "@/lib/store";
import { KES } from "@/lib/format";
import { useInit } from "@/hooks/useInit";
import { clearToken } from "@/lib/api";

export const Route = createFileRoute("/_app")({
  component: AppLayout,
});

const menu = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard, exact: true },
  { to: "/campaigns", label: "Campaigns", icon: Megaphone },
  { to: "/creatives", label: "Creatives", icon: ImageIcon },
  { to: "/cdn", label: "CDN", icon: Globe },
  { to: "/targeting", label: "Targeting", icon: Target },
  { to: "/rtb", label: "RTB/SSP", icon: BarChart3 },
  { to: "/compliance", label: "Compliance", icon: Shield },
  { to: "/kyc", label: "KYC", icon: FileText },
  { to: "/support", label: "Support", icon: MessageSquare },
  { to: "/wallet", label: "Wallet", icon: Wallet },
  { to: "/stats", label: "Statistics", icon: BarChart3 },
] as const;

function NavItem({
  to,
  label,
  icon: Icon,
  exact,
  onClick,
}: {
  to: string;
  label: string;
  icon: typeof LayoutDashboard;
  exact?: boolean;
  onClick?: () => void;
}) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const active = exact ? pathname === to : pathname === to || pathname.startsWith(to + "/");
  return (
    <Link
      to={to}
      onClick={onClick}
      className={`group flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
        active
          ? "bg-primary text-primary-foreground"
          : "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
      }`}
    >
      <Icon className="size-4" />
      <span className="font-medium">{label}</span>
    </Link>
  );
}

function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const balance = useStore((s) => s.balance) || 0;
  useInit();

  const handleLogout = () => {
    clearToken();
    window.location.href = "/auth/login";
  };

  const Sidebar = ({ onNav }: { onNav?: () => void }) => (
    <>
      <Link to="/" onClick={onNav} className="mb-8 flex items-center gap-2.5">
        <img src="/logo.png" alt="OtexAds" className="size-9 object-contain" />
        <div className="leading-tight">
          <div className="text-sm font-bold text-foreground">OtexAds</div>
          <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
            Advertiser
          </div>
        </div>
      </Link>
      <div className="label-eyebrow mb-2 px-3">Workspace</div>
      <nav className="flex flex-col gap-0.5">
        {menu.map((m) => (
          <NavItem key={m.to} {...m} onClick={onNav} />
        ))}
      </nav>
      <div className="mt-auto flex flex-col gap-3">
        <div className="rounded-xl bg-muted/60 p-4 text-xs">
          <div className="label-eyebrow">Wallet balance</div>
          <div className="num mt-1 text-lg font-semibold text-foreground">{KES(balance)}</div>
          <Link
            to="/wallet"
            onClick={onNav}
            className="mt-2 inline-flex items-center gap-1 text-primary hover:underline"
          >
            <Plus className="size-3" /> Top up
          </Link>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
        >
          <LogOut className="size-4" />
          <span className="font-medium">Sign out</span>
        </button>
      </div>
    </>
  );

  return (
    <div className="flex min-h-screen bg-background">
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}
      {mobileOpen && (
        <aside className="fixed inset-y-0 left-0 z-50 flex w-[260px] flex-col gap-1 border-r border-border bg-sidebar p-5 lg:hidden">
          <button
            onClick={() => setMobileOpen(false)}
            className="absolute right-3 top-3 grid size-7 place-items-center rounded-md text-muted-foreground hover:bg-sidebar-accent"
          >
            <X className="size-3.5" />
          </button>
          <Sidebar onNav={() => setMobileOpen(false)} />
        </aside>
      )}

      <aside className="sticky top-0 hidden h-screen w-[260px] shrink-0 flex-col gap-1 border-r border-border bg-sidebar p-5 lg:flex">
        <Sidebar />
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex items-center justify-between gap-4 border-b border-border bg-background/80 px-4 py-4 backdrop-blur md:px-8 md:py-5">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setMobileOpen(true)}
              className="grid size-9 place-items-center rounded-md border border-border bg-card text-muted-foreground hover:text-foreground lg:hidden"
            >
              <Menu className="size-4" />
            </button>
            <div>
              <div className="text-base font-bold text-foreground">Advertiser Console</div>
              <div className="text-xs text-muted-foreground">
                Manage campaigns, creatives, and spend
              </div>
            </div>
          </div>
          <div className="hidden items-center gap-4 md:flex">
            <div className="rounded-full border border-border bg-card px-3 py-1.5 text-xs">
              <span className="text-muted-foreground">Balance · </span>
              <span className="num font-semibold text-foreground">{KES(balance)}</span>
            </div>
          </div>
        </header>
        <motion.main
          key={useRouterState({ select: (s) => s.location.pathname })}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.25 }}
          className="min-w-0 flex-1 p-4 md:p-8"
        >
          <Outlet />
        </motion.main>
      </div>
    </div>
  );
}
