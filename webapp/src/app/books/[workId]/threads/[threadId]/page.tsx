import { notFound } from "next/navigation";
import Link from "next/link";
import ThreadComments from "@/components/thread-comments";
import ThreadLockToggle from "@/components/thread-lock-toggle";
import SimilarThreads from "@/components/similar-threads";
import { getUser } from "@/lib/auth";

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
  id: string;
  book: string;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  title: string;
  body: string;
  spoiler: boolean;
  created_at: string;
  locked_at: string | null;
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

  const isModerator = currentUser?.is_moderator ?? false;

  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* Breadcrumb */}
        <div className="text-xs text-text-primary mb-6">
          <Link href={`/books/${workId}`} className="hover:text-text-primary transition-colors">
            {bookTitle ?? "Book"}
          </Link>
          <span className="mx-2">/</span>
          <span>Discussion</span>
        </div>

        {/* Thread header */}
        <article className="mb-10">
          <div className="flex items-start justify-between gap-4 mb-3">
            <h1 className="text-xl font-bold text-text-primary">
              {data.locked_at && (
                <svg className="inline w-4 h-4 text-text-primary mr-1.5 align-text-bottom" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 1a4.5 4.5 0 00-4.5 4.5V9H5a2 2 0 00-2 2v6a2 2 0 002 2h10a2 2 0 002-2v-6a2 2 0 00-2-2h-.5V5.5A4.5 4.5 0 0010 1zm3 8V5.5a3 3 0 10-6 0V9h6z" clipRule="evenodd" />
                </svg>
              )}
              {data.spoiler && (
                <span className="text-xs font-medium text-amber-600 border border-amber-200 rounded px-1.5 py-0.5 mr-2 leading-none align-middle">
                  Spoiler
                </span>
              )}
              {data.title}
            </h1>
            {isModerator && (
              <ThreadLockToggle
                threadId={threadId}
                initialLockedAt={data.locked_at}
              />
            )}
          </div>

          <div className="flex items-center gap-3 mb-4">
            <Link href={`/${data.username}`} className="shrink-0">
              {data.avatar_url ? (
                <img
                  src={data.avatar_url}
                  alt={data.display_name ?? data.username}
                  className="w-8 h-8 rounded-full object-cover"
                />
              ) : (
                <div className="w-8 h-8 rounded-full bg-surface-2" />
              )}
            </Link>
            <div>
              <Link
                href={`/${data.username}`}
                className="text-sm font-medium text-text-primary hover:underline"
              >
                {data.display_name ?? data.username}
              </Link>
              <p className="text-xs text-text-primary">
                {formatDate(data.created_at)}
              </p>
            </div>
          </div>

          {data.spoiler ? (
            <details>
              <summary className="text-xs text-text-primary cursor-pointer select-none hover:text-text-primary transition-colors">
                Show thread body (contains spoilers)
              </summary>
              <p className="mt-3 text-sm text-text-primary leading-relaxed whitespace-pre-wrap">
                {data.body}
              </p>
            </details>
          ) : (
            <p className="text-sm text-text-primary leading-relaxed whitespace-pre-wrap">
              {data.body}
            </p>
          )}
        </article>

        {/* Comments */}
        <section className="border-t border-border pt-8">
          <ThreadComments
            threadId={threadId}
            initialComments={data.comments}
            isLoggedIn={!!currentUser}
            currentUserId={currentUser?.user_id ?? null}
            isLocked={!!data.locked_at}
          />
        </section>

        {/* Similar threads */}
        <SimilarThreads threadId={threadId} workId={workId} />
      </main>
    </div>
  );
}
