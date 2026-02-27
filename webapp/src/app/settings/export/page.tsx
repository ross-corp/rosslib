import { redirect } from "next/navigation";
import Link from "next/link";
import ExportForm from "@/components/export-form";
import SettingsNav from "@/components/settings-nav";
import { getUser } from "@/lib/auth";

export default async function ExportPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-text-primary">
          <Link href={`/${user.username}`} className="hover:text-text-primary transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <span className="text-text-primary">Settings</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-4">Settings</h1>

        <SettingsNav />

        <h2 className="text-xl font-semibold text-text-primary mb-2">Export to CSV</h2>
        <p className="text-sm text-text-primary mb-8">
          Download your entire library as a CSV file.
        </p>

        <ExportForm />
      </main>
    </div>
  );
}
