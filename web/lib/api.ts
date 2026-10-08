import type { ChatMessage, MatchHistoryItem, MatchResult, Student } from "@/lib/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3000";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...init?.headers },
      cache: "no-store",
    });
  } catch {
    throw new Error("Could not reach the matching server. Check that the Go API is running.");
  }

  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(typeof body.error === "string" ? body.error : `Request failed (${response.status})`);
  }
  return body as T;
}

export interface RegistrationInput extends Omit<Student, "id" | "createdAt" | "email"> {
  email: string;
  password: string;
}

export function registerStudent(input: RegistrationInput): Promise<Student> {
  return request<Student>("/api/students/register", { method: "POST", body: JSON.stringify(input) });
}

export function loginStudent(email: string, password: string): Promise<Student> {
  return request<Student>("/api/students/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export function getStudent(studentId: string): Promise<Student> {
  return request<Student>(`/api/students/${encodeURIComponent(studentId)}`);
}

export function findMatch(studentId: string): Promise<MatchResult> {
  return request<MatchResult>("/api/match/find", {
    method: "POST",
    body: JSON.stringify({ studentId }),
  });
}

export async function getMatchHistory(studentId: string): Promise<MatchHistoryItem[]> {
  const result = await request<{ matches: MatchHistoryItem[] }>(`/api/match/history/${encodeURIComponent(studentId)}`);
  return result.matches;
}

export async function getMessages(matchId: string, studentId: string): Promise<ChatMessage[]> {
  const result = await request<{ messages: ChatMessage[] }>(
    `/api/match/${encodeURIComponent(matchId)}/messages?student_id=${encodeURIComponent(studentId)}`,
  );
  return result.messages;
}

export function chatSocketUrl(studentId: string): string {
  const base = process.env.NEXT_PUBLIC_WS_URL ?? API_BASE.replace(/^http/, "ws");
  return `${base.replace(/\/$/, "")}/ws/chat?student_id=${encodeURIComponent(studentId)}`;
}
