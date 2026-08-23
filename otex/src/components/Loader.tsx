import { useEffect, useState } from "react";
import { LogoFull } from "./Logo";

export function Loader() {
  const [gone, setGone] = useState(false);
  const [fade, setFade] = useState(false);

  useEffect(() => {
    const t1 = setTimeout(() => setFade(true), 1400);
    const t2 = setTimeout(() => setGone(true), 1900);
    return () => { clearTimeout(t1); clearTimeout(t2); };
  }, []);

  if (gone) return null;
  return (
    <div
      className={`fixed inset-0 z-[100] flex flex-col items-center justify-center bg-[var(--cream)] transition-opacity duration-500 ${fade ? "opacity-0" : "opacity-100"}`}
      aria-hidden
    >
      <div className="relative flex items-center justify-center">
        <span className="absolute h-40 w-40 rounded-full bg-[var(--sun)]/30 blur-3xl loader-pulse" />
        <span className="absolute h-28 w-28 rounded-full border border-[var(--sun)]/40 loader-ring" />
        <LogoFull className="relative h-14 md:h-16 float-slow drop-shadow-[0_8px_30px_rgba(0,0,0,0.12)]" />
      </div>
      <div className="mt-10 text-center">
        <div className="text-[11px] uppercase tracking-[0.32em] text-foreground/45">
          Self-serve ad network · Billed in shillings
        </div>
      </div>
      <style>{`
        @keyframes loader-pulse { 0%,100%{opacity:.35;transform:scale(.95)} 50%{opacity:.7;transform:scale(1.08)} }
        .loader-pulse { animation: loader-pulse 2.2s ease-in-out infinite; }
        @keyframes loader-ring { 0%{transform:scale(.9);opacity:.6} 100%{transform:scale(1.6);opacity:0} }
        .loader-ring { animation: loader-ring 2s ease-out infinite; }
      `}</style>
    </div>
  );
}
