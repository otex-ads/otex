import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Pencil, Trash2, Globe, ArrowUpRight } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/PageHeader";
import { SurfaceCard } from "@/components/Card";
import { StatusPill } from "@/components/StatusPill";
import { EmptyState } from "@/components/EmptyState";
import { Modal, FormField, inputCls, selectCls, ModalActions } from "@/components/Modal";
import { api, type Site } from "@/lib/api";
import { USD, Num } from "@/lib/format";

export const Route = createFileRoute("/_app/sites")({
  component: SitesPage,
});

const CATEGORIES = [
  "News",
  "Entertainment",
  "Sports",
  "Tech",
  "Finance",
  "Lifestyle",
  "Gaming",
  "Education",
  "Other",
];

interface SiteForm {
  name: string;
  domain: string;
  category: string;
  description: string;
}

function useSiteMutations() {
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: ["sites"] });
  return {
    create: useMutation({
      mutationFn: (input: SiteForm) => api.createSite(input),
      onSuccess: () => {
        invalidate();
        toast.success("Site created");
      },
      onError: (e: Error) => toast.error(e.message),
    }),
    update: useMutation({
      mutationFn: ({ id, patch }: { id: string; patch: Partial<Site> }) =>
        api.updateSite(id, patch),
      onSuccess: () => {
        invalidate();
        toast.success("Site updated");
      },
      onError: (e: Error) => toast.error(e.message),
    }),
    remove: useMutation({
      mutationFn: (id: string) => api.deleteSite(id),
      onSuccess: () => {
        invalidate();
        toast.success("Site deleted");
      },
      onError: (e: Error) => toast.error(e.message),
    }),
  };
}

function SitesPage() {
  const sites = useQuery<Site[]>({ queryKey: ["sites"], queryFn: api.listSites, retry: false });
  const mut = useSiteMutations();
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Site | null>(null);
  const [deleting, setDeleting] = useState<Site | null>(null);

  const list = sites.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Sites"
        description="Every domain you monetize with PropelAds."
        actions={
          <button
            onClick={() => setCreating(true)}
            className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:brightness-110"
          >
            <Plus className="size-4" /> Add site
          </button>
        }
      />

      <SurfaceCard className="p-0">
        {list.length === 0 ? (
          <EmptyState
            icon={Globe}
            title={sites.isLoading ? "Loading sites…" : "No sites yet"}
            description="Add your first site to generate ad zones and start earning."
            action={
              <button
                onClick={() => setCreating(true)}
                className="inline-flex items-center gap-2 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
              >
                <Plus className="size-4" /> Add site
              </button>
            }
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Site</th>
                  <th className="px-6 py-3 font-medium">Domain</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 text-right font-medium">Impressions</th>
                  <th className="px-6 py-3 text-right font-medium">Revenue</th>
                  <th className="px-6 py-3 text-right font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {list.map((s) => (
                  <tr key={s.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-6 py-4">
                      <Link
                        to="/sites/$siteId"
                        params={{ siteId: s.id }}
                        className="flex items-center gap-2 font-medium text-foreground hover:text-primary"
                      >
                        {s.name}{" "}
                        <ArrowUpRight className="size-3 opacity-0 transition group-hover:opacity-100" />
                      </Link>
                      {s.category && (
                        <div className="text-xs text-muted-foreground">{s.category}</div>
                      )}
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">{s.domain}</td>
                    <td className="px-6 py-4">
                      <StatusPill
                        status={s.status[0].toUpperCase() + s.status.slice(1)}
                        tone={s.status === "active" ? "success" : "warning"}
                      />
                    </td>
                    <td className="num px-6 py-4 text-right">{Num(s.impressions)}</td>
                    <td className="num px-6 py-4 text-right">{USD(s.revenue)}</td>
                    <td className="px-6 py-4">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => setEditing(s)}
                          className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                        >
                          <Pencil className="size-3.5" />
                        </button>
                        <button
                          onClick={() => setDeleting(s)}
                          className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-danger-soft hover:text-danger"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </SurfaceCard>

      {creating && (
        <SiteFormModal
          onClose={() => setCreating(false)}
          onSubmit={(v) => mut.create.mutate(v, { onSuccess: () => setCreating(false) })}
          loading={mut.create.isPending}
          title="Add site"
        />
      )}
      {editing && (
        <SiteFormModal
          site={editing}
          onClose={() => setEditing(null)}
          onSubmit={(v) =>
            mut.update.mutate({ id: editing.id, patch: v }, { onSuccess: () => setEditing(null) })
          }
          loading={mut.update.isPending}
          title="Edit site"
          onToggle={() =>
            mut.update.mutate(
              {
                id: editing.id,
                patch: { status: editing.status === "active" ? "paused" : "active" },
              },
              { onSuccess: () => setEditing(null) },
            )
          }
        />
      )}
      <Modal
        open={!!deleting}
        onClose={() => setDeleting(null)}
        title="Delete site?"
        description={
          deleting
            ? `"${deleting.name}" and all its zones will be removed. This can't be undone.`
            : ""
        }
      >
        <ModalActions
          onCancel={() => setDeleting(null)}
          onConfirm={() =>
            deleting && mut.remove.mutate(deleting.id, { onSuccess: () => setDeleting(null) })
          }
          loading={mut.remove.isPending}
          confirmLabel="Delete"
          confirmVariant="danger"
        />
      </Modal>
    </div>
  );
}

function SiteFormModal({
  site,
  onClose,
  onSubmit,
  loading,
  title,
  onToggle,
}: {
  site?: Site;
  onClose: () => void;
  onSubmit: (v: SiteForm) => void;
  loading?: boolean;
  title: string;
  onToggle?: () => void;
}) {
  const [form, setForm] = useState<SiteForm>({
    name: site?.name ?? "",
    domain: site?.domain ?? "",
    category: site?.category ?? "Other",
    description: site?.description ?? "",
  });

  const submit = () => {
    if (!form.name.trim() || !form.domain.trim()) {
      toast.error("Name and domain are required.");
      return;
    }
    onSubmit(form);
  };

  return (
    <Modal open onClose={onClose} title={title}>
      <div className="flex flex-col gap-4">
        <FormField label="Site name" required>
          <input
            className={inputCls}
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="My Awesome Blog"
          />
        </FormField>
        <FormField label="Domain" required hint="e.g. example.com">
          <input
            className={inputCls}
            value={form.domain}
            onChange={(e) => setForm({ ...form, domain: e.target.value })}
            placeholder="example.com"
          />
        </FormField>
        <FormField label="Category">
          <select
            className={selectCls}
            value={form.category}
            onChange={(e) => setForm({ ...form, category: e.target.value })}
          >
            {CATEGORIES.map((c) => (
              <option key={c}>{c}</option>
            ))}
          </select>
        </FormField>
        <FormField label="Description">
          <textarea
            rows={3}
            className={inputCls + " h-auto py-2"}
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="Short description of the site's audience and content."
          />
        </FormField>
        {site && onToggle && (
          <button
            type="button"
            onClick={onToggle}
            className="self-start rounded-full border border-border bg-card px-4 py-2 text-xs font-medium text-foreground hover:bg-muted/50"
          >
            {site.status === "active" ? "Pause site" : "Resume site"}
          </button>
        )}
      </div>
      <ModalActions
        onCancel={onClose}
        onConfirm={submit}
        loading={loading}
        confirmLabel={site ? "Save changes" : "Create site"}
      />
    </Modal>
  );
}
