# Sideby Web

Next.js App Router frontend for the Fiber student matching and chat API.

## Run locally

1. Start the Go API on `http://localhost:3000` with MongoDB available.
2. From this folder run `npm install` and `npm run dev`.
3. Open `http://localhost:3001` and create a student profile.

The API defaults to `http://localhost:3000` and the WebSocket defaults to `ws://localhost:3000`. Override these with `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` when needed. The Go API should allow this app's origin via `FRONTEND_ORIGINS` (defaults include ports 3001 on localhost and 127.0.0.1).

The browser stores only the student's ID in `localStorage`. Authentication is not currently enabled by the backend, so this is a development flow and must not be treated as secure identity verification.
