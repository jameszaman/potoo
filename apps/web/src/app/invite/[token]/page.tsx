"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { getInvite, acceptInvite } from "@/lib/auth";

type Step = "org" | "user";

export default function InvitePage() {
  const { token } = useParams<{ token: string }>();
  const router = useRouter();

  const [orgName, setOrgName] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [inviteError, setInviteError] = useState("");
  const [step, setStep] = useState<Step>("org");
  const [animating, setAnimating] = useState(false);
  const [form, setForm] = useState({ first_name: "", last_name: "", phone: "", email: "", password: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    getInvite(token)
      .then((data) => {
        setOrgName(data.org_name);
        setExpiresAt(new Date(data.expires_at).toLocaleDateString());
      })
      .catch((e) => setInviteError(String(e)));
  }, [token]);

  const set = (field: string) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const goToUser = (e: React.FormEvent) => {
    e.preventDefault();
    setAnimating(true);
    setTimeout(() => { setStep("user"); setAnimating(false); }, 300);
  };

  const goToOrg = () => {
    setAnimating(true);
    setTimeout(() => { setStep("org"); setAnimating(false); }, 300);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await acceptInvite({
        token,
        first_name: form.first_name,
        last_name: form.last_name,
        phone: form.phone || undefined,
        email: form.email,
        password: form.password,
      });
      router.push("/sign-in");
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const inputClass =
    "w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900";
  const labelClass = "block text-xs font-medium text-gray-700 mb-1";

  if (inviteError) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white border border-gray-200 rounded-lg p-8 w-full max-w-sm shadow-sm text-center">
          <h1 className="text-lg font-semibold mb-2">Invite unavailable</h1>
          <p className="text-sm text-red-600">{inviteError}</p>
        </div>
      </div>
    );
  }

  if (!orgName) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <p className="text-sm text-gray-400">Checking invite…</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12">
      <div className="bg-white border border-gray-200 rounded-lg shadow-sm w-full max-w-md overflow-hidden">
        {/* Progress bar */}
        <div className="h-1 bg-gray-100">
          <div
            className="h-1 bg-gray-900 transition-all duration-500"
            style={{ width: step === "org" ? "50%" : "100%" }}
          />
        </div>

        <div className="p-8">
          {/* Step indicator */}
          <div className="flex items-center gap-2 mb-6">
            <div className="flex items-center gap-1.5">
              <span className={`w-5 h-5 rounded-full text-xs flex items-center justify-center font-medium ${step === "org" ? "bg-gray-900 text-white" : "bg-gray-200 text-gray-500"}`}>1</span>
              <span className={`text-xs font-medium ${step === "org" ? "text-gray-900" : "text-gray-400"}`}>Your invitation</span>
            </div>
            <div className="h-px flex-1 bg-gray-200" />
            <div className="flex items-center gap-1.5">
              <span className={`w-5 h-5 rounded-full text-xs flex items-center justify-center font-medium ${step === "user" ? "bg-gray-900 text-white" : "bg-gray-200 text-gray-500"}`}>2</span>
              <span className={`text-xs font-medium ${step === "user" ? "text-gray-900" : "text-gray-400"}`}>Create account</span>
            </div>
          </div>

          <div
            className="transition-all duration-300"
            style={{ opacity: animating ? 0 : 1, transform: animating ? "translateX(20px)" : "translateX(0)" }}
          >
            {step === "org" && (
              <>
                <h1 className="text-lg font-semibold mb-1">You&apos;ve been invited</h1>
                <p className="text-sm text-gray-500 mb-6">Review your invitation before creating an account.</p>

                <div className="bg-gray-50 border border-gray-200 rounded-md p-4 mb-6 space-y-2">
                  <div>
                    <p className="text-xs text-gray-400">Organization</p>
                    <p className="text-sm font-medium text-gray-900">{orgName}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-400">Expires</p>
                    <p className="text-sm text-gray-700">{expiresAt}</p>
                  </div>
                </div>

                <form onSubmit={goToUser}>
                  <button
                    type="submit"
                    className="w-full bg-gray-900 text-white text-sm font-medium py-2 rounded-md hover:bg-gray-700 transition-colors"
                  >
                    Accept invitation →
                  </button>
                </form>
              </>
            )}

            {step === "user" && (
              <>
                <h1 className="text-lg font-semibold mb-1">Create your account</h1>
                <p className="text-sm text-gray-500 mb-6">
                  You&apos;ll join{" "}
                  <span className="font-medium text-gray-800">{orgName}</span> as a member.
                </p>

                {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

                <form onSubmit={handleSubmit} className="space-y-4">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className={labelClass}>First name</label>
                      <input value={form.first_name} onChange={set("first_name")} placeholder="Morgan" required className={inputClass} />
                    </div>
                    <div>
                      <label className={labelClass}>Last name</label>
                      <input value={form.last_name} onChange={set("last_name")} placeholder="Taylor" required className={inputClass} />
                    </div>
                  </div>
                  <div>
                    <label className={labelClass}>Phone <span className="text-gray-400">(optional)</span></label>
                    <input value={form.phone} onChange={set("phone")} placeholder="+1 555 000 0000" className={inputClass} />
                  </div>
                  <div>
                    <label className={labelClass}>Email</label>
                    <input type="email" value={form.email} onChange={set("email")} placeholder="you@example.com" required className={inputClass} />
                  </div>
                  <div>
                    <label className={labelClass}>Password</label>
                    <input type="password" value={form.password} onChange={set("password")} placeholder="At least 8 characters" required minLength={8} className={inputClass} />
                  </div>
                  <div className="flex gap-3">
                    <button
                      type="button"
                      onClick={goToOrg}
                      className="flex-1 border border-gray-300 text-gray-700 text-sm font-medium py-2 rounded-md hover:bg-gray-50 transition-colors"
                    >
                      ← Back
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="flex-1 bg-gray-900 text-white text-sm font-medium py-2 rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50"
                    >
                      {loading ? "Joining…" : "Create account"}
                    </button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
