import { Search } from "lucide-react";
import { motion } from "framer-motion";
import type { ReactNode } from "react";

export function PageHeader({
  title,
  description,
  searchPlaceholder,
  actions,
}: {
  title: string;
  description?: string;
  searchPlaceholder?: string;
  actions?: ReactNode;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: -6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
      className="mb-8 flex flex-wrap items-end justify-between gap-4"
    >
      <div>
        <div className="label-eyebrow mb-2">Overview</div>
        <h1 className="display text-4xl font-normal tracking-tight text-foreground md:text-[2.75rem]">{title}</h1>
        {description && (
          <p className="mt-2 max-w-xl text-sm text-muted-foreground">{description}</p>
        )}
      </div>
      <div className="flex items-center gap-3">
        {searchPlaceholder && (
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              placeholder={searchPlaceholder}
              className="h-10 w-72 rounded-full border border-border bg-card pl-9 pr-4 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/40"
            />
          </div>
        )}
        {actions}
      </div>
    </motion.div>
  );
}