export function LogoMark({ className = "h-8 w-8" }: { className?: string }) {
  return (
    <img
      src="/logo.png"
      alt="OtexAds"
      className={`object-contain ${className}`}
      loading="eager"
      decoding="async"
    />
  );
}

export function LogoFull({ className = "h-8" }: { className?: string }) {
  return (
    <img
      src="/logo.png"
      alt="OtexAds — Self-serve ad network"
      className={`object-contain ${className}`}
      loading="eager"
      decoding="async"
    />
  );
}
