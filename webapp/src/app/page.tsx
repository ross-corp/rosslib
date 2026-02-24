import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import { getUser } from "@/lib/auth";

const features = [
  {
    title: "Collections",
    description:
      "Organize your books your way. Flexible lists with set operations — intersect, union, or diff your shelves.",
  },
  {
    title: "Social",
    description:
      "Follow readers you trust. See what they're reading and share your own library with the people who matter.",
  },
  {
    title: "Discussions",
    description:
      "Threaded conversations on every book page. Spoiler flags, honest reviews, no algorithmic noise.",
  },
];

export default async function Home() {
  const user = await getUser();
  if (user) redirect("/feed");
  return (
    <div className="min-h-screen">
      <Nav />

      {/* Hero */}
      <section className="max-w-5xl mx-auto px-4 sm:px-6 pt-24 pb-20 text-center">
        <h1 className="text-5xl sm:text-6xl font-bold text-stone-900 tracking-tight leading-tight mb-5">
          Track your reading.
          <br />
          Build your library.
        </h1>
        <p className="text-xl text-stone-500 mb-10 max-w-xl mx-auto leading-relaxed">
          A better home for your books. Flexible collections, community
          discussions, and clean design — built for people who take reading
          seriously.
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
          <Link
            href="/register"
            className="bg-stone-900 text-white px-7 py-3 rounded font-medium hover:bg-stone-700 transition-colors w-full sm:w-auto"
          >
            Get started — it&rsquo;s free
          </Link>
          <Link
            href="/login"
            className="text-stone-600 px-7 py-3 rounded font-medium hover:text-stone-900 transition-colors w-full sm:w-auto"
          >
            Sign in
          </Link>
        </div>
      </section>

      {/* Divider */}
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        <div className="border-t border-stone-200" />
      </div>

      {/* Features */}
      <section className="max-w-5xl mx-auto px-4 sm:px-6 py-20 grid grid-cols-1 sm:grid-cols-3 gap-10">
        {features.map((f) => (
          <div key={f.title}>
            <h3 className="font-semibold text-stone-900 mb-2">{f.title}</h3>
            <p className="text-stone-500 text-sm leading-relaxed">
              {f.description}
            </p>
          </div>
        ))}
      </section>

      {/* Footer */}
      <footer className="border-t border-stone-200">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-6 flex items-center justify-between text-xs text-stone-400">
          <span>rosslib</span>
          <span>Better than Goodreads.</span>
        </div>
      </footer>
    </div>
  );
}
