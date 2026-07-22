import { createFileRoute } from "@tanstack/react-router";
import { SurfaceCard } from "@/components/Card";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";

export const Route = createFileRoute("/_app/users")({
  component: UsersPage,
});

function UsersPage() {
  const { data: users, isLoading } = useQuery({
    queryKey: ["admin-users"],
    queryFn: () => api.fetch("/admin/users"),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Users</h1>
        <p className="mt-2 text-sm text-muted-foreground">Manage publishers and advertisers</p>
      </div>
      <SurfaceCard className="p-0">
        {isLoading ? (
          <div className="p-8 text-center text-sm text-muted-foreground">Loading...</div>
        ) : users && users.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border bg-muted/40 text-left text-[11px] uppercase tracking-wider text-muted-foreground">
                  <th className="px-6 py-3 font-medium">Email</th>
                  <th className="px-6 py-3 font-medium">Type</th>
                  <th className="px-6 py-3 font-medium">Company</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user: any) => (
                  <tr key={user.id} className="border-b border-border hover:bg-muted/30">
                    <td className="px-6 py-4 font-medium">{user.email}</td>
                    <td className="px-6 py-4 capitalize">{user.type}</td>
                    <td className="px-6 py-4">{user.company_name || "-"}</td>
                    <td className="px-6 py-4">
                      <StatusPill status={user.status} />
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">
                      {new Date(user.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-sm text-muted-foreground">No users found</div>
        )}
      </SurfaceCard>
    </div>
  );
}
