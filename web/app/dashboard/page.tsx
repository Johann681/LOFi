"use client";

import { AnimatePresence, motion } from "framer-motion";
import {
  ArrowRight, ArrowUpRight, Check, CircleHelp, Compass,
  Heart, LoaderCircle, LogOut, MessageCircle, MoreHorizontal, RefreshCw, Send,
  Settings2, Sparkles, Users, X,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { chatSocketUrl, findMatch, getMatchHistory, getMessages, getStudent } from "@/lib/api";
import type { ChatMessage, MatchHistoryItem, MatchResult, Student, WebSocketEvent } from "@/lib/types";

type SocketState = "connecting" | "online" | "offline";

function initials(name: string) {
  return name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join("").toUpperCase() || "S";
}

function timeLabel(value: string) {
  return new Intl.DateTimeFormat(undefined, { hour: "numeric", minute: "2-digit" }).format(new Date(value));
}

export default function DashboardPage() {
  const router = useRouter();
  const [studentId, setStudentId] = useState("");
  const [profile, setProfile] = useState<Student | null>(null);
  const [history, setHistory] = useState<MatchHistoryItem[]>([]);
  const [active, setActive] = useState<MatchHistoryItem | null>(null);
  const [result, setResult] = useState<MatchResult | null>(null);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState("");
  const [socketState, setSocketState] = useState<SocketState>("connecting");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [chatError, setChatError] = useState("");
  const socketRef = useRef<WebSocket | null>(null);
  const activeIdRef = useRef("");
  const endOfMessagesRef = useRef<HTMLDivElement>(null);
  const [mobileChatOpen, setMobileChatOpen] = useState(false);

  useEffect(() => {
    const savedId = localStorage.getItem("sideby.studentId");
    if (!savedId) {
      router.replace("/login");
      return;
    }
    setStudentId(savedId);
  }, [router]);

  const refreshHistory = useCallback(async () => {
    if (!studentId) return [] as MatchHistoryItem[];
    const items = await getMatchHistory(studentId);
    setHistory(items);
    return items;
  }, [studentId]);

  useEffect(() => {
    if (!studentId) return;
    let alive = true;
    getStudent(studentId).then((student) => { if (alive) setProfile(student); }).catch((requestError) => {
      if (!alive) return;
      setError(requestError instanceof Error ? requestError.message : "Could not load your profile.");
    });
    refreshHistory().catch((requestError) => {
      if (alive) setError(requestError instanceof Error ? requestError.message : "Could not load your matches.");
    });
    return () => { alive = false; };
  }, [studentId, refreshHistory]);

  useEffect(() => {
    if (!studentId) return;
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
    let shouldReconnect = true;

    const connect = () => {
      setSocketState("connecting");
      const socket = new WebSocket(chatSocketUrl(studentId));
      socketRef.current = socket;
      socket.onopen = () => setSocketState("online");
      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(String(event.data)) as WebSocketEvent;
          if (payload.type === "CHAT_MESSAGE") {
            if (payload.message.matchId === activeIdRef.current) {
              setMessages((previous) => previous.some((item) => item.id === payload.message.id) ? previous : [...previous, payload.message]);
            }
          } else if (payload.type === "MATCH_FOUND") {
            setResult({ status: "matched", matchId: payload.data.matchId });
            setSearching(false);
            refreshHistory().then((items) => {
              const found = items.find((item) => item.match.id === payload.data.matchId);
              if (found) setActive(found);
            }).catch(() => undefined);
          } else if (payload.type === "ERROR") {
            setChatError(payload.error);
          }
        } catch {
          setChatError("Received an unreadable message from the chat server.");
        }
      };
      socket.onclose = () => {
        if (!shouldReconnect) return;
        setSocketState("offline");
        reconnectTimer = setTimeout(connect, 1800);
      };
      socket.onerror = () => setSocketState("offline");
    };

    connect();
    return () => {
      shouldReconnect = false;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [studentId, refreshHistory]);

  useEffect(() => {
    activeIdRef.current = active?.match.id ?? "";
    if (!active || !studentId) {
      setMessages([]);
      return;
    }
    let alive = true;
    setMessages([]);
    getMessages(active.match.id, studentId).then((loaded) => {
      if (alive) setMessages(loaded);
    }).catch((requestError) => {
      if (alive) setChatError(requestError instanceof Error ? requestError.message : "Could not load chat history.");
    });
    return () => { alive = false; };
  }, [active?.match.id, studentId]);

  useEffect(() => {
    endOfMessagesRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages]);

  async function handleFindMatch() {
    if (!studentId) return;
    setError("");
    setSearching(true);
    setResult(null);
    try {
      const response = await findMatch(studentId);
      setResult(response);
      if (response.status === "matched") {
        setSearching(false);
        const items = await refreshHistory();
        const found = items.find((item) => item.match.id === response.matchId);
        if (found) setActive(found);
        else if (response.student && response.matchId) {
          setActive({ match: { id: response.matchId, student1Id: studentId, student2Id: response.student.id, status: "active", createdAt: new Date().toISOString() }, student: response.student });
        }
      }
    } catch (requestError) {
      setSearching(false);
      setError(requestError instanceof Error ? requestError.message : "Could not search for a match.");
    }
  }

  function sendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const socket = socketRef.current;
    const content = draft.trim();
    if (!content || !active || !socket || socket.readyState !== WebSocket.OPEN) {
      if (socket?.readyState !== WebSocket.OPEN) setChatError("Chat is reconnecting. Try sending again in a moment.");
      return;
    }
    socket.send(JSON.stringify({ matchId: active.match.id, content }));
    setDraft("");
    setChatError("");
  }

  function signOut() {
    localStorage.removeItem("sideby.studentId");
    router.push("/login");
  }

  const peer = active?.student;
  const sharedLikes = profile && peer ? profile.likes.filter((like) => peer.likes.some((peerLike) => peerLike.toLowerCase() === like.toLowerCase())) : [];
  const currentMatching = !searching && (result?.status === "matched" || Boolean(active));

  return (
    <main className="shell min-h-screen p-3 sm:p-5 lg:p-7">
      <div className="mx-auto grid min-h-[calc(100vh-3.5rem)] max-w-[1600px] grid-cols-1 gap-3 lg:grid-cols-[220px_minmax(360px,1fr)_minmax(320px,370px)]">
        <aside className="glass flex flex-col px-4 py-5 lg:min-h-[calc(100vh-3.5rem)]">
          <Link href="/dashboard" className="mb-9 flex items-center gap-3 px-2">
            <span className="grid h-9 w-9 place-items-center border border-white/10 bg-white/[.04] text-[var(--acid)]"><Sparkles size={16} /></span>
            <span className="font-display text-xl">sideby<span className="text-[var(--acid)]">.</span></span>
          </Link>
          <div className="mb-3 px-3 eyebrow">Your space</div>
          <nav className="grid gap-1">
            <a href="#discover" className="flex items-center gap-3 border-l-2 border-[var(--acid)] bg-white/[.055] px-3 py-3 text-sm text-white"><Compass size={16} className="text-[var(--acid)]" /> Discover <span className="ml-auto text-[10px] text-[#82907f]">01</span></a>
            <button onClick={() => document.getElementById("conversation-list")?.scrollIntoView({ behavior: "smooth" })} className="flex items-center gap-3 border-l-2 border-transparent px-3 py-3 text-left text-sm text-[#aab2a8] transition hover:bg-white/[.035] hover:text-white"><MessageCircle size={16} /> Conversations <span className="ml-auto text-[10px] text-[#82907f]">{history.length.toString().padStart(2, "0")}</span></button>
          </nav>
          <div className="mb-3 mt-9 px-3 eyebrow">Your connections</div>
          <div id="conversation-list" className="scrollbar-thin flex-1 overflow-y-auto">
            {history.length ? history.map((item) => <button key={item.match.id} onClick={() => { setActive(item); setMobileChatOpen(true); }} className={`flex w-full items-center gap-3 border-l-2 px-3 py-3 text-left transition ${active?.match.id === item.match.id ? "border-[var(--acid)] bg-white/[.05]" : "border-transparent hover:bg-white/[.035]"}`}>
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full border border-white/10 bg-[#334334] text-xs font-semibold text-[var(--acid)]">{initials(item.student?.name ?? "Student")}</span>
              <span className="min-w-0 flex-1"><span className="block truncate text-sm">{item.student?.name ?? "Student"}</span><span className="mt-1 block truncate text-[11px] text-[#8d978d]">{item.student?.department ?? "New connection"}</span></span>
              {item.match.status === "active" && <span className="status-dot" aria-label="Active match" />}
            </button>) : <p className="px-3 py-3 text-xs leading-5 text-[#7f897f]">Your conversations will show up here.</p>}
          </div>
          <div className="mt-5 border-t border-white/10 pt-4">
            <button className="flex w-full items-center gap-3 px-3 py-3 text-left text-sm text-[#aab2a8] hover:text-white"><CircleHelp size={16} /> Help & support</button>
            <div className="mt-3 flex items-center gap-3 border-t border-white/10 px-2 pt-4">
              <span className="grid h-9 w-9 place-items-center rounded-full border border-white/10 bg-[#454337] text-xs font-semibold text-[#f1ddae]">{initials(profile?.name ?? "You")}</span>
              <span className="min-w-0 flex-1"><span className="block truncate text-sm">{profile?.name ?? "Your profile"}</span><span className="block truncate text-[11px] text-[#818b81]">{profile?.department ?? "Student"}</span></span>
              <button title="Sign out" aria-label="Sign out" className="icon-button h-8 w-8 border-0 bg-transparent" onClick={signOut}><LogOut size={15} /></button>
            </div>
          </div>
        </aside>

        <section id="discover" className="flex min-w-0 flex-col gap-3">
          <header className="glass flex min-h-[68px] items-center justify-between gap-4 px-5 sm:px-7">
            <div><p className="eyebrow">Tuesday / campus connections</p><p className="mt-1 text-sm text-[#d7ded4]">Your next good conversation starts here.</p></div>
            <div className="flex items-center gap-3"><span className="hidden items-center gap-2 text-xs text-[#aab2a8] sm:flex"><span className={`status-dot ${socketState !== "online" ? "offline" : ""}`} />{socketState === "online" ? "Chat connected" : socketState === "connecting" ? "Connecting" : "Reconnecting"}</span><button className="icon-button" title="Settings" aria-label="Settings"><Settings2 size={17} /></button></div>
          </header>

          <section className="glass relative flex min-h-[410px] flex-1 flex-col overflow-hidden p-6 sm:p-8 lg:min-h-[520px]">
            <div className="pointer-events-none absolute inset-0 opacity-[.55]" style={{ backgroundImage: "linear-gradient(115deg, rgba(212,247,106,.055), transparent 42%), url('https://images.unsplash.com/photo-1529156069898-49953e39b3ac?auto=format&fit=crop&w=1400&q=80')", backgroundSize: "cover", backgroundPosition: "center", maskImage: "linear-gradient(to right, rgba(0,0,0,.68), transparent 92%)" }} />
            <div className="relative flex items-center justify-between"><div className="eyebrow flex items-center gap-2"><span className="h-px w-5 bg-[var(--acid)]" /> Discover</div><button className="ghost-button h-9 min-h-9 text-xs" onClick={() => void refreshHistory().catch(() => undefined)}><RefreshCw size={13} /> Refresh</button></div>
            <div className="relative grid flex-1 items-center gap-5 py-8 xl:grid-cols-[minmax(0,1fr)_260px]">
              <div className="max-w-xl">
                <AnimatePresence mode="wait">
                  {currentMatching && peer ? <motion.div key={active?.match.id ?? "found"} initial={{ opacity: 0, x: -24 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: 20 }} transition={{ duration: .45 }}>
                    <p className="eyebrow mb-4 flex items-center gap-2"><span className="status-dot" /> A new connection</p>
                    <h1 className="display text-5xl leading-[1.03] sm:text-6xl">You found<br /><i className="text-[var(--acid)]">your sideby.</i></h1>
                    <div className="mt-7 flex items-center gap-4 border-l border-[var(--acid)]/50 pl-4">
                      <span className="grid h-14 w-14 shrink-0 place-items-center rounded-full border border-[var(--acid)]/40 bg-[#384334] text-lg text-[var(--acid)]">{initials(peer.name)}</span>
                      <div className="min-w-0"><p className="truncate text-lg font-medium">{peer.name}, {peer.age}</p><p className="mt-1 truncate text-sm text-[#aab2a8]">{peer.department} · {peer.class}</p></div>
                    </div>
                    <div className="mt-6 flex flex-wrap gap-2">
                      {sharedLikes.slice(0, 4).map((like) => <span key={like} className="border border-[var(--acid)]/25 bg-[var(--acid)]/[.07] px-2.5 py-1.5 text-xs text-[#e3f5b5]">Both like {like}</span>)}
                      {profile?.department === peer.department && <span className="border border-white/10 bg-white/[.04] px-2.5 py-1.5 text-xs text-[#ccd4ca]">Same department</span>}
                      {profile?.class === peer.class && <span className="border border-white/10 bg-white/[.04] px-2.5 py-1.5 text-xs text-[#ccd4ca]">Same class</span>}
                      {!sharedLikes.length && profile?.department !== peer.department && <span className="border border-white/10 bg-white/[.04] px-2.5 py-1.5 text-xs text-[#ccd4ca]">A fresh perspective</span>}
                    </div>
                    <button className="primary-button mt-7" onClick={() => setMobileChatOpen(true)}>Say hello <ArrowRight size={15} /></button>
                  </motion.div> : <motion.div key="search" initial={{ opacity: 0, x: 18 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -18 }} transition={{ duration: .35 }}>
                    <p className="eyebrow mb-4 flex items-center gap-2"><span className={`status-dot ${searching ? "breathe" : ""}`} /> {searching ? "In the mix" : "A better kind of introduction"}</p>
                    <h1 className="display max-w-lg text-5xl leading-[1.03] sm:text-6xl">{searching ? <>Good things<br /><i className="text-[var(--acid)]">take a moment.</i></> : <>Find a little<br /><i className="text-[var(--acid)]">common ground.</i></>}</h1>
                    <p className="mt-5 max-w-md text-sm leading-6 text-[#afb7ad]">{searching ? "You’re in the matching pool. We’ll let you know the moment someone clicks." : "Meet a student who gets your interests, your campus, and maybe your next coffee order."}</p>
                    {searching && <div className="mt-6 flex items-center gap-2 text-xs text-[#b6c28e]"><LoaderCircle size={15} className="animate-spin" /> Searching profiles <span className="ml-1 text-[#737e70]">·</span><span className="text-[#899384]">You can keep chatting</span></div>}
                    <button onClick={() => void handleFindMatch()} disabled={searching} className="primary-button mt-7"><Sparkles size={16} /> {searching ? "Searching…" : result?.status === "matched" ? "Find another match" : "Find my match"} <ArrowUpRight size={15} /></button>
                  </motion.div>}
                </AnimatePresence>
                {error && <p role="alert" className="relative mt-5 max-w-md border-l-2 border-[var(--clay)] bg-black/20 px-3 py-2 text-sm text-[#f0ad93]">{error}</p>}
              </div>
              <div className="relative mx-auto hidden h-[250px] w-[250px] place-items-center xl:grid">
                <div className="absolute inset-0 rounded-full border border-white/10" />
                <div className="absolute inset-[22px] rounded-full border border-white/[.08]" />
                <div className="absolute inset-[48px] rounded-full border border-dashed border-[var(--acid)]/25" />
                <div className="sweep absolute inset-[10px] rounded-full border-t border-[var(--acid)]/80" />
                <div className="breathe grid h-28 w-28 place-items-center rounded-full border border-[var(--acid)]/25 bg-[rgba(212,247,106,.06)] shadow-[0_0_70px_rgba(212,247,106,.12)]"><Compass size={34} strokeWidth={1.1} className="text-[var(--acid)]" /></div>
                <span className="absolute left-2 top-12 h-2 w-2 rounded-full bg-[var(--clay)]" /><span className="absolute bottom-7 right-9 h-1.5 w-1.5 rounded-full bg-[var(--acid)]" />
                <span className="absolute bottom-1 text-[9px] uppercase tracking-[.18em] text-white/35">Finding common ground</span>
              </div>
            </div>
            <div className="relative flex items-center justify-between border-t border-white/10 pt-4 text-[11px] text-[#899288]"><span className="flex items-center gap-2"><Users size={14} /> Thoughtful connections, one at a time</span><span className="hidden sm:inline">Based on mutual preferences · shared interests</span></div>
          </section>

          <div className="grid gap-3 sm:grid-cols-3">
            <div className="glass-soft flex items-center gap-3 p-4"><span className="grid h-9 w-9 place-items-center border border-[var(--acid)]/20 bg-[var(--acid)]/[.05] text-[var(--acid)]"><Heart size={16} /></span><span><span className="eyebrow block">Shared interests</span><span className="mt-1 block text-xs text-[#b5bdb3]">Little things, in common</span></span></div>
            <div className="glass-soft flex items-center gap-3 p-4"><span className="grid h-9 w-9 place-items-center border border-white/10 bg-white/[.04] text-[#b3beb2]"><MessageCircle size={16} /></span><span><span className="eyebrow block">Real conversations</span><span className="mt-1 block text-xs text-[#b5bdb3]">Right here, in real time</span></span></div>
            <div className="glass-soft flex items-center gap-3 p-4"><span className="grid h-9 w-9 place-items-center border border-white/10 bg-white/[.04] text-[#b3beb2]"><Check size={16} /></span><span><span className="eyebrow block">Your pace</span><span className="mt-1 block text-xs text-[#b5bdb3]">Come and go as you like</span></span></div>
          </div>
        </section>

        <aside className={`glass flex min-h-[560px] flex-col overflow-hidden lg:min-h-0 ${mobileChatOpen ? "fixed inset-3 z-40" : "hidden lg:flex"}`}>
          <div className="flex min-h-[68px] items-center justify-between border-b border-white/10 px-5">
            <div><p className="eyebrow">Your conversations</p><p className="mt-1 text-sm">{peer?.name ?? "A place to say hello"}</p></div>
            <div className="flex items-center gap-2"><button title="More options" aria-label="More options" className="icon-button border-0 bg-transparent"><MoreHorizontal size={19} /></button><button className="icon-button lg:hidden" aria-label="Close chat" onClick={() => setMobileChatOpen(false)}><X size={16} /></button></div>
          </div>
          {active ? <>
            <div className="flex items-center gap-3 border-b border-white/[.07] px-5 py-3">
              <span className="grid h-9 w-9 place-items-center rounded-full border border-white/10 bg-[#334334] text-xs font-semibold text-[var(--acid)]">{initials(peer?.name ?? "Student")}</span>
              <div className="min-w-0 flex-1"><p className="truncate text-sm">{peer?.name ?? "Your match"}</p><p className="truncate text-[11px] text-[#879187]">{peer?.department ?? "Sideby connection"}</p></div>
              <span className="flex items-center gap-1.5 text-[10px] text-[#879187]"><span className={`status-dot ${socketState === "online" ? "" : "offline"}`} />{socketState === "online" ? "Online" : "Offline"}</span>
            </div>
            <div className="scrollbar-thin flex flex-1 flex-col gap-3 overflow-y-auto px-4 py-4">
              <div className="my-2 text-center text-[10px] text-[#707a70]">A new conversation, made together.</div>
              {messages.map((message) => {
                const mine = message.senderId === studentId;
                return <motion.div initial={{ opacity: 0, y: 7 }} animate={{ opacity: 1, y: 0 }} key={message.id} className={`flex ${mine ? "justify-end" : "justify-start"}`}>
                  <div className={`max-w-[88%] px-3 py-2.5 ${mine ? "border border-[var(--acid)]/20 bg-[rgba(212,247,106,.1)] text-[#e5eed6]" : "border border-white/[.08] bg-white/[.045] text-[#d5dcd2]"}`}>
                    <p className="break-words text-[13px] leading-5">{message.content}</p><p className="mt-1.5 text-right text-[9px] text-[#879187]">{timeLabel(message.createdAt)}</p>
                  </div>
                </motion.div>;
              })}
              {!messages.length && <div className="m-auto max-w-[220px] text-center"><span className="mx-auto grid h-12 w-12 place-items-center rounded-full border border-white/10 bg-white/[.035] text-[var(--acid)]"><MessageCircle size={18} /></span><p className="mt-4 text-sm text-[#d6ddd3]">Break the ice.</p><p className="mt-2 text-xs leading-5 text-[#858f84]">You both like {sharedLikes[0] ?? "finding common ground"}. There’s your opening.</p></div>}
              <div ref={endOfMessagesRef} />
            </div>
            {chatError && <p role="status" className="mx-4 mb-2 text-xs text-[#e7aa91]">{chatError}</p>}
            <form onSubmit={sendMessage} className="flex items-center gap-2 border-t border-white/10 p-3">
              <input className="input min-h-[42px] flex-1" value={draft} onChange={(event) => setDraft(event.target.value)} placeholder="Write a message…" aria-label="Message" maxLength={4000} />
              <button className="primary-button h-[42px] min-h-[42px] w-[42px] !px-0" type="submit" disabled={!draft.trim() || socketState !== "online"} aria-label="Send message"><Send size={16} /></button>
            </form>
          </> : <div className="flex flex-1 flex-col">
            <div className="relative flex flex-1 flex-col items-center justify-center px-8 text-center">
              <div className="absolute inset-x-6 top-8 h-32 opacity-25" style={{ backgroundImage: "url('https://images.unsplash.com/photo-1511632765486-a01980e01a18?auto=format&fit=crop&w=700&q=75')", backgroundSize: "cover", backgroundPosition: "center", maskImage: "linear-gradient(to bottom,black,transparent)" }} />
              <span className="relative grid h-14 w-14 place-items-center border border-[var(--acid)]/20 bg-[rgba(212,247,106,.06)] text-[var(--acid)]"><MessageCircle size={21} /></span>
              <h2 className="display mt-5 text-2xl">Your next hello<br />is waiting.</h2>
              <p className="mt-3 max-w-[230px] text-xs leading-5 text-[#8e988e]">Find a match to open a real-time conversation. Past chats live in your sidebar.</p>
              <button onClick={() => void handleFindMatch()} disabled={searching} className="ghost-button mt-5"><Sparkles size={14} /> {searching ? "Searching…" : "Find someone"}</button>
            </div>
            <div className="border-t border-white/10 p-4 text-[10px] leading-5 text-[#798379]">Messages are saved to your match history. Be kind; this is a real person on the other side.</div>
          </div>}
        </aside>
      </div>
    </main>
  );
}
