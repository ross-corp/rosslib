import { redirect } from "next/navigation";
import Nav from "@/components/nav";
import AdminUserList from "@/components/admin-user-list";
import { getUser } from "@/lib/auth";

export default async function AdminPage() {
  const user = await getUser();
  if (!user || !user.is_moderator) redirect("/");

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-stone-900 mb-8">Admin</h1>
        <section>
          <h2 className="text-lg font-semibold text-stone-800 mb-4">
            Manage Moderators
          </h2>
          <p className="text-sm text-stone-500 mb-6">
            Grant or revoke moderator status. Changes take effect on the
            user&apos;s next login.
          </p>
          <AdminUserList />
        </section>
      </main>
    </div>
  );
}
