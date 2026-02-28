import { redirect } from "next/navigation";
import Link from "next/link";
import APITokensForm from "@/components/api-tokens-form";
import SettingsNav from "@/components/settings-nav";
import { getUser } from "@/lib/auth";

export default async function APITokensPage() {
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

        <h2 className="text-xl font-semibold text-text-primary mb-2">API Tokens</h2>
        <p className="text-sm text-text-secondary mb-6">
          Create API tokens for external integrations like CLI tools or Calibre.
          Tokens authenticate as your account via the <code className="bg-surface-2 px-1 rounded">Authorization: Bearer &lt;token&gt;</code> header.
        </p>

        <APITokensForm />
      </main>
    </div>
  );
}
