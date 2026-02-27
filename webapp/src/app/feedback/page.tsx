import { redirect } from "next/navigation";
import FeedbackForm from "@/components/feedback-form";
import MyFeedback from "@/components/my-feedback";
import { getUser } from "@/lib/auth";

export default async function FeedbackPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-text-primary mb-2">Feedback</h1>
        <p className="text-sm text-text-secondary mb-8">
          Found a bug or have an idea? Let us know.
        </p>
        <FeedbackForm />
        <div className="mt-10">
          <MyFeedback />
        </div>
      </main>
    </div>
  );
}
