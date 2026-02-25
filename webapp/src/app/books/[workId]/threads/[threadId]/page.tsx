import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import ThreadComments from "@/components/thread-comments";
import SimilarThreads from "@/components/similar-threads";
import { getUser } from "@/lib/auth";

type Thread = {
  id: string;
  book_id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  title: string;
  body: string;
  spoiler: boolean;
  created_at: string;
  comment_count: number;
};

type Comment = {
  id: string;
  thread_id: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  parent_id: string | null;
  body: string;
  created_at: string;
};

type ThreadDetail = {
  thread: Thread;
  comments: Comment[];
};

async function fetchThread(threadId: string): Promise<ThreadDetail | null> {
  const res = await fetch(`${process.env.API_URL}/threads/${threadId}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchBookTitle(workId: string): Promise<string | null> {
  const res = await fetch(`${process.env.API_URL}/books/${workId}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  const data = await res.json();
  return data.title;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });
}

export default async function ThreadPage({
  params,
}: {
  params: Promise<{ workId: string; threadId: string }>;
}) {
  const { workId, threadId } = await params;
  const currentUser = await getUser();

  const [data, bookTitle] = await Promise.all([
    fetchThread(threadId),
    fetchBookTitle(workId),
  ]);

  if (!data) notFound();

  const { thread, comments } = data;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* Breadcrumb */}
        <div className="text-xs text-stone-400 mb-6">
          <Link href={`/books/${workId}`} className="hover:text-stone-600 transition-colors">
            {bookTitle ?? "Book"}
          </Link>
          <span className="mx-2">/</span>
          <span>Discussion</span>
        </div>

        {/* Thread header */}
        <article className="mb-10">
          <h1 className="text-xl font-bold text-stone-900 mb-3">
            {thread.spoiler && (
              <span className="text-xs font-medium text-amber-600 border border-amber-200 rounded px-1.5 py-0.5 mr-2 leading-none align-middle">
                Spoiler
              </span>
            )}
            {thread.title}
          </h1>

          <div className="flex items-center gap-3 mb-4">
            <Link href={`/${thread.username}`} className="shrink-0">
              {thread.avatar_url ? (
                <img
                  src={thread.avatar_url}
                  alt={thread.display_name ?? thread.username}
                  className="w-8 h-8 rounded-full object-cover"
                />
              ) : (
                <div className="w-8 h-8 rounded-full bg-stone-200" />
              )}
            </Link>
            <div>
              <Link
                href={`/${thread.username}`}
                className="text-sm font-medium text-stone-900 hover:underline"
              >
                {thread.display_name ?? thread.username}
              </Link>
              <p className="text-xs text-stone-400">
                {formatDate(thread.created_at)}
              </p>
            </div>
          </div>

          {thread.spoiler ? (
            <details>
              <summary className="text-xs text-stone-400 cursor-pointer select-none hover:text-stone-600 transition-colors">
                Show thread body (contains spoilers)
              </summary>
              <p className="mt-3 text-sm text-stone-700 leading-relaxed whitespace-pre-wrap">
                {thread.body}
              </p>
            </details>
          ) : (
            <p className="text-sm text-stone-700 leading-relaxed whitespace-pre-wrap">
              {thread.body}
            </p>
          )}
        </article>

        {/* Comments */}
        <section className="border-t border-stone-100 pt-8">
          <ThreadComments
            threadId={threadId}
            initialComments={comments}
            isLoggedIn={!!currentUser}
            currentUserId={currentUser?.user_id ?? null}
          />
        </section>

        {/* Similar threads */}
        <SimilarThreads threadId={threadId} workId={workId} />
      </main>
    </div>
  );
}
