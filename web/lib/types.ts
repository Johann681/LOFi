export interface StudentPreferences {
  targetGender: string;
  orientation: string;
}

export interface Student {
  id: string;
  email: string;
  name: string;
  age: number;
  height: string;
  department: string;
  class: string;
  gender: string;
  orientation: string;
  verificationProof: string;
  preferences: StudentPreferences;
  likes: string[];
  dislikes: string[];
  createdAt: string;
}

export interface Match {
  id: string;
  student1Id: string;
  student2Id: string;
  status: "active" | "unmatched";
  createdAt: string;
}

export interface MatchResult {
  status: "searching" | "matched";
  matchId?: string;
  student?: Student;
}

export interface MatchHistoryItem {
  match: Match;
  student?: Student;
}

export interface ChatMessage {
  id: string;
  matchId: string;
  senderId: string;
  recipientId: string;
  content: string;
  createdAt: string;
}

export type WebSocketEvent =
  | { type: "MATCH_FOUND"; data: { matchId: string; status: string } }
  | { type: "CHAT_MESSAGE"; message: ChatMessage }
  | { type: "ERROR"; error: string };
