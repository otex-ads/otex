import { X } from "lucide-react";
import { useEffect, type ReactNode } from "react";
import { motion, AnimatePresence } from "framer-motion";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: ReactNode;
  size?: "sm" | "md" | "lg";
}

export function Modal({ open, onClose, title, description, children, size = "md" }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  const widths = { sm: "max-w-md", md: "max-w-lg", lg: "max-w-2xl" };

  return (
    <AnimatePresence>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="absolute inset-0 bg-black/50 backdrop-blur-sm"
            onClick={onClose}
          />
          <motion.div
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: 8 }}
            transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
            className={`relative flex max-h-[85vh] w-full ${widths[size]} flex-col rounded-2xl border border-border bg-card p-6 shadow-2xl`}
          >
            <div className="mb-5 flex shrink-0 items-start justify-between gap-4">
              <div>
                <h2 className="text-base font-semibold text-foreground">{title}</h2>
                {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
              </div>
              <button
                onClick={onClose}
                className="grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-muted"
              >
                <X className="size-4" />
              </button>
            </div>
            <div className="overflow-y-auto">{children}</div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}

interface FormFieldProps {
  label: string;
  required?: boolean;
  children: ReactNode;
  hint?: string;
}

export function FormField({ label, required, children, hint }: FormFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium text-foreground">
        {label}{required && <span className="ml-0.5 text-danger">*</span>}
      </label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  );
}

export const inputCls = "h-9 w-full rounded-lg border border-border bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/40";
export const selectCls = "h-9 w-full rounded-lg border border-border bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring/40";

interface ModalActionsProps {
  onCancel: () => void;
  onConfirm: () => void;
  loading?: boolean;
  confirmLabel?: string;
  confirmVariant?: "primary" | "danger" | "success";
}

export function ModalActions({ onCancel, onConfirm, loading, confirmLabel = "Confirm", confirmVariant = "primary" }: ModalActionsProps) {
  const variants = {
    primary: "bg-primary text-primary-foreground hover:brightness-110",
    danger: "bg-danger text-danger-foreground hover:brightness-110",
    success: "bg-success text-success-foreground hover:brightness-110",
  };
  return (
    <div className="mt-6 flex justify-end gap-2 border-t border-border pt-4">
      <button
        type="button"
        onClick={onCancel}
        className="rounded-full border border-border bg-card px-4 py-2 text-sm font-medium text-foreground hover:bg-muted/50"
      >
        Cancel
      </button>
      <button
        type="button"
        onClick={onConfirm}
        disabled={loading}
        className={`rounded-full px-4 py-2 text-sm font-medium shadow-sm transition ${variants[confirmVariant]} disabled:opacity-50`}
      >
        {loading ? "Please wait…" : confirmLabel}
      </button>
    </div>
  );
}
