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
    <div className="min-h-screen flex flex-col">
      <Nav />

      {/* Hero */}
      <section className="max-w-5xl mx-auto px-4 sm:px-6 pt-24 pb-20 text-center flex-1 flex flex-col justify-center">
        <h1 className="text-5xl sm:text-7xl font-bold tracking-tight leading-tight mb-8 uppercase">
          Track your reading.
          <br />
          Build your library.
        </h1>
        <p className="text-xl text-stone-600 mb-12 max-w-xl mx-auto leading-relaxed font-mono">
          A better home for your books. Flexible collections, community
          discussions, and clean design — built for people who take reading
          seriously.
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-6">
          <Link
            href="/register"
            className="bg-black text-white border-2 border-black px-8 py-4 font-bold uppercase hover:bg-white hover:text-black transition-all shadow-[6px_6px_0_0_#000] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[4px_4px_0_0_#000] active:translate-x-[6px] active:translate-y-[6px] active:shadow-none w-full sm:w-auto"
          >
            Get started — it&rsquo;s free
          </Link>
          <Link
            href="/login"
            className="bg-white text-black border-2 border-black px-8 py-4 font-bold uppercase hover:bg-black hover:text-white transition-all shadow-[6px_6px_0_0_#000] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[4px_4px_0_0_#000] active:translate-x-[6px] active:translate-y-[6px] active:shadow-none w-full sm:w-auto"
          >
            Sign in
          </Link>
        </div>
      </section>

      {/* Divider */}
      <div className="max-w-5xl mx-auto px-4 sm:px-6 w-full">
        <div className="border-t-2 border-black" />
      </div>

      {/* Features */}
      <section className="max-w-5xl mx-auto px-4 sm:px-6 py-20 grid grid-cols-1 sm:grid-cols-3 gap-8">
        {features.map((f) => (
          <div key={f.title} className="border-2 border-black p-6 bg-white shadow-[8px_8px_0_0_#000]">
            <h3 className="font-bold text-xl mb-4 uppercase border-b-2 border-black pb-2 inline-block">{f.title}</h3>
            <p className="text-stone-600 text-sm leading-relaxed font-mono">
              {f.description}
            </p>
          </div>
        ))}
      </section>

      {/* Footer */}
      <footer className="border-t-2 border-black bg-white mt-auto">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8 flex items-center justify-between text-xs font-bold uppercase tracking-wider">
          <span>rosslib // EST. 2025</span>
          <span>Better than Goodreads.</span>
        </div>
      </footer>
    </div>
  );
}
