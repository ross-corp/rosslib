import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser } from "@/lib/auth";

const features = [
  {
    title: "Collections",
    description:
      "Organize your books your way. Flexible lists with set operations — intersect, union, or diff your labels.",
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
    <>
      {/* Hero */}
      <section className="pt-20 pb-16 text-center">
        <h1 className="text-5xl font-bold text-text-primary tracking-tight leading-tight mb-5">
          Track your reading.
          <br />
          Build your library.
        </h1>
        <p className="text-lg text-text-secondary mb-10 max-w-xl mx-auto leading-relaxed">
          A better home for your books. Flexible collections, community
          discussions, and clean design — built for people who take reading
          seriously.
        </p>
        <div className="flex items-center justify-center gap-3">
          <Link href="/register" className="btn-primary px-7 py-3">
            Get started — it&rsquo;s free
          </Link>
          <Link
            href="/login"
            className="text-text-secondary px-7 py-3 rounded font-medium hover:text-text-primary transition-colors"
          >
            Sign in
          </Link>
        </div>
      </section>

      {/* Divider */}
      <div className="divider" />

      {/* Features */}
      <section className="py-16 grid grid-cols-3 gap-10">
        {features.map((f) => (
          <div key={f.title}>
            <h3 className="section-heading mb-2">{f.title}</h3>
            <p className="text-text-secondary text-sm leading-relaxed">
              {f.description}
            </p>
          </div>
        ))}
      </section>
    </>
  );
}
