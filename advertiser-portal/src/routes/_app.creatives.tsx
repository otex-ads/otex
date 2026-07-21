import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { Plus, Trash2, ImageIcon, Upload } from "lucide-react";
import { SurfaceCard, SectionLabel } from "@/components/Card";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { Modal } from "@/components/Modal";
import { store, useStore } from "@/lib/store";
import { type CreativeFormat } from "@/lib/api";
import { formatDate } from "@/lib/format";

export const Route = createFileRoute("/_app/creatives")({
  component: CreativesPage,
});

const FORMATS: CreativeFormat[] = ["push", "popunder", "native", "banner", "interstitial"];

function CreativesPage() {
  const creatives = useStore((s) => s.creatives) || [];
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [format, setFormat] = useState<CreativeFormat>("push");
  const [headline, setHeadline] = useState("");
  const [description, setDescription] = useState("");
  const [landingUrl, setLandingUrl] = useState("");
  const [imageUrl, setImageUrl] = useState("");

  const onFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    const reader = new FileReader();
    reader.onload = () => setImageUrl(String(reader.result));
    reader.readAsDataURL(f);
  };

  const reset = () => {
    setName(""); setFormat("push"); setHeadline(""); setDescription(""); setLandingUrl(""); setImageUrl("");
  };

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !landingUrl.trim()) return toast.error("Name and landing URL are required");
    store.addCreative({ name: name.trim(), format, headline, description, landingUrl: landingUrl.trim(), imageUrl });
    toast.success("Creative added");
    setOpen(false); reset();
  };

  return (
    <div>
      <PageHeader
        title="Creatives"
        description="Reusable ad assets — attach them to campaigns of matching format."
        actions={
          <button onClick={() => setOpen(true)} className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:brightness-110">
            <Plus className="size-4" /> New creative
          </button>
        }
      />

      {creatives.length === 0 ? (
        <EmptyState
          icon={ImageIcon}
          title="No creatives yet"
          description="Upload creatives once, reuse them across campaigns."
          action={<button onClick={() => setOpen(true)} className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"><Plus className="size-4" /> Add creative</button>}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {creatives.map((c) => (
            <SurfaceCard key={c.id}>
              <div className="mb-3 flex aspect-video items-center justify-center overflow-hidden rounded-lg bg-muted">
                {c.imageUrl ? (
                  <img src={c.imageUrl} alt={c.name} className="size-full object-cover" />
                ) : (
                  <ImageIcon className="size-8 text-muted-foreground/40" />
                )}
              </div>
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="truncate text-sm font-semibold text-foreground">{c.name}</div>
                  <div className="mt-0.5 text-xs capitalize text-muted-foreground">{c.format} · {formatDate(c.createdAt)}</div>
                </div>
                <button
                  onClick={() => { if (confirm(`Delete "${c.name}"?`)) store.removeCreative(c.id); }}
                  className="grid size-8 place-items-center rounded-md text-danger hover:bg-danger-soft"
                ><Trash2 className="size-3.5" /></button>
              </div>
              {c.headline && <div className="mt-3 text-sm text-foreground">{c.headline}</div>}
              {c.description && <div className="mt-1 text-xs text-muted-foreground">{c.description}</div>}
              <div className="mt-3 truncate text-xs text-primary" title={c.landingUrl}>{c.landingUrl}</div>
            </SurfaceCard>
          ))}
        </div>
      )}

      <Modal open={open} onClose={() => { setOpen(false); reset(); }} title="New creative">
        <form onSubmit={submit} className="space-y-4">
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Name</div>
            <input value={name} onChange={(e) => setName(e.target.value)} className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" placeholder="e.g. Push Ad · Q3 Fintech" />
          </label>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Format</div>
            <select value={format} onChange={(e) => setFormat(e.target.value as CreativeFormat)} className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm capitalize">
              {FORMATS.map((f) => <option key={f} value={f}>{f}</option>)}
            </select>
          </label>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Headline</div>
            <input value={headline} onChange={(e) => setHeadline(e.target.value)} className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" placeholder="Instant KES Loans" />
          </label>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Description</div>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} className="h-20 w-full rounded-lg border border-border bg-card p-3 text-sm" placeholder="Get approved in 60 seconds." />
          </label>
          <label className="block">
            <div className="mb-1.5 text-xs font-medium">Landing URL</div>
            <input value={landingUrl} onChange={(e) => setLandingUrl(e.target.value)} className="h-10 w-full rounded-lg border border-border bg-card px-3 text-sm" placeholder="https://example.com/offer" />
          </label>
          <div>
            <div className="mb-1.5 text-xs font-medium">Image / creative asset</div>
            <label className="flex cursor-pointer items-center gap-3 rounded-lg border border-dashed border-border bg-card p-4 text-sm text-muted-foreground hover:border-primary/40">
              <Upload className="size-4" />
              <span>{imageUrl ? "Change file…" : "Click to upload image"}</span>
              <input type="file" accept="image/*" onChange={onFile} className="hidden" />
            </label>
            {imageUrl && <img src={imageUrl} alt="preview" className="mt-3 h-24 rounded-lg object-cover" />}
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={() => { setOpen(false); reset(); }} className="rounded-full border border-border bg-card px-4 py-2 text-sm font-medium">Cancel</button>
            <button type="submit" className="rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground">Save creative</button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
