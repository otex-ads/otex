import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

export function Shimmer({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-md bg-muted",
        className,
      )}
    >
      <motion.div
        className="absolute inset-0"
        style={{
          backgroundImage:
            "linear-gradient(90deg, transparent 0%, oklch(1 0 0 / 0.6) 50%, transparent 100%)",
        }}
        initial={{ x: "-100%" }}
        animate={{ x: "100%" }}
        transition={{ duration: 1.4, repeat: Infinity, ease: "linear" }}
      />
    </div>
  );
}

export function PageLoader() {
  return (
    <div className="space-y-6">
      <div className="flex gap-4">
        {[0, 1, 2, 3].map((i) => (
          <Shimmer key={i} className="h-28 flex-1" />
        ))}
      </div>
      <Shimmer className="h-80" />
    </div>
  );
}