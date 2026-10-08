"use client";

import { motion } from "framer-motion";
import { ArrowLeft, ArrowRight, Check, Fingerprint, Sparkles } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { registerStudent } from "@/lib/api";

const fieldClass = "input";

export default function RegisterPage() {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [step, setStep] = useState(0);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setBusy(true);
    const form = new FormData(event.currentTarget);
    const likes = String(form.get("likes") ?? "").split(",").map((item) => item.trim()).filter(Boolean);
    const dislikes = String(form.get("dislikes") ?? "").split(",").map((item) => item.trim()).filter(Boolean);
    try {
      const student = await registerStudent({
        name: String(form.get("name") ?? "").trim(),
        email: String(form.get("email") ?? "").trim().toLowerCase(),
        password: String(form.get("password") ?? ""),
        age: Number(form.get("age")),
        height: String(form.get("height") ?? "").trim(),
        department: String(form.get("department") ?? "").trim(),
        class: String(form.get("class") ?? "").trim(),
        gender: String(form.get("gender") ?? ""),
        orientation: String(form.get("orientation") ?? ""),
        verificationProof: String(form.get("verificationProof") ?? "").trim(),
        preferences: {
          targetGender: String(form.get("targetGender") ?? ""),
          orientation: String(form.get("targetOrientation") ?? ""),
        },
        likes,
        dislikes,
      });
      localStorage.setItem("sideby.studentId", student.id);
      router.push("/dashboard");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Registration failed. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="shell min-h-screen px-5 py-6 md:px-10 md:py-9">
      <header className="mx-auto flex max-w-6xl items-center justify-between">
        <Link href="/" className="flex items-center gap-3" aria-label="Sideby home">
          <span className="grid h-9 w-9 place-items-center border border-[var(--line)] bg-white/[.04] text-[var(--acid)]"><Sparkles size={16} /></span>
          <span className="font-display text-lg tracking-[.01em]">sideby<span className="text-[var(--acid)]">.</span></span>
        </Link>
        <Link href="/login" className="flex items-center gap-2 text-sm text-[#aab2a8] transition hover:text-white"><ArrowLeft size={15} /> Back to sign in</Link>
      </header>

      <div className="mx-auto grid max-w-6xl gap-10 pb-10 pt-12 lg:grid-cols-[.78fr_1.22fr] lg:items-start lg:gap-16 lg:pt-20">
        <motion.section initial={{ opacity: 0, x: -18 }} animate={{ opacity: 1, x: 0 }} transition={{ duration: .45 }} className="pt-2">
          <p className="eyebrow mb-5 flex items-center gap-2"><span className="status-dot" /> A campus, a little closer</p>
          <h1 className="display max-w-xl text-5xl leading-[1.02] sm:text-6xl">Good people<br /><i className="text-[var(--acid)]">find their way.</i></h1>
          <p className="mt-6 max-w-md text-[15px] leading-7 text-[#aab2a8]">Set up your student profile and meet someone who shares your corner of campus.</p>
          <div className="mt-12 hidden max-w-md overflow-hidden border border-white/10 bg-[#202720] md:block">
            <div className="relative h-[198px] bg-cover bg-center" style={{ backgroundImage: "linear-gradient(90deg, rgba(17,22,17,.86), rgba(17,22,17,.08)), url('https://images.unsplash.com/photo-1529156069898-49953e39b3ac?auto=format&fit=crop&w=1100&q=85')" }}>
              <div className="absolute inset-0 bg-gradient-to-t from-[#111611]/90 via-transparent to-transparent" />
              <div className="absolute bottom-5 left-5"><p className="eyebrow text-white/65">Find your people</p><p className="display mt-1 text-2xl">More than a first hello.</p></div>
              <div className="absolute right-5 top-5 border border-white/25 bg-black/25 px-3 py-2 text-[10px] uppercase tracking-[.13em] text-white/85 backdrop-blur">Student community</div>
            </div>
          </div>
        </motion.section>

        <motion.section initial={{ opacity: 0, y: 18 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: .5, delay: .08 }} className="glass p-5 sm:p-8">
          <div className="mb-7 flex items-start justify-between gap-5">
            <div><p className="eyebrow">Your profile / 0{step + 1}</p><h2 className="mt-2 font-display text-2xl">{step === 0 ? "The essentials" : "Your kind of person"}</h2></div>
            <div className="flex gap-1.5 pt-1" aria-label={`Step ${step + 1} of 2`}>
              {[0, 1].map((item) => <span key={item} className={`h-1 w-8 ${item <= step ? "bg-[var(--acid)]" : "bg-white/15"}`} />)}
            </div>
          </div>
          <form onSubmit={submit}>
            <div className={step === 0 ? "grid gap-x-4 gap-y-4 sm:grid-cols-2" : "hidden"}>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 block">Full name</span><input className={fieldClass} name="name" autoComplete="name" placeholder="How should we call you?" required={step === 0} /></label>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 block">Student email</span><input className={fieldClass} name="email" type="email" autoComplete="email" placeholder="you@campus.edu" required={step === 0} /></label>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 block">Password</span><input className={fieldClass} name="password" type="password" autoComplete="new-password" placeholder="Choose a secure password" required={step === 0} /></label>
              <label><span className="eyebrow mb-2 block">Age</span><input className={fieldClass} name="age" type="number" min="16" max="99" placeholder="20" required={step === 0} /></label>
              <label><span className="eyebrow mb-2 block">Height</span><input className={fieldClass} name="height" placeholder="170 cm" required={step === 0} /></label>
              <label><span className="eyebrow mb-2 block">Department</span><input className={fieldClass} name="department" placeholder="Computer science" required={step === 0} /></label>
              <label><span className="eyebrow mb-2 block">Class / year</span><input className={fieldClass} name="class" placeholder="Class of 2028" required={step === 0} /></label>
              <label><span className="eyebrow mb-2 block">Gender</span><select className={fieldClass} name="gender" defaultValue="" required={step === 0}><option value="" disabled>Select</option><option value="woman">Woman</option><option value="man">Man</option><option value="nonbinary">Non-binary</option><option value="self-described">Self-described</option></select></label>
              <label><span className="eyebrow mb-2 block">Orientation</span><select className={fieldClass} name="orientation" defaultValue="" required={step === 0}><option value="" disabled>Select</option><option value="straight">Straight</option><option value="gay">Gay</option><option value="lesbian">Lesbian</option><option value="bisexual">Bisexual</option><option value="pansexual">Pansexual</option><option value="queer">Queer</option><option value="other">Other</option></select></label>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 flex items-center gap-2"><Fingerprint size={13} /> Verification proof</span><input className={fieldClass} name="verificationProof" placeholder="Student email, URL, or verification hash" required={step === 0} /></label>
            </div>
            <div className={step === 1 ? "grid gap-x-4 gap-y-4 sm:grid-cols-2" : "hidden"}>
              <label><span className="eyebrow mb-2 block">Looking for</span><select className={fieldClass} name="targetGender" defaultValue="any" required={step === 1}><option value="any">Anyone</option><option value="woman">Women</option><option value="man">Men</option><option value="nonbinary">Non-binary students</option></select></label>
              <label><span className="eyebrow mb-2 block">Orientation preference</span><select className={fieldClass} name="targetOrientation" defaultValue="any" required={step === 1}><option value="any">Any orientation</option><option value="straight">Straight</option><option value="gay">Gay</option><option value="lesbian">Lesbian</option><option value="bisexual">Bisexual</option><option value="pansexual">Pansexual</option><option value="queer">Queer</option><option value="other">Other</option></select></label>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 block">Things you like</span><input className={fieldClass} name="likes" placeholder="film photography, matcha, climbing" /><span className="mt-2 block text-xs text-[#7f897f]">Separate a few with commas.</span></label>
              <label className="sm:col-span-2"><span className="eyebrow mb-2 block">Not your thing</span><input className={fieldClass} name="dislikes" placeholder="crowds, early mornings" /></label>
            </div>
            {error && <p role="alert" className="mt-4 border-l-2 border-[var(--clay)] bg-[rgba(228,138,104,.08)] px-3 py-2 text-sm text-[#f0ad93]">{error}</p>}
            <div className="mt-7 flex justify-between gap-3 border-t border-white/10 pt-5">
              {step === 1 ? <button type="button" className="ghost-button" onClick={() => setStep(0)}><ArrowLeft size={15} /> Previous</button> : <span className="flex items-center gap-2 text-xs text-[#818b81]"><Check size={14} className="text-[var(--acid)]" /> Your data stays yours</span>}
              {step === 0 ? <button type="button" className="primary-button" onClick={() => setStep(1)}>Continue <ArrowRight size={15} /></button> : <button type="submit" disabled={busy} className="primary-button">{busy ? "Creating profile…" : "Join Sideby"}<ArrowRight size={15} /></button>}
            </div>
          </form>
        </motion.section>
      </div>
    </main>
  );
}
