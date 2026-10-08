"use client";

import { motion } from "framer-motion";
import { ArrowRight, Sparkles } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { loginStudent } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setBusy(true);

    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "").trim().toLowerCase();
    const password = String(form.get("password") ?? "");

    try {
      const student = await loginStudent(email, password);
      localStorage.setItem("sideby.studentId", student.id);
      router.push("/dashboard");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Login failed. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="shell min-h-screen px-5 py-8 md:px-10">
      <header className="mx-auto flex max-w-5xl items-center justify-between">
        <Link href="/" className="flex items-center gap-3" aria-label="Sideby home">
          <span className="grid h-9 w-9 place-items-center border border-[var(--line)] bg-white/[.04] text-[var(--acid)]"><Sparkles size={16} /></span>
          <span className="font-display text-lg tracking-[.01em]">sideby<span className="text-[var(--acid)]">.</span></span>
        </Link>
      </header>

      <div className="mx-auto mt-14 max-w-md rounded-2xl border border-white/10 bg-[#121914]/80 p-6 shadow-[0_16px_50px_rgba(0,0,0,0.22)] backdrop-blur-sm sm:p-8">
        <p className="eyebrow mb-3">Welcome back</p>
        <h1 className="font-display text-3xl text-white">Sign in to continue</h1>
        <p className="mt-3 text-sm text-[#aab2a8]">Use your student email and the password you chose during signup.</p>

        <form onSubmit={submit} className="mt-7 space-y-4">
          <label>
            <span className="eyebrow mb-2 block">Student email</span>
            <input className="input" name="email" type="email" autoComplete="email" placeholder="you@campus.edu" required />
          </label>
          <label>
            <span className="eyebrow mb-2 block">Password</span>
            <input className="input" name="password" type="password" autoComplete="current-password" placeholder="Enter your password" required />
          </label>

          {error && <p role="alert" className="border-l-2 border-[var(--clay)] bg-[rgba(228,138,104,.08)] px-3 py-2 text-sm text-[#f0ad93]">{error}</p>}

          <button type="submit" disabled={busy} className="primary-button mt-2 w-full justify-center">
            {busy ? "Signing in..." : "Sign in"}
            <ArrowRight size={15} />
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-[#aab2a8]">
          New here? <Link href="/register" className="text-[var(--acid)] hover:underline">Create an account</Link>
        </p>
      </div>
    </main>
  );
}
