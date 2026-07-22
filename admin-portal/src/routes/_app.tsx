import { createFileRoute, Outlet, Link, useRouterState, redirect } from "@tanstack/react-router";
import { motion } from "framer-motion";
import {
  LayoutDashboard, Users, TrendingUp, DollarSign, Settings as SettingsIcon, Menu, X, LogOut, CheckCircle,
} from "lucide-react";
import { useState } from "react";
import { clearToken, getToken, getUser } from "@/lib/api";

export const Route = createFileRoute("/_app")({
  // Client-only: we gate on localStorage which isn't available during SSR.
  ssr: false,
  beforeLoad: () => {
    if (!getToken()) {
      throw redirect({ to: "/auth/login" });
    }
  },
  component: AppLayout,
});

const menu = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard, exact: true },
  { to: "/approvals", label: "Approvals", icon: CheckCircle },
  { to: "/users", label: "Users", icon: Users },
  { to: "/campaigns", label: "Campaigns", icon: TrendingUp },
  { to: "/revenue", label: "Revenue", icon: DollarSign },
  { to: "/analytics", label: "Analytics", icon: TrendingUp },
  { to: "/settings", label: "Settings", icon: SettingsIcon },
] as const;

function NavItem({ to, label, icon: Icon, exact, onClick }: {
  to: string; label: string; icon: typeof LayoutDashboard; exact?: boolean; onClick?: () => void;
}) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const active = exact ? pathname === to : pathname === to || pathname.startsWith(to + "/");
  return (
    <Link
      to={to}
      onClick={onClick}
      className={`group flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
        active ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
      }`}
    >
      <Icon className="size-4" />
      <span className="font-medium">{label}</span>
    </Link>
  );
}

function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const user = getUser();

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
          <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Admin</div>
        </div>
      </Link>
      <div className="label-eyebrow mb-2 px-3">Admin Console</div>
      <nav className="flex flex-col gap-0.5">
        {menu.map((m) => <NavItem key={m.to} {...m} onClick={onNav} />)}
      </nav>
      <div className="mt-auto flex flex-col gap-3">
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
        <div className="fixed inset-0 z-40 bg-black/50 lg:hidden" onClick={() => setMobileOpen(false)} />
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
              <div className="text-base font-bold text-foreground">Admin Console</div>
              <div className="text-xs text-muted-foreground">
                {user?.email ? `Signed in as ${user.email}` : "Manage OtexAds platform"}
              </div>
            </div>
          </div>
          <div className="hidden items-center gap-4 md:flex">
            <div className="rounded-full border border-border bg-card px-3 py-1.5 text-xs">
              <span className="text-muted-foreground">Admin Console</span>
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
