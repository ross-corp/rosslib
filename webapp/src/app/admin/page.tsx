import { redirect } from "next/navigation";
import AdminUserList from "@/components/admin-user-list";
import AdminLinkEdits from "@/components/admin-link-edits";
import { getUser } from "@/lib/auth";

export default async function AdminPage() {
  const user = await getUser();
  if (!user || !user.is_moderator) redirect("/");

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-text-primary mb-8">Admin</h1>

        <section className="mb-12">
          <h2 className="text-lg font-semibold text-text-primary mb-4">
            Link Edit Queue
          </h2>
          <p className="text-sm text-text-primary mb-6">
            Review proposed edits to community book links. Approved edits are
            applied immediately.
          </p>
          <AdminLinkEdits />
        </section>

        <section>
          <h2 className="text-lg font-semibold text-text-primary mb-4">
            Manage Moderators
          </h2>
          <p className="text-sm text-text-primary mb-6">
            Grant or revoke moderator status. Changes take effect on the
            user&apos;s next login.
          </p>
          <AdminUserList />
        </section>
      </main>
    </div>
  );
}
