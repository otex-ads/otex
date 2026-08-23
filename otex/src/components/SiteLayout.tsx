import { Link } from "@tanstack/react-router";
import { ArrowUpRight } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import { LogoFull, LogoMark } from "./Logo";

const nav = [
  { href: "/#advertisers", label: "For Advertisers" },
  { href: "/#publishers", label: "For Publishers" },
  { href: "/#formats", label: "Formats" },
  { href: "/pricing", label: "Pricing" },
] as const;

export function SiteHeader() {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const fn = () => setScrolled(window.scrollY > 8);
    fn();
    window.addEventListener("scroll", fn);
    return () => window.removeEventListener("scroll", fn);
  }, []);
  return (
    <header
      className={`fixed top-0 inset-x-0 z-50 transition-colors duration-300 bg-[var(--cream)] ${
        scrolled ? "backdrop-blur-xl bg-[var(--cream)]/90 border-b border-border/60" : ""
      }`}
    >
      <div className="mx-auto max-w-6xl px-6 h-20 md:h-24 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-3 group" aria-label="OtexAds — Home">
          <LogoFull className="h-8 md:h-9" />
        </Link>
        <nav className="hidden md:flex items-center gap-9 text-[13px] text-foreground/65">
          {nav.map((n) =>
            n.href.startsWith("/#") ? (
              <a key={n.href} href={n.href} className="hover:text-foreground transition-colors">
                {n.label}
              </a>
            ) : (
              <Link
                key={n.href}
                to={n.href}
                className="hover:text-foreground transition-colors"
                activeProps={{ className: "text-foreground" }}
              >
                {n.label}
              </Link>
            )
          )}
        </nav>
        <div className="flex items-center gap-1.5">
          <a
            href="https://advertiser.otexads.com/auth/login"
            className="hidden sm:inline-flex text-[13px] px-3 py-1.5 text-foreground/70 hover:text-foreground transition"
          >
            Login
          </a>
          <a
            href="https://publisher.otexads.com/auth/register"
            className="hidden md:inline-flex text-[13px] font-medium px-3.5 py-1.5 rounded-full border border-border text-foreground/80 hover:text-foreground hover:border-foreground/40 transition"
          >
            Monetize
          </a>
          <a
            href="https://advertiser.otexads.com/auth/register"
            className="inline-flex items-center gap-1 text-[13px] font-medium px-3.5 py-1.5 rounded-full bg-foreground text-background hover:opacity-90 transition"
          >
            Start advertising <ArrowUpRight className="h-3 w-3" />
          </a>
        </div>
      </div>
    </header>
  );
}

export function SiteFooter() {
  return (
    <footer className="border-t border-border/60 mt-32">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="grid md:grid-cols-4 gap-10 text-sm">
          <div className="md:col-span-2">
            <Link to="/" className="flex items-center gap-2">
              <LogoMark className="h-7" />
            </Link>
            <p className="mt-4 text-foreground/55 max-w-xs leading-relaxed">
              Self-serve ad network for African advertisers and publishers. Billed in shillings, paid via M-Pesa. A Siohioma Group company.
            </p>
          </div>
          <div>
            <div className="text-[11px] uppercase tracking-[0.2em] text-foreground/40 mb-4">Product</div>
            <ul className="space-y-2.5">
              <li><a href="/#advertisers" className="text-foreground/70 hover:text-foreground transition">For Advertisers</a></li>
              <li><a href="/#publishers" className="text-foreground/70 hover:text-foreground transition">For Publishers</a></li>
              <li><a href="/#formats" className="text-foreground/70 hover:text-foreground transition">Ad Formats</a></li>
              <li><Link to="/pricing" className="text-foreground/70 hover:text-foreground transition">Pricing</Link></li>
            </ul>
          </div>
          <div>
            <div className="text-[11px] uppercase tracking-[0.2em] text-foreground/40 mb-4">Sign in</div>
            <ul className="space-y-2.5">
              <li><a href="https://advertiser.otexads.com/auth/login" className="text-foreground/70 hover:text-foreground transition">Advertiser Login</a></li>
              <li><a href="https://publisher.otexads.com/auth/login" className="text-foreground/70 hover:text-foreground transition">Publisher Login</a></li>
              <li><a href="mailto:support@otexads.com" className="text-foreground/70 hover:text-foreground transition">Support</a></li>
            </ul>
          </div>
        </div>
        <div className="mt-14 pt-6 border-t border-border/60 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs text-foreground/45">
          <span>© {new Date().getFullYear()} OtexAds. All rights reserved.</span>
          <span>Built in Africa. Paid in M-Pesa.</span>
        </div>
      </div>
    </footer>
  );
}

export function SiteLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground">
      <SiteHeader />
      <main className="flex-1 pt-20 md:pt-24">{children}</main>
      <SiteFooter />
    </div>
  );
}

export function PageHeader({
  eyebrow,
  title,
  lede,
}: {
  eyebrow?: string;
  title: ReactNode;
  lede?: ReactNode;
}) {
  return (
    <section className="mx-auto max-w-4xl px-6 pt-28 pb-16 text-center">
      {eyebrow && (
        <div className="text-[11px] uppercase tracking-[0.28em] text-foreground/45 mb-6">{eyebrow}</div>
      )}
      <h1 className="text-balance text-[clamp(2.25rem,5.5vw,4.25rem)] font-medium tracking-[-0.03em] leading-[1.05]">
        {title}
      </h1>
      {lede && (
        <p className="mt-6 text-lg md:text-xl text-foreground/60 max-w-2xl mx-auto leading-relaxed">{lede}</p>
      )}
    </section>
  );
}
