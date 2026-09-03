type GtagFn = (...args: unknown[]) => void;

const GOOGLE_ADS_ID = import.meta.env.VITE_GOOGLE_ADS_ID || "AW-18351116680";
const SIGNUP_CONVERSION_LABEL = import.meta.env.VITE_GOOGLE_ADS_SIGNUP_LABEL || "";

function getGtag(): GtagFn | null {
  if (typeof window === "undefined") return null;
  const fn = (window as unknown as { gtag?: GtagFn }).gtag;
  return typeof fn === "function" ? fn : null;
}

export function trackSignupConversion(role: "advertiser" | "publisher") {
  const gtag = getGtag();
  if (!gtag) return;

  gtag("event", "sign_up", { method: "email", account_type: role });

  if (SIGNUP_CONVERSION_LABEL) {
    gtag("event", "conversion", {
      send_to: `${GOOGLE_ADS_ID}/${SIGNUP_CONVERSION_LABEL}`,
    });
  }
}
