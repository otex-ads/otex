import { createFileRoute, useRouter } from "@tanstack/react-router";
import { ArrowRight, ArrowUpRight, Bell, LayoutGrid, MousePointerClick, Rows3, Square, Check } from "lucide-react";
import { useMemo, useState } from "react";
import { SiteLayout } from "@/components/SiteLayout";
import { Loader } from "@/components/Loader";
import { TypewriterText } from "@/components/TypewriterText";

export const Route = createFileRoute("/")({
  component: Index,
  head: () => ({
    meta: [
      { title: "OtexAds — Self-Serve Ad Network | Push, Native & Popunder Ads" },
      { name: "description", content: "Advertise or monetize with OtexAds. 5 ad formats, real-time stats, M-Pesa payouts. Built for African publishers and advertisers." },
      { name: "theme-color", content: "#F4F1EA" },
      { property: "og:title", content: "OtexAds — Self-Serve Ad Network" },
      { property: "og:description", content: "5 ad formats, real-time stats, M-Pesa payouts. Built for African publishers and advertisers." },
      { property: "og:type", content: "website" },
      { property: "og:url", content: "/" },
      { name: "twitter:card", content: "summary_large_image" },
    ],
    links: [
      { rel: "icon", type: "image/png", href: "/favicon.png" },
      { rel: "apple-touch-icon", href: "/favicon.png" },
      { rel: "canonical", href: "/" },
    ],
    scripts: [
      {
        type: "application/ld+json",
        children: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "SoftwareApplication",
          name: "OtexAds",
          applicationCategory: "BusinessApplication",
          operatingSystem: "Web",
          description: "Self-serve ad network connecting advertisers with publisher traffic across Africa. Push, native, popunder, in-page push, and banner ads. M-Pesa billing and payouts.",
          offers: { "@type": "Offer", price: "0", priceCurrency: "KES", description: "No minimum spend for advertisers" },
          publisher: { "@type": "Organization", name: "OtexAds" },
          featureList: [
            "Push ads",
            "Native ads",
            "Popunder ads",
            "In-Page Push ads",
            "Banner ads",
            "M-Pesa billing and payouts",
            "Real-time campaign stats",
            "Weekly publisher payouts",
          ],
        }),
      },
    ],
  }),
});

function Index() {
  return (
    <SiteLayout>
      <Loader />
      <Hero />
      <StatsBar />
      <Sides />
      <HowItWorks />
      <Formats />
      <ClosingCTA />
    </SiteLayout>
  );
}

function Hero() {
  const router = useRouter();
  const segments = useMemo(
    () => [
      { text: "Traffic that converts," },
      { text: " revenue that scales.", className: "font-serif-italic font-normal text-foreground/90" },
    ],
    []
  );

  return (
    <section className="relative">
      <div className="mx-auto max-w-7xl px-6 pt-10 md:pt-14">
        <h2 className="text-center md:text-left text-[clamp(1.5rem,3vw,2.5rem)] font-black tracking-[-0.055em] leading-[0.9]">
          OtexAds
        </h2>
      </div>
      <div className="mx-auto max-w-7xl px-6 pt-10 md:pt-14 pb-16 md:pb-24">
        <div className="grid md:grid-cols-2 gap-10 md:gap-16 items-center">
          <div className="fade-up">
            <h1 className="text-balance text-[clamp(2.75rem,6vw,5rem)] font-black tracking-[-0.045em] leading-[0.98]">
              <TypewriterText
                trigger={router.state.location.href}
                segments={segments}
                speed={40}
              />
            </h1>
            <p className="mt-7 text-lg md:text-xl text-foreground/70 max-w-lg leading-relaxed">
              Buy traffic that <span className="font-serif-italic">converts.</span> Monetize traffic you <span className="font-serif-italic">own.</span>
            </p>
            <p className="mt-4 text-[15px] md:text-base text-foreground/60 max-w-lg leading-relaxed">
              Push, native, popunder, and banner ads — self-serve for advertisers, instant payouts for publishers. Billed in shillings, paid via M-Pesa.
            </p>
            <div className="mt-10 flex flex-col sm:flex-row gap-3">
              <a
                href="https://advertiser.otexads.com/auth/register"
                className="group inline-flex items-center justify-center gap-2 px-6 py-3 rounded-full bg-foreground text-background font-bold hover:opacity-90 transition"
              >
                Start advertising
                <ArrowRight className="h-4 w-4 group-hover:translate-x-0.5 transition" />
              </a>
              <a
                href="https://publisher.otexads.com/auth/register"
                className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-full border border-border bg-background/60 font-semibold hover:bg-background transition"
              >
                Monetize your site
              </a>
            </div>
            <a href="#how" className="mt-6 inline-block text-sm text-foreground/55 hover:text-foreground transition">
              See how it works ↓
            </a>
          </div>
          <HeroPanel />
        </div>
      </div>
    </section>
  );
}

function HeroPanel() {
  return (
    <div className="relative aspect-[4/5] rounded-3xl overflow-hidden border border-border/70 bg-[var(--ink)] text-[var(--cream)] shadow-[0_30px_80px_-30px_rgba(0,0,0,0.35)] p-6 md:p-8">
      <div className="flex items-center justify-between text-[11px] uppercase tracking-[0.24em] text-[var(--cream)]/60">
        <span>Live · Last 24h</span>
        <span className="inline-flex items-center gap-1.5">
          <span className="h-1.5 w-1.5 rounded-full bg-[var(--herb)] animate-pulse" /> Serving
        </span>
      </div>
      <div className="mt-6">
        <div className="text-[11px] uppercase tracking-[0.2em] text-[var(--cream)]/50">Impressions</div>
        <div className="mt-1 text-5xl md:text-6xl font-black tracking-[-0.04em] tabular-nums">4.82M</div>
        <div className="mt-1 text-sm text-[var(--herb)]">▲ 12.4% vs. yesterday</div>
      </div>
      <div className="mt-8 grid grid-cols-2 gap-4">
        <Stat label="CTR" value="1.84%" />
        <Stat label="eCPM" value="KES 42" />
        <Stat label="Conversions" value="2,391" />
        <Stat label="Spend" value="KES 118k" />
      </div>
      <div className="mt-8 pt-6 border-t border-white/10">
        <div className="text-[11px] uppercase tracking-[0.2em] text-[var(--cream)]/50 mb-3">Next payout</div>
        <div className="flex items-center justify-between">
          <div>
            <div className="font-semibold">M-Pesa · +254 7•• ••• 421</div>
            <div className="text-xs text-[var(--cream)]/50 mt-1">Weekly · Fri 09:00 EAT</div>
          </div>
          <div className="text-xl font-black tabular-nums">KES 24,180</div>
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-white/[0.04] border border-white/10 p-3">
      <div className="text-[10px] uppercase tracking-[0.2em] text-[var(--cream)]/45">{label}</div>
      <div className="mt-1 text-lg font-bold tabular-nums">{value}</div>
    </div>
  );
}

function StatsBar() {
  const stats = [
    { n: "3,200+", l: "Publishers" },
    { n: "47", l: "Countries" },
    { n: "180M", l: "Impressions / day" },
    { n: "KES 8.4M", l: "Paid out last month" },
  ];
  return (
    <section className="mx-auto max-w-6xl px-6 pb-16">
      <div className="rounded-2xl border border-border/70 bg-card grid grid-cols-2 md:grid-cols-4 divide-x divide-y md:divide-y-0 divide-border/60 overflow-hidden">
        {stats.map((s) => (
          <div key={s.l} className="px-6 py-6 text-center">
            <div className="text-2xl md:text-3xl font-black tracking-[-0.03em] tabular-nums">{s.n}</div>
            <div className="mt-1 text-[11px] uppercase tracking-[0.2em] text-foreground/50">{s.l}</div>
          </div>
        ))}
      </div>
    </section>
  );
}

function Sides() {
  const advertiser = [
    "Self-serve campaign creation — no minimum spend, no account manager required",
    "5 ad formats: Push, Native, Popunder, In-Page Push, Banner",
    "Real-time stats, budget pacing, M-Pesa top-up",
  ];
  const publisher = [
    "Monetize any site with one tag",
    "Weekly M-Pesa payouts, KES 500 minimum threshold",
    "Fraud-filtered traffic = advertisers stay, rates stay high",
  ];
  return (
    <section className="mx-auto max-w-6xl px-6 py-24">
      <div className="text-center mb-14">
        <div className="text-[11px] uppercase tracking-[0.28em] text-foreground/45 mb-4">Two sides, one network</div>
        <h2 className="text-balance text-[clamp(1.75rem,4vw,2.75rem)] font-black tracking-[-0.035em]">
          Built for both ends of the <span className="font-serif-italic font-normal">exchange.</span>
        </h2>
      </div>
      <div className="grid md:grid-cols-2 gap-5">
        <SideCard
          id="advertisers"
          eyebrow="For Advertisers"
          title={<>Launch a campaign in <span className="font-serif-italic font-normal">minutes,</span> not meetings</>}
          bullets={advertiser}
          cta={{ label: "Start advertising", href: "https://advertiser.otexads.com/auth/register" }}
        />
        <SideCard
          id="publishers"
          eyebrow="For Publishers"
          title={<>Turn your traffic into <span className="font-serif-italic font-normal">shillings</span></>}
          bullets={publisher}
          cta={{ label: "Monetize your site", href: "https://publisher.otexads.com/auth/register" }}
          tone="ink"
        />
      </div>
    </section>
  );
}

function SideCard({
  id,
  eyebrow,
  title,
  bullets,
  cta,
  tone = "cream",
}: {
  id: string;
  eyebrow: string;
  title: React.ReactNode;
  bullets: string[];
  cta: { label: string; href: string };
  tone?: "cream" | "ink";
}) {
  const isInk = tone === "ink";
  return (
    <div
      id={id}
      className={`scroll-mt-24 rounded-3xl border p-8 md:p-10 ${
        isInk
          ? "bg-[var(--ink)] text-[var(--cream)] border-transparent"
          : "bg-card border-border/70"
      }`}
    >
      <div className={`text-[11px] uppercase tracking-[0.25em] mb-4 ${isInk ? "text-[var(--cream)]/55" : "text-foreground/45"}`}>
        {eyebrow}
      </div>
      <h3 className="text-2xl md:text-3xl font-black tracking-[-0.035em] leading-[1.05]">{title}</h3>
      <ul className="mt-7 space-y-3.5">
        {bullets.map((b) => (
          <li key={b} className="flex items-start gap-3 text-[15px] leading-relaxed">
            <Check className={`h-4 w-4 mt-1 shrink-0 ${isInk ? "text-[var(--sun)]" : "text-[var(--herb)]"}`} />
            <span className={isInk ? "text-[var(--cream)]/85" : "text-foreground/80"}>{b}</span>
          </li>
        ))}
      </ul>
      <a
        href={cta.href}
        className={`mt-8 inline-flex items-center gap-2 px-5 py-2.5 rounded-full font-semibold text-sm transition ${
          isInk
            ? "bg-[var(--cream)] text-[var(--ink)] hover:opacity-90"
            : "bg-foreground text-background hover:opacity-90"
        }`}
      >
        {cta.label} <ArrowUpRight className="h-3.5 w-3.5" />
      </a>
    </div>
  );
}

function HowItWorks() {
  const [tab, setTab] = useState<"advertiser" | "publisher">("advertiser");
  const advertiser = [
    { t: "Sign up", d: "Create your OtexAds account. Campaign review typically completes within 24 hours." },
    { t: "Set up", d: "Build your campaign, upload creatives, and pick targeting — countries, devices, categories." },
    { t: "Start scaling", d: "Top up via M-Pesa. Real-time stats, budget pacing, and pause/resume in one click." },
  ];
  const publisher = [
    { t: "Sign up", d: "Instant approval. Add your site and get access to your publisher dashboard." },
    { t: "Set up", d: "Create a zone, drop one script tag, and pick your ad formats — push, native, popunder." },
    { t: "Get paid", d: "Weekly M-Pesa payouts. KES 500 minimum. No net-30, no wire fees, no waiting." },
  ];
  const steps = tab === "advertiser" ? advertiser : publisher;

  return (
    <section id="how" className="mx-auto max-w-6xl px-6 py-24 scroll-mt-24">
      <div className="text-center mb-10">
        <div className="text-[11px] uppercase tracking-[0.28em] text-foreground/45 mb-4">How it works</div>
        <h2 className="text-balance text-[clamp(1.75rem,4vw,2.75rem)] font-black tracking-[-0.035em]">
          Three steps. <span className="font-serif-italic font-normal">Either side.</span>
        </h2>
      </div>
      <div className="flex justify-center mb-10">
        <div className="inline-flex p-1 rounded-full border border-border/70 bg-card">
          {(["advertiser", "publisher"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-1.5 text-sm rounded-full font-semibold capitalize transition ${
                tab === t ? "bg-foreground text-background" : "text-foreground/60 hover:text-foreground"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>
      <ol className="grid md:grid-cols-3 gap-5">
        {steps.map((s, i) => (
          <li key={s.t} className="rounded-2xl border border-border/70 bg-card p-7">
            <div className="flex items-center gap-3">
              <div className="h-8 w-8 rounded-full bg-foreground text-background grid place-items-center text-sm font-black tabular-nums">
                {i + 1}
              </div>
              <div className="text-lg font-black tracking-tight">{s.t}</div>
            </div>
            <p className="mt-4 text-[15px] leading-relaxed text-foreground/70">{s.d}</p>
          </li>
        ))}
      </ol>
    </section>
  );
}

function Formats() {
  const items = [
    { icon: Bell, name: "Push", desc: "Native OS-style notifications that reach users even after they leave your site." },
    { icon: LayoutGrid, name: "Native", desc: "In-feed placements that match the look and feel of surrounding content." },
    { icon: Square, name: "Popunder", desc: "Full-page placements that open behind the current tab — high reach, low intrusion." },
    { icon: MousePointerClick, name: "In-Page Push", desc: "Push-style widgets rendered inside the page. Works on iOS and every browser." },
    { icon: Rows3, name: "Banner", desc: "Classic IAB display sizes — 300×250, 728×90, 320×50. Fast-loading, brand-safe." },
  ];
  return (
    <section id="formats" className="mx-auto max-w-6xl px-6 py-24 scroll-mt-24">
      <div className="text-center mb-12">
        <div className="text-[11px] uppercase tracking-[0.28em] text-foreground/45 mb-4">Ad formats</div>
        <h2 className="text-balance text-[clamp(1.75rem,4vw,2.75rem)] font-black tracking-[-0.035em]">
          Five formats. <span className="font-serif-italic font-normal">One integration.</span>
        </h2>
      </div>
      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-px bg-border rounded-2xl overflow-hidden border border-border">
        {items.map((i) => (
          <div key={i.name} className="bg-background p-7 hover:bg-card transition">
            <div className="h-11 w-11 rounded-xl bg-foreground/[0.04] border border-border/60 grid place-items-center">
              <i.icon className="h-5 w-5 text-foreground/75" strokeWidth={1.6} />
            </div>
            <div className="mt-5 text-xl font-black tracking-tight">{i.name}</div>
            <p className="mt-2 text-[15px] leading-relaxed text-foreground/70">{i.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

function ClosingCTA() {
  return (
    <section className="mx-auto max-w-3xl px-6 py-32 text-center">
      <h2 className="text-balance text-[clamp(2rem,5vw,3.5rem)] font-medium tracking-[-0.03em] leading-[1.05]">
        Ship campaigns. <span className="font-serif-italic">Grow revenue.</span> Get paid.
      </h2>
      <p className="mt-5 text-lg text-foreground/60">
        Whether you're spending or earning, OtexAds settles in shillings — straight to M-Pesa.
      </p>
      <div className="mt-9 flex flex-col sm:flex-row gap-3 justify-center">
        <a
          href="https://advertiser.otexads.com/auth/register"
          className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-full bg-foreground text-background font-medium hover:opacity-90 transition"
        >
          Start advertising <ArrowUpRight className="h-4 w-4" />
        </a>
        <a
          href="https://publisher.otexads.com/auth/register"
          className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-full border border-border font-medium hover:bg-card transition"
        >
          Monetize your site
        </a>
      </div>
    </section>
  );
}
