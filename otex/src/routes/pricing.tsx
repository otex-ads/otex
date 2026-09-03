import { createFileRoute } from "@tanstack/react-router";
import { Check } from "lucide-react";
import { SiteLayout, PageHeader } from "@/components/SiteLayout";

export const Route = createFileRoute("/pricing")({
  component: PricingPage,
  head: () => ({
    meta: [
      { title: "Pricing — OtexAds" },
      { name: "description", content: "No minimum spend for advertisers. Weekly M-Pesa payouts for publishers from KES 500. Priced in shillings — no FX tax." },
      { property: "og:title", content: "Pricing — OtexAds" },
      { property: "og:description", content: "CPM-based advertising, revenue share for publishers, M-Pesa settlement." },
      { property: "og:url", content: "/pricing" },
    ],
    links: [{ rel: "canonical", href: "/pricing" }],
  }),
});

const advertiserTiers = [
  {
    name: "Self-serve",
    price: "No minimum",
    suffix: "· CPM / CPC bidding",
    desc: "Launch a campaign the same day. Top up any amount.",
    features: [
      "All 5 ad formats — Push, Native, Popunder, In-Page Push, Banner",
      "Real-time stats & budget pacing",
      "Country, device, browser & category targeting",
      "M-Pesa, card & bank top-up",
      "Campaign review within 24 hours",
    ],
    cta: "Start advertising",
    href: "https://advertiser.otexads.com/auth/register",
    highlight: true,
  },
  {
    name: "Managed",
    price: "From KES 250k",
    suffix: "/ month spend",
    desc: "For agencies and brands with a dedicated media budget.",
    features: [
      "Named account manager",
      "Creative & landing-page review",
      "Custom whitelists & audience packages",
      "Priority campaign approval",
      "Consolidated monthly invoice",
    ],
    cta: "Talk to sales",
    href: "mailto:support@otexads.com",
    highlight: false,
  },
];

function PricingPage() {
  return (
    <SiteLayout>
      <PageHeader
        eyebrow="Pricing"
        title={<>Simple. <span className="font-serif-italic">Honest.</span> In shillings.</>}
        lede="No per-seat traps. No FX surprises. Advertisers pay per impression or click. Publishers get paid weekly, in M-Pesa."
      />

      <section className="mx-auto max-w-6xl px-6 pb-24">
        <div className="text-[11px] uppercase tracking-[0.28em] text-foreground/45 mb-6 text-center">
          For Advertisers
        </div>
        <div className="grid md:grid-cols-2 gap-5">
          {advertiserTiers.map((t) => (
            <div
              key={t.name}
              className={`relative rounded-2xl border p-8 ${
                t.highlight ? "border-foreground bg-card shadow-sm" : "border-border bg-background"
              }`}
            >
              {t.highlight && (
                <div className="absolute -top-2.5 left-8 text-[10px] tracking-[0.25em] uppercase bg-foreground text-background px-2.5 py-1 rounded-full">
                  Most popular
                </div>
              )}
              <div className="text-sm text-foreground/55">{t.name}</div>
              <div className="mt-4 flex items-baseline gap-2 flex-wrap">
                <span className="text-3xl font-black tracking-tight">{t.price}</span>
                {t.suffix && <span className="text-sm text-foreground/55">{t.suffix}</span>}
              </div>
              <p className="mt-2 text-sm text-foreground/60">{t.desc}</p>
              <ul className="mt-8 space-y-3 text-sm">
                {t.features.map((f) => (
                  <li key={f} className="flex items-start gap-2.5">
                    <Check className="h-4 w-4 mt-0.5 text-[var(--herb)] shrink-0" />
                    <span className="text-foreground/75">{f}</span>
                  </li>
                ))}
              </ul>
              <a
                href={t.href}
                className={`mt-8 inline-flex w-full items-center justify-center px-4 py-2.5 rounded-full font-bold text-sm transition ${
                  t.highlight
                    ? "bg-foreground text-background hover:opacity-90"
                    : "border border-border hover:bg-card"
                }`}
              >
                {t.cta}
              </a>
            </div>
          ))}
        </div>
        <p className="mt-10 text-center text-xs text-foreground/45">
          Prices in Kenyan Shillings. Pay by M-Pesa, card, or bank transfer. VAT included where applicable.
        </p>
      </section>
    </SiteLayout>
  );
}
